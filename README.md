# CAPI-Civo  

A simple Cluster API (CAPI) infrastructure provider for provisioning Kubernetes clusters on [Civo](https://www.civo.com).  

## 🚀 Getting Started  

### **Prerequisites**  
- A Kubernetes cluster running locally (kind, k3d, or minikube).  
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/) installed and configured.  
- A valid Civo API key.  

### **Setup**  

1. **Export your Civo API Key:**  
   ```sh
   export CIVO_API_KEY="your-api-key"

2. **Ensure your kubeconfig is set to a local cluster**

```sh
kubectl config current-context
```

3. Install CRDs
Apply the Custom Resource Definitions (CRDs) required by the controller:

```
make install
```

4. Run the controller locally
Start the Civo Cluster API controller:

```
make run
```

5. Apply a sample CivoCluster resource
Create a Kubernetes cluster in Civo by applying a sample CR:

```
kubectl apply -f config/samples/CivoCluster.yaml
```
6. Delete the sample CivoCluster resource

```
kubectl delete -f config/samples/CivoCluster.yaml
```
7. Apply a sample CivoMachine resource
Create a Kubernetes cluster in Civo by applying a sample CR:

```
kubectl apply -f config/samples/CivoMachine.yaml
```
8. Delete the sample CivoMachine resource

```
kubectl delete -f config/samples/CivoMachine.yaml
```

Using clusterctl 

1. Create a clusterctl.yaml file with contents as below 

```
providers:
  - name: "civo"
    url: "https://github.com/jokestax/cluster-api-provider-civo/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

2. Run clusterctl init cmd 

```
clusterctl init --infrastructure civo --config=<PATH_TO_ABOVE_CLUSTERCTL_YAML>
```
