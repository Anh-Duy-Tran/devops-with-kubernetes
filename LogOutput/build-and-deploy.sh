#!/bin/bash

set -e

echo "Building Docker image for log-output app..."

# Build the Docker image
docker build -t log-output:latest .

echo "Docker image built successfully!"

# Import the image into k3d cluster
echo "Importing image into k3d cluster..."
k3d image import log-output:latest

echo "Applying Kubernetes manifests..."
kubectl apply -f manifests/deployment.yaml

echo "Checking deployment status..."
kubectl get deployments
kubectl get pods -l app=log-output-app
kubectl get services

echo ""
echo "Deployment completed!"
echo "To view logs, run: kubectl logs -l app=log-output-app -f"
echo "To check pod status: kubectl get pods -l app=log-output-app"
