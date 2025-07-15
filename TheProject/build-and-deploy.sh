#!/bin/bash

set -e

echo "Building Docker images..."

# Build the todo-app image
echo "Building todo-app..."
docker build -t todo-app:latest .

# Build the todo-backend image
echo "Building todo-backend..."
docker build -t todo-backend:latest ./todo-backend/

echo "Docker images built successfully!"

# Import the images into k3d cluster
echo "Importing images into k3d cluster..."
k3d image import todo-app:latest
k3d image import todo-backend:latest

echo "Applying Kubernetes manifests..."
# Apply PV and PVC first
kubectl apply -f storage

# Check PVC status (kubectl wait for PVC binding is unreliable)
echo "Checking PVC status..."
for i in {1..12}; do
  PVC_STATUS=$(kubectl get pvc todo-app-images-pvc -n project -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")
  if [ "$PVC_STATUS" = "Bound" ]; then
    echo "PVC is bound successfully!"
    break
  elif [ "$i" -eq 12 ]; then
    echo "Warning: PVC not bound after 60 seconds, but continuing deployment..."
    kubectl get pvc todo-app-images-pvc -n project
    break
  else
    echo "Waiting for PVC to bind... (attempt $i/12)"
    sleep 5
  fi
done

# Apply the rest of the resources
kubectl apply -f manifests/todo-backend-deployment.yaml
kubectl apply -f manifests/todo-backend-service.yaml
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml

echo "Checking deployment status..."
kubectl get pv
kubectl get pvc -n project
kubectl get deployments -n project
kubectl get pods -l app=todo-app -n project
kubectl get pods -l app=todo-backend -n project
kubectl get services -n project

echo ""
echo "Deployment completed!"
echo "Image caching is enabled with persistent storage!"
echo ""
echo "Useful commands:"
echo "  View todo-app logs: kubectl logs -l app=todo-app -n project -f"
echo "  View todo-backend logs: kubectl logs -l app=todo-backend -n project -f"
echo "  Check pod status: kubectl get pods -n project"
echo "  Port forward todo-app: kubectl port-forward service/todo-app-service 8080:80 -n project"
echo "  Port forward todo-backend: kubectl port-forward service/todo-backend-service 3001:3001 -n project"
echo "  Test container restart: curl http://localhost:8080/shutdown (after port-forward)"
echo "  Check persistent volume: kubectl exec -it <pod-name> -n project -- ls -la /usr/src/app/images/"
echo "  Test todo-backend directly: curl http://localhost:3001/todos (after port-forward)"
