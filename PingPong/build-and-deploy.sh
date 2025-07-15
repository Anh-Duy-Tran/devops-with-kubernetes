#!/bin/bash

echo "=== Building and Deploying PingPong with Postgres StatefulSet ==="

# Create namespace if it doesn't exist
echo "Creating namespace..."
kubectl create namespace exercises --dry-run=client -o yaml | kubectl apply -f -

# Deploy Postgres StatefulSet first
echo "Deploying Postgres StatefulSet..."
kubectl apply -f manifests/postgres-statefulset.yaml

# Wait for Postgres to be ready
echo "Waiting for Postgres to be ready..."
kubectl wait --for=condition=ready pod/postgres-stset-0 -n exercises --timeout=300s

# Check if Postgres is running
echo "Checking Postgres status..."
kubectl get pods -n exercises -l app=postgres

# Build the Docker image
echo "Building PingPong Docker image..."
docker build -t pingpong:latest .

# Deploy the PingPong application
echo "Deploying PingPong application..."
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml

# Wait for PingPong app to be ready
echo "Waiting for PingPong application to be ready..."
kubectl wait --for=condition=available deployment/pingpong-app -n exercises --timeout=300s

# Show the deployment status
echo "=== Deployment Status ==="
kubectl get all -n exercises

echo ""
echo "=== PersistentVolumeClaims ==="
kubectl get pvc -n exercises

echo ""
echo "PingPong application with Postgres deployed successfully!"
echo "HTTP endpoints available:"
echo "  /pingpong - Increments counter and returns pong message"
echo "  /pingpongcount - Returns current counter value"

# Show logs
echo ""
echo "=== Recent Logs ==="
echo "PingPong application logs:"
kubectl logs -l app=pingpong-app -n exercises --tail=10

echo ""
echo "=== Testing Instructions ==="
echo "To test the database connection:"
echo "kubectl run -it --rm --restart=Never --image postgres:15 --namespace=exercises psql-debug -- psql 'postgres://pingponguser:pingpongpass@postgres-stset-0.postgres-svc.exercises.svc.cluster.local:5432/pingpongdb'"

echo ""
echo "To test the PingPong application:"
echo "kubectl port-forward -n exercises service/pingpong-service 8080:8080"
echo "Then visit: http://localhost:8080/pingpong"

echo ""
echo "To monitor:"
echo "kubectl logs -f -n exercises deployment/pingpong-app"
echo "kubectl logs -f -n exercises statefulset/postgres-stset"
