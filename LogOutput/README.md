# Log Output Application with Persistent Volumes

This application demonstrates the use of persistent volumes in Kubernetes to share data between applications across different pods. It integrates with a PingPong application to display shared counter data.

## Architecture

The application consists of:

### LogOutput Application (Single Pod, Two Containers)

1. **Log Writer Container**: Generates a random string on startup and writes log entries with timestamp and ping-pong count every 5 seconds to a shared file
2. **Log Reader Container**: Reads the log file and serves the content via HTTP GET endpoint

### PingPong Application (Single Pod, Single Container)

1. **PingPong Container**: Handles HTTP requests, increments a counter, and saves the counter to a shared persistent volume

Both applications share a **persistent volume** mounted at `/usr/src/app/files/` to exchange data.

## Features

- **Persistent Volume**: Data survives pod restarts and can be shared between applications
- **Single Go Binary**: LogOutput can run in two modes based on command-line arguments
  - **Writer mode**: `./log-output writer` - Generates logs with ping-pong count to a file
  - **Reader mode**: `./log-output reader` - Serves HTTP API to read logs
- **Cross-Application Data Sharing**: PingPong saves counter, LogOutput reads and displays it
- **RESTful API**: LogOutput serves data in the required format

## Output Format

The LogOutput application serves data in this format:

```
2020-03-30T12:15:17.705Z: 8523ecb1-c716-4cb6-a044-b9e83bb98e43.
Ping / Pongs: 3
```

## Persistent Volume Structure

```
/tmp/kube/                     # Host directory (k3d node)
├── pingpong_counter.txt       # PingPong counter file
└── output.log                 # LogOutput log file
```

## Deployment

### Quick Deploy (Recommended)

```bash
# Deploy both applications with persistent volumes
./deploy-persistent.sh
```

### Manual Deployment

```bash
# 1. Create persistent volume directory in k3d node
docker exec k3d-k3s-default-agent-0 mkdir -p /tmp/kube

# 2. Build and import images
docker build -t log-output:latest .
k3d image import log-output:latest

cd ../PingPong
docker build -t pingpong:latest .
k3d image import pingpong:latest

# 3. Apply persistent volume resources
cd ../LogOutput
kubectl apply -f storage/persistentvolume.yaml
kubectl apply -f storage/persistentvolumeclaim.yaml

# 4. Deploy applications
kubectl apply -f ../PingPong/manifests/
kubectl apply -f manifests/
```

## Persistent Volume Characteristics

- **Type**: Local persistent volume (not suitable for production)
- **Storage Class**: `shared-storage`
- **Capacity**: 1Gi
- **Access Mode**: ReadWriteOnce
- **Lifecycle**: Independent of pod lifecycle - data persists across pod restarts
- **Node Affinity**: Bound to specific k3d node (`k3d-k3s-default-agent-0`)

## Testing

### Comprehensive Test Suite

```bash
# Run complete test suite
./test-persistent.sh
```

### Manual Testing

```bash
# Test PingPong application
kubectl port-forward service/pingpong-service 8080:8080
curl http://localhost:8080/pingpong

# Test LogOutput application
kubectl port-forward service/log-output-service 8081:8080
curl http://localhost:8081/

# Verify data persistence
kubectl delete pod -l app=pingpong-app
kubectl wait --for=condition=ready pod -l app=pingpong-app
curl http://localhost:8080/pingpong  # Counter should continue from previous value
```

## Monitoring

```bash
# Check persistent volume status
kubectl get pv,pvc

# Check pod status
kubectl get pods -l app=pingpong-app
kubectl get pods -l app=log-output-app

# View application logs
kubectl logs -l app=pingpong-app -f
kubectl logs -l app=log-output-app -c log-writer -f
kubectl logs -l app=log-output-app -c log-reader -f

# Check shared files
kubectl exec deployment/pingpong-app -- ls -la /usr/src/app/files/
kubectl exec deployment/log-output-app -c log-writer -- ls -la /usr/src/app/files/
```

## Data Flow

1. **PingPong Application**:

   - Receives HTTP requests
   - Increments counter
   - Saves counter to `/usr/src/app/files/pingpong_counter.txt`

2. **LogOutput Writer**:

   - Generates UUID on startup
   - Every 5 seconds: reads ping-pong counter, writes formatted log entry
   - Saves to `/usr/src/app/files/output.log`

3. **LogOutput Reader**:
   - Serves HTTP requests
   - Reads and returns content from `/usr/src/app/files/output.log`

## Cleanup

```bash
# Remove deployments
kubectl delete -f manifests/
kubectl delete -f ../PingPong/manifests/

# Remove persistent volume resources
kubectl delete -f storage/
```

## Screenshot

### 1.7

![LogOutput App 1.7 Screenshot](./screenshots/1.7.png)

Add an endpoint to request the current status (timestamp and the random string) and an Ingress so that you can access it with a browser.

### 1.1

![Application Output](./screenshots/1.1.png)

_Example of the log output application running in Kubernetes, showing timestamped entries with UUID generated every 5 seconds._
