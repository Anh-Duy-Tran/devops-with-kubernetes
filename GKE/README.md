# PingPong Application - GKE Deployment

This folder contains Kubernetes manifests for deploying the PingPong application to Google Kubernetes Engine (GKE).

## Prerequisites

1. **GKE Cluster**: You should have a running GKE cluster
2. **kubectl**: Configured to point to your GKE cluster
3. **Container Image**: You need to build and push the PingPong image to a container registry

## Building and Pushing the Image

### Option 1: Google Container Registry (GCR)

```bash
# From the PingPong directory
cd ../PingPong

# Build the image
docker build -t gcr.io/YOUR_PROJECT_ID/pingpong:latest .

# Push to GCR
docker push gcr.io/YOUR_PROJECT_ID/pingpong:latest
```

### Option 2: Docker Hub

```bash
# From the PingPong directory
cd ../PingPong

# Build the image
docker build -t YOUR_DOCKERHUB_USER/pingpong:latest .

# Push to Docker Hub
docker push YOUR_DOCKERHUB_USER/pingpong:latest
```

## Updating the Deployment

After building and pushing your image, update the `deployment.yaml` file:

```yaml
# In deployment.yaml, replace:
image: pingpong:latest

# With your actual image:
image: gcr.io/YOUR_PROJECT_ID/pingpong:latest
# OR
image: YOUR_DOCKERHUB_USER/pingpong:latest
```

## Deployment

### Quick Deploy

```bash
chmod +x deploy.sh
./deploy.sh
```

### Manual Deployment

```bash
# Deploy Postgres first
kubectl apply -f postgres-statefulset.yaml

# Wait for Postgres to be ready
kubectl wait --for=condition=ready pod/postgres-stset-0 --timeout=300s

# Deploy the application
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Accessing the Application

After deployment, get the external IP:

```bash
kubectl get svc pingpong-service
```

Once you have the external IP, you can access:

- `http://EXTERNAL_IP/pingpong` - Get ping pong response with counter
- `http://EXTERNAL_IP/pingpongcount` - Get just the counter

## Key Differences from Local Deployment

1. **LoadBalancer Service**: Uses LoadBalancer instead of ClusterIP for external access
2. **Storage**: Uses GKE's automatic persistent disk provisioning (no storage class specified)
3. **SubPath**: Added `subPath: postgres` to avoid PostgreSQL initialization issues
4. **No Namespace**: Deploys to default namespace instead of "exercises"
5. **Image Pull Policy**: Set to `Always` for production images

## Cleanup

To avoid costs, delete the cluster when not needed:

```bash
# Delete the application
kubectl delete -f .

# Delete the entire cluster (from gcloud CLI)
gcloud container clusters delete dwk-cluster --zone=YOUR_ZONE
```

## Cost Considerations

- LoadBalancer service incurs additional costs (~$20/month)
- Persistent disks for PostgreSQL storage
- Compute resources for the cluster nodes

Monitor your GCP billing to stay within your credits!
