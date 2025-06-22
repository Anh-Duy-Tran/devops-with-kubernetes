# Log Output Application

This application demonstrates the use of emptyDir volumes in Kubernetes to share data between containers within the same pod.

## Architecture

The application consists of two containers running in a single pod:

1. **Log Writer Container**: Generates a random string on startup and writes log entries with timestamp every 5 seconds to a shared file
2. **Log Reader Container**: Reads the log file and serves the content via HTTP GET endpoint

Both containers share an emptyDir volume mounted at `/usr/src/app/files/` to exchange log data.

## Features

- Single Go binary that can run in two modes based on command-line arguments
- **Writer mode**: `./log-output writer` - Generates logs to a file
- **Reader mode**: `./log-output reader` - Serves HTTP API to read logs
- EmptyDir volume for file sharing between containers
- RESTful API endpoints at `/` and `/status`

## API Response Format

```json
{
  "timestamp": "2023-12-07T10:30:45Z",
  "string": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Deployment

### Build and Deploy

```bash
# Build Docker image and deploy to k3d cluster
./build-and-deploy.sh
```

### Manual Deployment

```bash
# Build Docker image
docker build -t log-output:latest .

# Import to k3d cluster
k3d image import log-output:latest

# Apply Kubernetes manifests
kubectl apply -f manifests/
```

## Volume Behavior

The emptyDir volume has the following characteristics:

- **Lifecycle**: Tied to the pod lifecycle - data persists only while the pod is running
- **Scope**: Shared between all containers in the pod
- **Persistence**: Data is lost when the pod is deleted or moved to another node
- **Storage**: Uses node's local storage (disk or memory)

## Accessing the Application

Once deployed, the application is accessible via:

- **Direct Service**: `kubectl port-forward service/log-output-service 8080:8080`
- **Ingress**: If ingress controller is configured, access via the configured host/path

## Monitoring

```bash
# Check pod status
kubectl get pods -l app=log-output-app

# View logs from both containers
kubectl logs -l app=log-output-app -c log-writer -f
kubectl logs -l app=log-output-app -c log-reader -f

# View all logs from the pod
kubectl logs -l app=log-output-app --all-containers=true -f
```

## Testing

```bash
# Test the API endpoint
curl http://localhost:8080/status

# Test via port-forward
kubectl port-forward service/log-output-service 8080:8080
curl http://localhost:8080/
```

## Screenshot

### 1.7

![LogOutput App 1.7 Screenshot](./screenshots/1.7.png)

Add an endpoint to request the current status (timestamp and the random string) and an Ingress so that you can access it with a browser.

### 1.1

![Application Output](./screenshots/1.1.png)

_Example of the log output application running in Kubernetes, showing timestamped entries with UUID generated every 5 seconds._
