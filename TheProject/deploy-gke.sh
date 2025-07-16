#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID=${GCP_PROJECT_ID:-"your-project-id"}
CLUSTER_NAME=${GKE_CLUSTER_NAME:-"dwk-cluster"}
ZONE=${GKE_ZONE:-"us-central1-a"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}

echo -e "${GREEN}=== Deploying Todo App to GKE ===${NC}"

# Check if required tools are installed
echo -e "${YELLOW}Checking prerequisites...${NC}"
command -v kubectl >/dev/null 2>&1 || { echo -e "${RED}kubectl is required but not installed. Aborting.${NC}" >&2; exit 1; }
command -v gcloud >/dev/null 2>&1 || { echo -e "${RED}gcloud is required but not installed. Aborting.${NC}" >&2; exit 1; }

# Check if kubectl is pointing to the right cluster
echo -e "${YELLOW}Checking kubectl context...${NC}"
CURRENT_CONTEXT=$(kubectl config current-context)
echo "Current kubectl context: $CURRENT_CONTEXT"

# Build and push images if requested
if [ "$1" = "--build" ]; then
    echo -e "${YELLOW}Building and pushing Docker images...${NC}"
    
    # Build main app
    echo "Building todo-app..."
    docker build -t gcr.io/${PROJECT_ID}/todo-app:${IMAGE_TAG} .
    docker push gcr.io/${PROJECT_ID}/todo-app:${IMAGE_TAG}
    
    # Build backend (assuming there's a backend Dockerfile)
    if [ -d "todo-backend" ]; then
        echo "Building todo-backend..."
        cd todo-backend
        docker build -t gcr.io/${PROJECT_ID}/todo-backend:${IMAGE_TAG} .
        docker push gcr.io/${PROJECT_ID}/todo-backend:${IMAGE_TAG}
        cd ..
    fi
    
    echo -e "${GREEN}Images built and pushed successfully!${NC}"
fi

# Update kustomization with current project ID and image tag
echo -e "${YELLOW}Updating kustomization with project ID: ${PROJECT_ID} and tag: ${IMAGE_TAG}${NC}"

# Backup original kustomization
cp kustomization.yaml kustomization.yaml.backup

# Update the existing kustomization for GKE deployment
sed -i.tmp "s|newName: gcr.io/your-project-id/todo-app|newName: gcr.io/${PROJECT_ID}/todo-app|g" kustomization.yaml
sed -i.tmp "s|newName: gcr.io/your-project-id/todo-backend|newName: gcr.io/${PROJECT_ID}/todo-backend|g" kustomization.yaml
sed -i.tmp "s|newTag: latest|newTag: ${IMAGE_TAG}|g" kustomization.yaml

# Add LoadBalancer patch for GKE
cat >> kustomization.yaml << EOF

# GKE-specific patches
patches:
  # Change main service to LoadBalancer for external access
  - target:
      kind: Service
      name: todo-app-service
    patch: |-
      - op: replace
        path: /spec/type
        value: LoadBalancer
EOF

# Deploy using Kustomize
echo -e "${YELLOW}Deploying to GKE using Kustomize...${NC}"
kubectl apply -k .

# Wait for deployment
echo -e "${YELLOW}Waiting for deployments to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/todo-app -n project
kubectl wait --for=condition=available --timeout=300s deployment/todo-backend -n project

# Get service information
echo -e "${GREEN}Deployment completed!${NC}"
echo -e "${YELLOW}Service information:${NC}"
kubectl get services -n project

# Get LoadBalancer IP
echo -e "${YELLOW}Waiting for LoadBalancer IP...${NC}"
echo "LoadBalancer IP:"
kubectl get service todo-app-service -n project -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
echo ""

# Restore original kustomization
echo -e "${YELLOW}Restoring original kustomization...${NC}"
mv kustomization.yaml.backup kustomization.yaml
rm -f kustomization.yaml.tmp

echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo -e "${YELLOW}Your app should be accessible via the LoadBalancer IP shown above${NC}" 