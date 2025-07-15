#!/bin/bash

set -e

echo "=== Deploying Applications with Persistent Volumes ==="
echo ""

# Ensure the k3d node directory exists
echo "Creating persistent volume directory in k3d node..."
docker exec k3d-k3s-default-agent-0 mkdir -p /tmp/kube

# Build and import LogOutput image
echo "Building LogOutput Docker image..."
cd /Users/duytran/Study/devOps-Kubernetes/LogOutput
docker build -t log-output:latest .
k3d image import log-output:latest

# Build and import PingPong image
echo "Building PingPong Docker image..."
cd /Users/duytran/Study/devOps-Kubernetes/PingPong
docker build -t pingpong:latest .
k3d image import pingpong:latest

# Go back to LogOutput directory
cd /Users/duytran/Study/devOps-Kubernetes/LogOutput

# Apply persistent volume resources
echo "Creating persistent volume resources..."
kubectl apply -f storage/persistentvolume.yaml
kubectl apply -f storage/persistentvolumeclaim.yaml

# Check PVC status (kubectl wait for PVC binding is unreliable)
echo "Checking PVC status..."
for i in {1..12}; do
  PVC_STATUS=$(kubectl get pvc shared-data-claim -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")
  if [ "$PVC_STATUS" = "Bound" ]; then
    echo "PVC is bound successfully!"
    break
  elif [ "$i" -eq 12 ]; then
    echo "Warning: PVC not bound after 60 seconds, but continuing deployment..."
    kubectl get pvc shared-data-claim
    break
  else
    echo "Waiting for PVC to bind... (attempt $i/12)"
    sleep 5
  fi
done

# Deploy applications
echo "Deploying PingPong application..."
kubectl apply -f ../PingPong/manifests/

echo "Deploying LogOutput application..."
kubectl apply -f manifests/

# Wait for deployments to be ready
echo "Waiting for deployments to be ready..."
kubectl rollout status deployment/pingpong-app -n exercises --timeout=120s
kubectl rollout status deployment/log-output-app -n exercises --timeout=120s

echo ""
echo "=== Deployment completed! ==="
echo ""
echo "Applications deployed:"
echo "  - PingPong: Saves request count to persistent volume"
echo "  - LogOutput: Reads ping-pong count and displays logs"
echo ""
echo "Persistent Volume Info:"
kubectl get pv,pvc -n exercises
echo ""
echo "Pod Status:"
kubectl get pods -l app=pingpong-app -n exercises -o wide
kubectl get pods -l app=log-output-app -n exercises -o wide
echo ""
echo "Services:"
kubectl get services -n exercises
echo ""
echo "Useful commands:"
echo "  Test PingPong: kubectl port-forward service/pingpong-service 8080:8080 -n exercises"
echo "  Test LogOutput: kubectl port-forward service/log-output-service 8081:8080 -n exercises"
echo "  View PingPong logs: kubectl logs -l app=pingpong-app -n exercises -f"
echo "  View LogOutput logs: kubectl logs -l app=log-output-app -n exercises --all-containers=true -f"
echo "  Check persistent volume: kubectl exec -it deployment/pingpong-app -n exercises -- ls -la /usr/src/app/files/"
