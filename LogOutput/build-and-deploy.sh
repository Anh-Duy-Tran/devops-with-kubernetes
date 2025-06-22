#!/bin/bash

set -e

echo "Building Docker image for log-output app with persistent volume support..."

# Build the Docker image
docker build -t log-output:latest .

echo "Docker image built successfully!"

# Import the image into k3d cluster
echo "Importing image into k3d cluster..."
k3d image import log-output:latest

echo "Applying Kubernetes manifests..."
kubectl apply -f manifests

echo "Waiting for deployment to be ready..."
kubectl rollout status deployment/log-output-app --timeout=60s

echo "Checking deployment status..."
kubectl get deployments
kubectl get pods -l app=log-output-app
kubectl get services

echo ""
echo "=== Deployment completed! ==="
echo "The application now runs with two containers and persistent volume support:"
echo "  - log-writer: Generates logs with ping-pong count to shared file"
echo "  - log-reader: Serves logs via HTTP API"
echo ""
echo "⚠️  Note: This deploys with persistent volume configuration."
echo "   For full persistent volume demo with PingPong integration, use:"
echo "   ./deploy-persistent.sh"
echo ""
echo "Useful commands:"
echo "  View writer logs: kubectl logs -l app=log-output-app -c log-writer -f"
echo "  View reader logs: kubectl logs -l app=log-output-app -c log-reader -f"
echo "  View all logs: kubectl logs -l app=log-output-app --all-containers=true -f"
echo "  Test API: kubectl port-forward service/log-output-service 8080:8080"
echo "  Run persistence tests: ./test-persistent.sh"
echo ""
echo "Persistent Volume is mounted at /usr/src/app/files/ in both containers"
