# 🚀 Kubernetes Scheduled Scaler Operator

**Kubernetes Scheduled Scaler Operator** provides an automated solution for scaling deployments based on predefined time windows. This operator allows users to schedule scaling events, ensuring that deployments dynamically adjust their replica count during specific timeframes. 

By intelligently managing replica counts, this operator helps **optimize resource utilization, reduce costs, and enhance workload performance** by scaling applications up when needed and down during off-peak hours.

---

## 📌 Features
✅ **Time-Based Scaling** – Define time windows for scaling up/down  
✅ **Automatic Resource Optimization** – Improve efficiency and reduce costs  
✅ **Multi-Deployment Support** – Scale multiple deployments in different namespaces  
✅ **Timezone-Aware** – Configure scaling schedules based on different time zones  

---

### How to use:
**Apply the operator to your cluster:**
```
kubectl apply -f https://raw.githubusercontent.com/omerrevach/k8s-scheduled-scaler-operator/main/install.yaml
```

**Wait about 30 seconds and run this command to check if its up**
```
kubectl get pod -l control-plane=controller-manager -n k8s-operator-system

# should see something like this - k8s-operator-controller-manager-5547d68d59-czlkw
```
**Create scheduled-scaler.yaml to store the configuration**
```
apiVersion: api.omerrevach.online/v1alpha1
kind: Scaler
metadata:
  name: scaler-sample
spec:
  start: "16:30"                  # Start scaling at 16:30 AM
  end: "18:00"                    # Stop scaling at 18:00 PM
  replicas: 5                      # Scale up to 5 pods when in the schedule
  normalReplicasAmount: 2          # Scale down to 2 pods outside the schedule
  timezone: "Asia/Jerusalem"       # Use the specified timezone
  deployments:                     
    - name: app1
      namespace: default
    - name: app2                    # Add here in the deployments section all the deployments you want to scale
      namespace: default
    - name: app3
      namespace: default
```
>**Warning**  - This is case sensitive and make sure the time is as shown in the example and the timezone is correctly spelled











## build and publish image for local deployment:
```
make docker-build IMG=rebachi/scheduled-scaler-op:v1
make docker-push IMG=rebachi/scheduled-scaler-op:v1
make deploy IMG=rebachi/scheduled-scaler-op:v1
kubectl get pods -n k8s-operator-system

cat config/crd/bases/api.omerrevach.online_scalers.yaml \
    config/rbac/role.yaml \
    config/rbac/role_binding.yaml \
    config/manager/manager.yaml > install.yaml
```

## To start the Operator locally:
```
kubectl apply -f config/crd/bases/api.omerrevach.online_scalers.yaml
make run

kubectl apply -f config/crd/bases/api.omerrevach.online_scalers.yaml

kubectl get crd

kubectl delete -f config/crd/bases/api.omerrevach.online_scalers.yaml
kubectl delete -f config/samples/api_v1alpha1_scaler.yaml
```

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/k8s-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/k8s-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/k8s-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/k8s-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

