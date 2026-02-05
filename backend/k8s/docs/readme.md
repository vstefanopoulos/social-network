# K8s

## Dependencies

| Dependency                        | Purpose                          | Instal |
| --------------------------------- | -------------------------------- |--------
| **Docker**                        | Builds images                    | https://docs.docker.com/get-docker/ |
| **Minicube**                      | Cluster container                | https://minikube.sigs.k8s.io/docs/start/ |
| **kubectl**                       | Applies manifests, logs, deletes | https://kubernetes.io/docs/tasks/tools/ |
| **Helm**                          | Installs ingress-nginx           | https://helm.sh/docs/intro/install/ |



## USAGE
### Preliminary Steps
#### Build base (only on first run)
```bash
make build-base
```

#### MacOS
```bash
colima start --kubernetes --memory 4 --cpu 4
```

#### Windows 
```bash
minikube start --driver=docker --cpus=4 --memory=4096
```

#### confirm cluster
```bash
kubectl get nodes
```
### Build and deploy cluster
#### 1. Build local users image
```bash
make build-services
```

#### 2. Create namespace + Postgres DB + users service
```bash
make all
```

#### 3. Watch pods
```bash
kubectl get pods -n social-network
```

#### 4. Logs
```bash
kubectl logs -l app=<app-name> -n social-network -f
```