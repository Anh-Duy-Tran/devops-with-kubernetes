# TheProject - Automatic Deployment Setup

This document describes the automatic deployment pipeline setup for TheProject using GitHub Actions and Google Kubernetes Engine (GKE).

## Overview

The project now includes automatic deployment using GitHub Actions that:
1. Builds Docker images for both the main todo-app and todo-backend
2. Pushes images to Google Artifact Registry
3. Deploys to GKE using Kustomize
4. **Creates separate environments for each branch**

## Components

### GitHub Actions Workflow
- **Location**: `.github/workflows/main.yaml`
- **Triggers**: 
  - Push to any branch with changes in the `TheProject/` directory
  - Manual workflow dispatch
- **Builds**: Two Docker images
  - Main todo application (`todo-app`)
  - Backend service (`todo-backend`)

### Required GitHub Secrets

You need to configure the following secrets in your GitHub repository:

1. **`GKE_PROJECT`**: Your Google Cloud Project ID
2. **`GKE_SA_KEY`**: Service Account JSON key with the following IAM roles:
   - Kubernetes Engine Service Agent
   - Storage Admin
   - Artifact Registry Administrator
   - Artifact Registry Create-on-Push Repository Administrator

### Kustomize Configuration

- **File**: `kustomization.yaml`
- **Image Placeholders**: 
  - `PROJECT/TODO-APP` → replaced with actual image during deployment
  - `PROJECT/TODO-BACKEND` → replaced with actual image during deployment
- **Namespace**: `project`

### Deployment Strategy

The main todo-app deployment uses `Recreate` strategy instead of the default `RollingUpdate` to handle the `ReadWriteOnce` Persistent Volume properly. This ensures the old pod is terminated before the new pod starts, preventing volume mounting conflicts.

## Branch-Specific Environments

The deployment pipeline automatically creates separate environments for each branch:

### Namespace Strategy
- **Main branch** (`main`) → Deployed to `project` namespace
- **Feature branches** → Deployed to namespace matching branch name (e.g., `feature-x` → `feature-x` namespace)

### How It Works
1. **Namespace Detection**: Pipeline determines target namespace based on branch name
2. **Namespace Creation**: Automatically creates namespace if it doesn't exist
3. **Isolated Deployment**: Each branch gets its own:
   - Pods and services
   - ConfigMaps and secrets  
   - Persistent volumes
   - Ingress rules

### Testing Feature Branches
To test this feature:
1. Create a new branch: `git checkout -b feature-test`
2. Make changes to your application
3. Push the branch: `git push origin feature-test`
4. Pipeline will automatically deploy to `feature-test` namespace

### Accessing Branch Environments
- **Main**: http://35.228.86.161/ (LoadBalancer service)
- **Feature branches**: Use `kubectl port-forward` or configure separate ingress

### Cleanup
To remove a feature branch environment:
```bash
kubectl delete namespace <branch-name>
```

## Prerequisites

1. **Google Cloud Setup**:
   - GKE cluster named `dwk-cluster` in `europe-north1-b`
   - Docker repository in Google Artifact Registry
   - Service account with proper permissions

2. **Kubernetes Cluster**:
   - Namespace `project` should exist
   - Persistent Volume storage class `manual` should be available

## Deployment Process

1. Push changes to the repository
2. GitHub Actions automatically:
   - Builds Docker images
   - Pushes to Artifact Registry  
   - Updates Kubernetes deployments
   - Verifies rollout status

## Image Naming Convention

Images are tagged with the format:
```
{REGISTRY}/{PROJECT_ID}/{REPOSITORY}/{IMAGE_NAME}:{BRANCH}-{COMMIT_SHA}
```

Example:
```
europe-north1-docker.pkg.dev/my-project/my-repository/todo-app:main-abc123def
```

## Troubleshooting

### Label Selector Immutability Error

If you encounter errors like "spec.selector: Invalid value... field is immutable", this happens when trying to update existing deployments with new labels in selectors.

**Solutions:**

1. **Delete and recreate deployments** (recommended for development):
   ```bash
   kubectl delete deployment todo-app todo-backend -n project
   kubectl delete statefulset postgres-stset -n project
   # Then run the pipeline again
   ```

2. **Manual cleanup and redeploy**:
   ```bash
   kubectl delete -k TheProject/
   kubectl apply -k TheProject/
   ```

3. **Force replace** (use with caution):
   ```bash
   kubectl replace --force -f <(kustomize build TheProject/)
   ```

**Note**: The configuration has been updated to use `labels` instead of `commonLabels` to prevent this issue in future deployments.

## Manual Deployment

To deploy manually:
```bash
cd TheProject
kubectl apply -k .
```

To see what Kustomize would deploy without applying:
```bash
kubectl kustomize .
``` 