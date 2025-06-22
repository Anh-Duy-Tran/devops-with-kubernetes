#!/bin/bash

echo "=== Deploying HTTP-based Log Output and PingPong Demo ==="
echo ""

# Deploy PingPong service first
echo "1. Deploying PingPong application..."
cd PingPong
docker build -t pingpong:latest .
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml
cd ..

echo "   Waiting for PingPong to be ready..."
kubectl wait --for=condition=ready pod -l app=pingpong-app --timeout=60s

# Deploy LogOutput application
echo ""
echo "2. Deploying LogOutput application..."
cd LogOutput
docker build -t log-output:latest .
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml
cd ..

echo "   Waiting for LogOutput to be ready..."
kubectl wait --for=condition=ready pod -l app=log-output-app --timeout=60s

echo ""
echo "=== Deployment Complete! ==="
echo ""
echo "Architecture:"
echo "  ┌─────────┐    HTTP GET     ┌──────────────┐"
echo "  │ Browser │ ──────────────► │ LogOutput    │"
echo "  └─────────┘                 │  - Writer    │"
echo "                              │  - Reader    │"
echo "                              └──────┬───────┘"
echo "                                     │"
echo "                                     │ HTTP GET /pingpongcount"
echo "                                     ▼"
echo "                              ┌──────────────┐"
echo "                              │ PingPong     │"
echo "                              │ Service      │"
echo "                              └──────────────┘"
echo ""

# Show status
echo "Current Status:"
kubectl get pods -l 'app in (log-output-app,pingpong-app)'
echo ""

# Show recent logs
echo "Recent PingPong logs:"
kubectl logs -l app=pingpong-app --tail=5
echo ""
echo "Recent LogOutput logs:"
kubectl logs -l app=log-output-app --tail=5
echo ""

echo "To access the application:"
echo "  kubectl port-forward service/log-output-service 8080:8080"
echo "  Then visit: http://localhost:8080"
echo ""
echo "To test PingPong directly:"
echo "  kubectl port-forward service/pingpong-service 8081:8080"
echo "  Then visit: http://localhost:8081/pingpong"
echo "  Or check count: http://localhost:8081/pingpongcount"
echo ""
echo "Expected LogOutput response format:"
echo "  2020-03-30T12:15:17.705Z: 8523ecb1-c716-4cb6-a044-b9e83bb98e43."
echo "  Ping / Pongs: 3" 