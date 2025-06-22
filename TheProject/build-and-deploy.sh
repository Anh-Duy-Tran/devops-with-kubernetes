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
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml

echo "Checking deployment status..."
kubectl get deployments
kubectl get pods -l app=todo-app
kubectl get services

echo ""
echo "Deployment completed!"
echo "To view logs, run: kubectl logs -l app=todo-app -f"
echo "To check pod status: kubectl get pods -l app=todo-app"
echo "To test the application: kubectl port-forward service/todo-app-service 8080:80"
