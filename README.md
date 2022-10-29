# kupdater
Kupdater - Kubernetes Updater is a minimal operator to track versions of your applications already deployed to kubernetes cluster.

## Description
In a Gitops world we usually end up having a mix of helm charts, our own yaml manifests, external dependencies packed in one way or another inside Git repositories. 

It can easily end-up as huge pain to keep up with all the changes we depend on. 
Once configured Kupdater watches that for you and lets you know if everything has changes.

```
tomasl@Tomass-Air ~ % kubectl get update -A
NAMESPACE        NAME                  TYPE     VERSION            STATUS                           SYNCED
argocd           argocd                github   v2.4.2             Outdated (v2.2.15 available)     12h
cnpg             cnpg                  helm     0.15.1             UpToDate                         12h
descheduler      descheduler           helm     0.25.2             UpToDate                         12h
dex              dex                   helm     0.12.1             UpToDate                         12h
forecastle       forecastle            github   v1.0.103           UpToDate                         12h
jackett          jackett               github   v0.20.2158-ls78    UpToDate                         12h
kyverno          kyverno               helm     v2.1.10            Outdated (2.6.1-rc1 available)   12h
loki             loki                  helm     2.8.3              UpToDate                         12h
minio-operator   minio-operator        helm     4.3.7              UpToDate                         12h
pihole           pihole                github   2022.10            UpToDate                         12h
prometheus       prometheus-operator   helm     34.*               Outdated (41.5.1 available)      12h
sonarr           sonarr                github   3.0.9.1549-ls160   UpToDate                         12h
traefik          traefik               helm     17.0.5             UpToDate                         12h
transmission     transmission          github   3.00-r5-ls137      UpToDate                         12h
velero           velero                helm     *                  Outdated (2.32.1 available)      12h
```

## Features
- Periodically watches configured applications for updates
- Configurable via:
  - CRDs [Done]
  - ArgoCD application discovery [Experimental]
  - Annotations on existing resources [TODO]

## Installation
### Helm chart
```
helm repo add getais https://getais.github.io/helm-charts
helm repo update
helm install kupdater getais/kupdater
```

### From source
Building docker image:
```
make docker-build docker-push IMG="somerepo/kupdater:v0.0.1"
```
Deploying manifests:
```bash
make deploy IMG="somerepo/kupdater:v0.0.1"
```

## Configuration
Example helm source:
```yaml
apiVersion: ops.getais.cloud/v1alpha1
kind: AppVersion
metadata:
  name: traefik
  namespace: traefik
spec:
  name: traefik
  source: https://helm.traefik.io/traefik
  type: helm
  version: "17.0.5"
```

Example Github source:
```yaml
apiVersion: ops.getais.cloud/v1alpha1
kind: AppVersion
metadata:
  name: pihole
  namespace: pihole
spec:
  name: pihole
  version: "2022.10"
  source: https://github.com/pi-hole/docker-pi-hole
  type: github
```

## Contributing
PRs are welcome. 
Github issues for feature-requests / bugs / ideas

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

## Development guide

### Running locally against kubernetes cluster
1. Install CRD definitions
```bash
make generate
make manifests
make install
```
2. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

3. Run operator on local machine against current k8s context
```bash
make run
```

### Cleaning up
To delete the CRDs from the cluster:

```sh
make uninstall
```

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

