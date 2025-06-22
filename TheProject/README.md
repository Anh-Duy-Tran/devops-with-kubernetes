# Todo App

A simple web server that outputs "Server started in port NNNN" when started and serves as the foundation for a todo application. Features configurable port via environment variable and health check endpoint.

![Todo App Screenshot](./screenshots/1.5.png)

## Quick Start

### Prerequisites

- Docker
- k3d cluster
- kubectl configured to connect to your cluster

### Deploy to Kubernetes

1. Make the build script executable (if not already):

   ```bash
   chmod +x build-and-deploy.sh
   ```

2. Run the automated build and deploy script:

   ```bash
   ./build-and-deploy.sh
   ```

3. View the logs to see the server started message:

   ```bash
   kubectl logs -l app=todo-app -f
   ```

4. Test the application:
   ```bash
   kubectl port-forward service/todo-app-service 8080:80
   curl http://localhost:8080/
   ```

### What the script does:

- Builds the Docker image from the Go source code
- Imports the image into your k3d cluster
- Deploys the application with Kubernetes manifests
- Shows deployment status

## Endpoints

- `GET /` - Returns "Hello from Todo App!"
- `GET /health` - Health check endpoint (returns "OK")
- `GET /headers` - Returns request headers for debugging

## Environment Variables

- `PORT` - Port number for the server to listen on (default: 8080)
