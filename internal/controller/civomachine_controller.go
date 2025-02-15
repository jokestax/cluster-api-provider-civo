/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/civo/civogo"
	infra "github.com/jokestax/cluster-api-provider-civo/api/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CivoMachineReconciler reconciles a CivoMachine object
type CivoMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=my.domain,resources=civomachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=my.domain,resources=civomachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=my.domain,resources=civomachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CivoMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *CivoMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	civoMachine := infra.CivoMachine{}
	err := r.Get(ctx, req.NamespacedName, &civoMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		l.Error(err, "Failed to fetch CivoCluster")
		return ctrl.Result{}, err
	}

	apiKey := os.Getenv("CIVO_API_KEY")

	client, err := civogo.NewClient(apiKey, civoMachine.Spec.Region)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create Civo client: %w", err)
	}

	instanceInfo, err := client.FindInstance(civoMachine.Spec.Name)
	if err != nil {
		if !errors.Is(err, civogo.ZeroMatchesError) {
			return ctrl.Result{}, err
		}
		log.FromContext(ctx).Info("Machine not found in Civo, proceeding with creation")
	}

	if civoMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(civoMachine.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer") {
			civoMachine.ObjectMeta.Finalizers = append(civoMachine.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer")
			if err := r.Client.Update(ctx, &civoMachine); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(civoMachine.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer") {
			_, err := client.DeleteInstance(instanceInfo.ID)
			if err != nil {
				return ctrl.Result{}, err
			}

			civoMachine.ObjectMeta.Finalizers = removeString(civoMachine.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer")
			if err := r.Update(ctx, &civoMachine); err != nil {
				l.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}

		}
		l.Info("Machine Deleted", "machine name", civoMachine.Spec.Name)
		return ctrl.Result{}, nil
	}

	if civoMachine.Status.Ready {
		l.Info("machine already provisioned", "machine_name", civoMachine.Spec.Name)
		return ctrl.Result{}, nil
	}

	instanceConfig := civogo.InstanceConfig{
		Hostname: civoMachine.Spec.Name,
		Count:    civoMachine.Spec.Count,
		Region:   civoMachine.Spec.Region,
		Size:     civoMachine.Spec.InstanceType,
	}

	instance, err := client.CreateInstance(&instanceConfig)
	if err != nil {
		if errors.Is(err, civogo.DatabaseInstanceDuplicateNameError) {
			l.Info("cluster with same name already present", "host name", civoMachine.Spec.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	l.Info("Provisioned Instance", "instance name", instance.Hostname)

	civoMachine.Status.Ready = true
	err = r.Status().Update(ctx, &civoMachine)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CivoMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infra.CivoMachine{}).
		Named("civomachine").
		Complete(r)
}
