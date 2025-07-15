# PingPong Application

A simple HTTP server that responds with a "pong" message and an incrementing counter.

## Original Implementation

The original implementation stored the counter in memory, which meant the counter would reset whenever the pod restarted.

## Exercise 2.7: StatefulSet with Postgres

This exercise implements the PingPong application with persistent storage using:
- **Postgres StatefulSet**: A single-replica Postgres database with persistent volume
- **Headless Service**: For direct pod access to the database
- **Dynamic Volume Provisioning**: Using K3s local-path storage

### Key Components

1. **Postgres StatefulSet** (`manifests/postgres-statefulset.yaml`):
   - Single replica Postgres database
   - Persistent volume with 1Gi storage
   - Uses K3s local-path storage class for dynamic provisioning
   - Headless service for direct pod access

2. **Updated PingPong Application** (`main.go`):
   - Connects to Postgres database on startup
   - Creates table and initializes counter if needed
   - Stores counter persistently in database
   - Handles database connection retries with backoff

### Database Schema

```sql
CREATE TABLE ping_counter (
    id SERIAL PRIMARY KEY,
    counter_value INTEGER NOT NULL DEFAULT 0
);
```

### Deployment

1. **Quick Deploy**:
   ```bash
   ./build-and-deploy.sh
   ```

2. **Manual Steps**:
   ```bash
   # Create namespace
   kubectl create namespace exercises

   # Deploy Postgres StatefulSet
   kubectl apply -f manifests/postgres-statefulset.yaml

   # Wait for Postgres to be ready
   kubectl wait --for=condition=ready pod/postgres-stset-0 -n exercises --timeout=300s

   # Build and deploy PingPong
   docker build -t pingpong:latest .
   kubectl apply -f manifests/deployment.yaml
   kubectl apply -f manifests/service.yaml
   ```

### Testing

1. **Run Test Suite**:
   ```bash
   ./test-postgres-setup.sh
   ```

2. **Manual Testing**:
   ```bash
   # Test database connection
   kubectl run -it --rm --restart=Never --image postgres:15 --namespace=exercises psql-debug -- \
     psql 'postgres://pingponguser:pingpongpass@postgres-stset-0.postgres-svc.exercises.svc.cluster.local:5432/pingpongdb'

   # Test PingPong application
   kubectl port-forward -n exercises service/pingpong-service 8080:8080
   curl http://localhost:8080/pingpong
   curl http://localhost:8080/pingpongcount
   ```

### Key Benefits of StatefulSet

1. **Persistent Identity**: Pod `postgres-stset-0` always has the same name and network identity
2. **Persistent Storage**: Each replica gets its own dedicated volume
3. **Data Safety**: Volumes persist even if StatefulSet is deleted
4. **Ordered Deployment**: Pods are created and deleted in order

### Database Configuration

- **Host**: `postgres-stset-0.postgres-svc.exercises.svc.cluster.local`
- **Port**: `5432`
- **Database**: `pingpongdb`
- **User**: `pingponguser`
- **Password**: `pingpongpass`

### Monitoring

```bash
# Check all resources
kubectl get all -n exercises

# Check persistent volumes
kubectl get pvc -n exercises

# Check logs
kubectl logs -f -n exercises deployment/pingpong-app
kubectl logs -f -n exercises statefulset/postgres-stset
```

## API Endpoints

- `GET /pingpong` - Returns pong message with counter and increments it
- `GET /pingpongcount` - Returns current counter value without incrementing
- `GET /` - Same as `/pingpong`

## Environment Variables

- `POSTGRES_HOST` - Database host (default: postgres-stset-0.postgres-svc.exercises.svc.cluster.local)
- `POSTGRES_PORT` - Database port (default: 5432)
- `POSTGRES_USER` - Database user (default: pingponguser)
- `POSTGRES_PASSWORD` - Database password (default: pingpongpass)
- `POSTGRES_DB` - Database name (default: pingpongdb)

