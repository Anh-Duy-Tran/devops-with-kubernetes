#!/bin/bash

echo "Building Log Output Docker image..."
docker build -t log-output:latest .

echo "Applying Log Output Kubernetes manifests..."
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml

echo "Checking deployment status..."
kubectl get pods -l app=log-output-app -n exercises

echo "Log Output application deployed successfully!"
echo "Note: This application now communicates with PingPong via HTTP service calls."
echo "Make sure the PingPong service is running for full functionality."

# Show logs
echo "Showing recent logs..."
kubectl logs -l app=log-output-app -n exercises --tail=10

echo "To access the application:"
echo "kubectl port-forward service/log-output-service 8080:8080 -n exercises"
echo "Then visit http://localhost:8080"
