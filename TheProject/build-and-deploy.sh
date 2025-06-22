#!/bin/bash

set -e

echo "Building Docker image for todo-app..."

# Build the Docker image
docker build -t todo-app:latest .

echo "Docker image built successfully!"

# Import the image into k3d cluster
echo "Importing image into k3d cluster..."
k3d image import todo-app:latest

echo "Applying Kubernetes manifests..."
# Apply PV and PVC first
kubectl apply -f storage

# Check PVC status (kubectl wait for PVC binding is unreliable)
echo "Checking PVC status..."
for i in {1..12}; do
  PVC_STATUS=$(kubectl get pvc todo-app-images-pvc -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")
  if [ "$PVC_STATUS" = "Bound" ]; then
    echo "PVC is bound successfully!"
    break
  elif [ "$i" -eq 12 ]; then
    echo "Warning: PVC not bound after 60 seconds, but continuing deployment..."
    kubectl get pvc todo-app-images-pvc
    break
  else
    echo "Waiting for PVC to bind... (attempt $i/12)"
    sleep 5
  fi
done

# Apply the rest of the resources
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml

echo "Checking deployment status..."
kubectl get pv
kubectl get pvc
kubectl get deployments
kubectl get pods -l app=todo-app
kubectl get services

echo ""
echo "Deployment completed!"
echo "Image caching is enabled with persistent storage!"
echo ""
echo "Useful commands:"
echo "  View logs: kubectl logs -l app=todo-app -f"
echo "  Check pod status: kubectl get pods -l app=todo-app"
echo "  Port forward: kubectl port-forward service/todo-app-service 8080:80"
echo "  Test container restart: curl http://localhost:8080/shutdown (after port-forward)"
echo "  Check persistent volume: kubectl exec -it <pod-name> -- ls -la /app/images/"
