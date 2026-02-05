# 🟧 **Kubernetes Status (kubectl)**
```bash
colima start --kubernetes --cpu 6 --memory 8
```
Everything below uses `kubectl` since Colima runs a local K8s distribution.

---

## 🧩 **Cluster & Node Status**

### View cluster info

```sh
kubectl cluster-info
```

### View nodes

```sh
kubectl get nodes
kubectl describe node <node-name>
```

---

## 📦 **Deployment / Pod Status**

### View all pods (all namespaces)

```sh
kubectl get pods -A
```

### View services

```sh
kubectl get svc -A
```

### View deployments

```sh
kubectl get deployments -A
```
### View PersistentVolumeClaim (PVC)

```sh
kubectl get pvc -n <namespace>
```

---

## 🔁 **Restarting K8s Deployments After Code Changes**

### **Force Kubernetes to pull new image**

If using `imagePullPolicy: Always`, just redeploy:

```sh
kubectl rollout restart deployment users -n users
```

If using local images, sometimes you need:

```sh
kubectl delete pod -l app=<pod-name> -n <namespace>
```

```sh
kubeclt delete job <job-name> -n <namespace>
```

---

## 🧹 **Clear CrashLoop Pods**

```sh
kubectl delete pod -n users --all
```

---

## 🔍 **Debug / Inspect**

### Describe any object

```sh
kubectl describe pod <name> -n users
kubectl describe deployment <name> -n users
kubectl describe svc <name> -n users
```

### Logs


```sh
kubectl logs -n users -l job-name=users-migrate  
kubectl logs -n users -l app=users-db  
```

Manual logs:

```sh
kubectl logs <podname> -n <namespace> -f
```

Get secrets from pod:
```sh
kubectl get secret <secretname> -n <namespace> -o yaml
```

---

## 🛠️ **Apply YAML Changes**

Any YAML changes:

```sh
kubectl apply -f <file>.yaml
```

Apply the whole folder:

```sh
kubectl apply -f k8s/
```

---

## 💣 **Delete + Reapply Deployment When Things Break**

```sh
kubectl delete -f k8s/users/deployment.yaml
kubectl apply -f k8s/users/deployment.yaml
```

---

# 🔄 **Common Update Workflow Cheat Sheet**

## 🧑‍💻 When changing Go code:

1. Rebuild image

   ```sh
   docker build -t social-network/users:dev -f services/users/Dockerfile .
   ```
2. Restart deployment

   ```sh
   kubectl rollout restart deployment users -n users
   ```

## 📄 When changing YAML:

```sh
kubectl apply -f k8s/users/deployment.yaml
kubectl apply -f k8s/...   # etc.
```

## 🏚️ When cluster gets stuck:

```sh
colima stop && colima start --kubernetes
```

## 🔥 When everything is completely broken:

```sh
colima delete
colima start --kubernetes
make all
```

---

If you'd like, I can also generate a **side-by-side workflow chart**, or a **Makefile target** like `make restart-cluster` to simplify this.
