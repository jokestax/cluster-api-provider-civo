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

type CivoClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=civoclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=civoclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=civoclusters/finalizers,verbs=update

func (r *CivoClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	civoCluster := infra.CivoCluster{}
	if err := r.Get(ctx, req.NamespacedName, &civoCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		l.Error(err, "Failed to fetch CivoCluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	apiKey := os.Getenv("CIVO_API_KEY")

	client, err := civogo.NewClient(apiKey, civoCluster.Spec.Region)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create Civo client: %w", err)
	}

	clusterConfig := civogo.KubernetesClusterConfig{
		Name:            civoCluster.Name,
		NumTargetNodes:  civoCluster.Spec.NumWorkerNodes,
		TargetNodesSize: civoCluster.Spec.DefaultMachineType,
	}

	clusterInfo, err := client.FindKubernetesCluster(civoCluster.Name)
	if err != nil {
		if err.Error() == fmt.Sprintf("ZeroMatchesError: unable to find %s, zero matches", civoCluster.Name) {
			log.FromContext(ctx).Info("Cluster not found in Civo, proceeding with creation")
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get cluster info: %w", err)
		}
	}

	if civoCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(civoCluster.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer") {
			civoCluster.ObjectMeta.Finalizers = append(civoCluster.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer")
			if err := r.Update(ctx, &civoCluster); err != nil {
				l.Error(err, "Failed to add finalizer to CivoCluster")
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(civoCluster.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer") {
			_, err := client.DeleteKubernetesCluster(clusterInfo.ID)
			if err != nil {
				l.Error(err, "Failed to delete cluster from Civo")
				return ctrl.Result{}, err
			}

			civoCluster.ObjectMeta.Finalizers = removeString(civoCluster.ObjectMeta.Finalizers, "infrastructure.cluster.x-k8s.io/finalizer")
			if err := r.Update(ctx, &civoCluster); err != nil {
				l.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		l.Info("Cluster deleted", "name", civoCluster.Name)
		return ctrl.Result{}, nil
	}
	// Check if cluster is already ready
	if civoCluster.Status.Ready {
		l.Info("Cluster already provisioned", "name", civoCluster.Name)
		return ctrl.Result{}, nil
	}

	cluster, err := client.NewKubernetesClusters(&clusterConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create cluster in Civo: %w", err)
	}

	l.Info("Cluster created in Civo", "id", cluster.ID, "name", cluster.Name)

	civoCluster.Status.Ready = true
	civoCluster.Status.ControlPlaneEndpoint = cluster.APIEndPoint
	civoCluster.Status.NodeCount = civoCluster.Spec.NumWorkerNodes
	// TODO(user): your logic here

	if err := r.Status().Update(ctx, &civoCluster); err != nil {
		l.Error(err, "Failed to update cluster status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CivoClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infra.CivoCluster{}).
		Named("civocluster").
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, a := range slice {
		if a == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	for i, a := range slice {
		if a == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
