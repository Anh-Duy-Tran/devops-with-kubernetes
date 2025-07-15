#!/bin/bash

echo "Building and deploying Log Output application with ConfigMap..."

# Set up kubeconfig for k3d cluster
echo "Setting up kubeconfig for k3d cluster..."
export KUBECONFIG=$(k3d kubeconfig write k3s-default)

# Verify cluster connection
echo "Verifying cluster connection..."
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo "❌ Error: Cannot connect to k3d cluster. Make sure k3d cluster is running."
    echo "Run: k3d cluster start k3s-default"
    exit 1
fi

# Create exercises namespace if it doesn't exist
echo "Creating exercises namespace..."
kubectl create namespace exercises --dry-run=client -o yaml | kubectl apply -f -

# Build the Docker image
echo "Building Docker image..."
docker build -t log-output:latest .

# Import the image into k3d cluster
echo "Importing Docker image into k3d cluster..."
k3d image import log-output:latest -c k3s-default

# Create persistent volume directory on k3d node
echo "Creating persistent volume directory..."
docker exec k3d-k3s-default-agent-0 mkdir -p /tmp/kube

# Apply storage resources
echo "Applying storage resources..."
kubectl apply -f storage/persistentvolume.yaml
kubectl apply -f storage/persistentvolumeclaim.yaml

# Verify storage is ready
echo "Waiting for persistent volume to be bound..."
kubectl wait --for=condition=Bound pvc/shared-data-claim -n exercises --timeout=30s

# Apply the ConfigMap
echo "Applying ConfigMap..."
kubectl apply -f manifests/configmap.yaml

# Apply the deployment with ConfigMap
echo "Applying deployment..."
kubectl apply -f manifests/deployment.yaml

# Apply service and ingress
echo "Applying service and ingress..."
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/ingress.yaml

# Wait for deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available deployment/log-output-app -n exercises --timeout=60s

echo "Checking deployment status..."
kubectl get pods -l app=log-output-app -n exercises

echo "Deployment complete! The application now displays:"
echo "- file content: this text is from file"
echo "- env variable: MESSAGE=hello world"
echo "- timestamp and UUID"
echo "- Ping / Pongs count"

# Show logs to verify ConfigMap is working
echo ""
echo "Showing recent logs with ConfigMap data..."
kubectl logs deployment/log-output-app -c log-writer -n exercises --tail=5

echo ""
echo "To access the application:"
echo "kubectl port-forward service/log-output-service 8080:8080 -n exercises"
echo "Then visit http://localhost:8080"

echo ""
echo "To view live logs:"
echo "kubectl logs -f deployment/log-output-app -c log-writer -n exercises"
