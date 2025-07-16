#!/bin/bash

echo "=== Deploying PingPong Application to GKE ==="

# Check if kubectl is configured for GKE
echo "Checking cluster connection..."
kubectl cluster-info

# Deploy Postgres StatefulSet first
echo "Deploying Postgres StatefulSet..."
kubectl apply -f postgres-statefulset.yaml

# Wait for Postgres to be ready
echo "Waiting for Postgres to be ready..."
kubectl wait --for=condition=ready pod/postgres-stset-0 --timeout=300s

# Check if Postgres is running
echo "Checking Postgres status..."
kubectl get pods -l app=postgres

# Deploy the PingPong application
echo "Deploying PingPong application..."
kubectl apply -f deployment.yaml

# Deploy the LoadBalancer service
echo "Deploying LoadBalancer service..."
kubectl apply -f service.yaml

# Wait for PingPong app to be ready
echo "Waiting for PingPong application to be ready..."
kubectl wait --for=condition=available deployment/pingpong-app --timeout=300s

# Wait for external IP
echo "Waiting for LoadBalancer to get external IP..."
echo "This may take a few minutes..."
kubectl get svc pingpong-service --watch
