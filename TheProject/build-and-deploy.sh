#!/bin/bash

set -e

echo "=== Building and Deploying TheProject with Postgres Database ==="

# Create namespace if it doesn't exist
echo "Creating namespace..."
kubectl create namespace project --dry-run=client -o yaml | kubectl apply -f -

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

# Apply Secret and ConfigMap first
echo "Applying Secret for database credentials..."
kubectl apply -f manifests/secret.yaml

echo "Applying ConfigMap..."
kubectl apply -f manifests/configmap.yaml

# Deploy Postgres StatefulSet
echo "Deploying Postgres StatefulSet..."
kubectl apply -f manifests/postgres-statefulset.yaml

# Wait for Postgres to be ready
echo "Waiting for Postgres to be ready..."
kubectl wait --for=condition=ready pod/postgres-stset-0 -n project --timeout=300s

# Check if Postgres is running
echo "Checking Postgres status..."
kubectl get pods -n project -l app=postgres

# Apply PV and PVC for image storage
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

# Deploy the backend (depends on database)
echo "Deploying todo-backend..."
kubectl apply -f manifests/todo-backend-deployment.yaml
kubectl apply -f manifests/todo-backend-service.yaml

# Wait for backend to be ready
echo "Waiting for todo-backend to be ready..."
kubectl wait --for=condition=available deployment/todo-backend -n project --timeout=300s

# Deploy the frontend
echo "Deploying todo-app frontend..."
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml

# Wait for frontend to be ready
echo "Waiting for todo-app to be ready..."
kubectl wait --for=condition=available deployment/todo-app -n project --timeout=300s

echo "=== Deployment Status ==="
kubectl get all -n project

echo ""
echo "=== PersistentVolumeClaims ==="
kubectl get pvc -n project

echo ""
echo "=== Secrets and ConfigMaps ==="
kubectl get secrets,configmaps -n project

echo ""
echo "=== TheProject with Postgres Database deployed successfully! ==="
echo ""
echo "Database Configuration (from Secret):"
echo "  Database: $(kubectl get secret postgres-secret -n project -o jsonpath='{.data.POSTGRES_DB}' | base64 -d)"
echo "  User: $(kubectl get secret postgres-secret -n project -o jsonpath='{.data.POSTGRES_USER}' | base64 -d)"
echo ""
echo "Application Configuration (from ConfigMap):"
echo "  Frontend Port: $(kubectl get configmap todo-app-config -n project -o jsonpath='{.data.FRONTEND_PORT}')"
echo "  Backend Port: $(kubectl get configmap todo-app-config -n project -o jsonpath='{.data.BACKEND_PORT}')"
echo "  Image URL: $(kubectl get configmap todo-app-config -n project -o jsonpath='{.data.IMAGE_URL}')"
echo "  Cache Duration: $(kubectl get configmap todo-app-config -n project -o jsonpath='{.data.CACHE_DURATION_MINUTES}') minutes"
echo ""
echo "=== Recent Logs ==="
echo "Todo-backend logs:"
kubectl logs -l app=todo-backend -n project --tail=5

echo ""
echo "=== Testing Instructions ==="
echo "To test the database connection:"
echo "kubectl run -it --rm --restart=Never --image postgres:15 --namespace=project psql-debug -- psql 'postgres://todouser:todopass123@postgres-stset-0.postgres-svc.project.svc.cluster.local:5432/tododb'"

echo ""
echo "=== Useful Commands ==="
echo "  View todo-app logs: kubectl logs -l app=todo-app -n project -f"
echo "  View todo-backend logs: kubectl logs -l app=todo-backend -n project -f"
echo "  View postgres logs: kubectl logs -l app=postgres -n project -f"
echo "  Check pod status: kubectl get pods -n project"
echo "  Port forward todo-app: kubectl port-forward service/todo-app-service 8080:80 -n project"
echo "  Port forward todo-backend: kubectl port-forward service/todo-backend-service 3001:3001 -n project"
echo "  Test todo-backend API: curl http://localhost:3001/todos (after port-forward)"
echo "  Test stats endpoint: curl http://localhost:3001/stats (after port-forward)"
echo "  Check database: kubectl exec -it postgres-stset-0 -n project -- psql -U todouser -d tododb -c 'SELECT * FROM todos;'"
echo "  Monitor database: kubectl exec -it postgres-stset-0 -n project -- psql -U todouser -d tododb -c '\\dt'"
