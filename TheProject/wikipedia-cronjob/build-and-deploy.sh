#!/bin/bash

set -e

echo "Building Wikipedia Todo Generator Docker image..."

# Build the Docker image
docker build -t wikipedia-todo-generator:latest .

echo "Docker image built successfully!"

# Load image into k3d cluster (if using k3d)
if command -v k3d &> /dev/null; then
    echo "Loading image into k3d cluster..."
    k3d image import wikipedia-todo-generator:latest
    echo "Image loaded into k3d cluster!"
fi

echo "Deploying CronJob..."

# Apply the CronJob manifest
kubectl apply -f ../manifests/wikipedia-todo-cronjob.yaml

echo "CronJob deployed successfully!"

# Show the status
echo "Checking CronJob status..."
kubectl get cronjobs -n project

echo ""
echo "To manually trigger the job for testing:"
echo "kubectl create job --from=cronjob/wikipedia-todo-generator manual-wikipedia-job -n project"

echo ""
echo "To view job logs:"
echo "kubectl logs -l app=wikipedia-todo-generator -n project" 