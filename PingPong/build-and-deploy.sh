#!/bin/bash

set -e

echo "Building Docker image for pingpong app..."

# Build the Docker image
docker build -t pingpong:latest .

echo "Docker image built successfully!"

# Import the image into k3d cluster
echo "Importing image into k3d cluster..."
k3d image import pingpong:latest

echo "Applying Kubernetes manifests..."
kubectl apply -f manifests

echo "Checking deployment status..."
kubectl get deployments
kubectl get pods -l app=pingpong-app
kubectl get services

echo ""
echo "Deployment completed!"
echo "To view logs, run: kubectl logs -l app=pingpong-app -f"
echo "To check pod status: kubectl get pods -l app=pingpong-app"
