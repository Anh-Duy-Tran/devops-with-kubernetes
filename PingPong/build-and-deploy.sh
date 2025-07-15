#!/bin/bash

echo "Building PingPong Docker image..."
docker build -t pingpong:latest .

echo "Applying PingPong Kubernetes manifests..."
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml

echo "Checking deployment status..."
kubectl get pods -l app=pingpong-app -n exercises

echo "PingPong application deployed successfully!"
echo "HTTP endpoints available:"
echo "  /pingpong - Increments counter and returns pong message"
echo "  /pingpongcount - Returns current counter value (used by LogOutput app)"

# Show logs
echo "Showing recent logs..."
kubectl logs -l app=pingpong-app -n exercises --tail=10

echo "To test the application:"
echo "kubectl port-forward service/pingpong-service 8080:8080 -n exercises"
echo "Then visit http://localhost:8080/pingpong"
echo "  http://localhost:8080/pingpongcount - to get current count"
