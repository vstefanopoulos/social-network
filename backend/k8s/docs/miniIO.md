Perfect! Let’s break this into two parts: **how to connect your Kubernetes `users` deployment to MinIO**, and **how your Go code interacts with it**.

---

## **1. Kubernetes Setup**

We need MinIO running inside your cluster (or locally), expose it, and provide credentials to your `users` deployment.

### **Step 1: Run MinIO in Kubernetes**

Here’s a simple Deployment + Service for MinIO:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: storage
---
apiVersion: v1
kind: Secret
metadata:
  name: minio-secret
  namespace: storage
type: Opaque
data:
  MINIO_ACCESS_KEY: bWluaW8=       # base64 of "minio"
  MINIO_SECRET_KEY: bWluaW8xMjM=   # base64 of "minio123"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: storage
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
        - name: minio
          image: minio/minio
          args:
            - server
            - /data
          env:
            - name: MINIO_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-secret
                  key: MINIO_ACCESS_KEY
            - name: MINIO_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-secret
                  key: MINIO_SECRET_KEY
          ports:
            - containerPort: 9000
          volumeMounts:
            - name: data
              mountPath: /data
      volumes:
        - name: data
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: storage
spec:
  type: ClusterIP
  selector:
    app: minio
  ports:
    - port: 9000
      targetPort: 9000
```

✅ This runs MinIO inside the `storage` namespace and exposes it on `minio.storage:9000` inside the cluster.

---

### **Step 2: Connect your `users` deployment to MinIO**

Add environment variables in your `users` deployment:

```yaml
        env:
          - name: MINIO_ENDPOINT
            value: "minio.storage:9000"
          - name: MINIO_ACCESS_KEY
            valueFrom:
              secretKeyRef:
                name: minio-secret
                key: MINIO_ACCESS_KEY
          - name: MINIO_SECRET_KEY
            valueFrom:
              secretKeyRef:
                name: minio-secret
                key: MINIO_SECRET_KEY
```

Now, your Go app can read these env vars and connect to MinIO over its S3-compatible API.

---

## **2. Go Code Interaction with MinIO**

You can use the official [MinIO Go SDK](https://pkg.go.dev/github.com/minio/minio-go/v7):

```go
package main

import (
    "context"
    "log"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "os"
)

func main() {
    endpoint := os.Getenv("MINIO_ENDPOINT")
    accessKey := os.Getenv("MINIO_ACCESS_KEY")
    secretKey := os.Getenv("MINIO_SECRET_KEY")

    // Initialize MinIO client
    minioClient, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: false, // true if using HTTPS
    })
    if err != nil {
        log.Fatalln(err)
    }

    bucketName := "user-pics"

    // Create bucket if it doesn't exist
    ctx := context.Background()
    exists, err := minioClient.BucketExists(ctx, bucketName)
    if err != nil {
        log.Fatalln(err)
    }
    if !exists {
        if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
            log.Fatalln(err)
        }
    }

    // Upload a file
    info, err := minioClient.FPutObject(ctx, bucketName, "my-pic.jpg", "/tmp/my-pic.jpg", minio.PutObjectOptions{})
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("Uploaded:", info.Key)
}
```

### **How it works:**

1. Your app connects to MinIO using endpoint + credentials.
2. It can **create buckets**, **upload files**, **list files**, **download files** — just like AWS S3.
3. The files are stored inside MinIO (backed by local disk or a Kubernetes volume).

---

### **3. Workflow Summary**

```
[User uploads pic] -> [users service] -> [MinIO via S3 API] -> [stored in /data in MinIO pod]
[users service reads pic] -> [MinIO API] -> [send to user]
```

---

If you want, I can rewrite your **`users` deployment fully ready for MinIO** with **Persistent Volume** for MinIO, so you can try it locally in dev and persist files between pod restarts.

Do you want me to do that?
