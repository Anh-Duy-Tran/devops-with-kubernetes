# Persistent Volumes Implementation Guide

This document explains how the LogOutput and PingPong applications were modified to use Kubernetes Persistent Volumes for data sharing.

## Overview

The implementation demonstrates:
- **Cross-pod data sharing** using persistent volumes
- **Data persistence** across pod restarts and rescheduling
- **Real-world integration** between two separate applications

## Architecture Changes

### From EmptyDir to Persistent Volumes

**Before (EmptyDir):**
- Data shared only between containers in the same pod
- Data lost when pod is deleted or moved
- Filesystem tied to pod lifecycle

**After (Persistent Volumes):**
- Data shared between different pods/applications
- Data persists across pod restarts and moves
- Filesystem independent of pod lifecycle

## Implementation Details

### 1. Persistent Volume Configuration

**File:** `storage/persistentvolume.yaml`
```yaml
spec:
  storageClassName: shared-storage
  capacity:
    storage: 1Gi
  local:
    path: /tmp/kube  # Host directory in k3d node
  nodeAffinity:     # Required for local volumes
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - k3d-k3s-default-agent-0
```

### 2. Persistent Volume Claim

**File:** `storage/persistentvolumeclaim.yaml`
```yaml
spec:
  storageClassName: shared-storage  # Must match PV
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

### 3. Application Modifications

#### PingPong Application Changes

**New Features:**
- Saves counter to `/usr/src/app/files/pingpong_counter.txt`
- Loads counter on startup (persistence across restarts)
- Thread-safe file operations

**Key Functions:**
```go
func loadCounter() {
    // Load counter from file on startup
}

func saveCounter() {
    // Save counter to file after each request
}
```

#### LogOutput Application Changes

**New Features:**
- Reads ping-pong counter from shared volume
- Displays counter in required format
- Updates output format to match specification

**Key Changes:**
```go
func readPingPongCount() int {
    // Read counter from shared file
}

// Output format: "timestamp: uuid.\nPing / Pongs: count"
```

## File Structure

```
/tmp/kube/                          # Persistent volume mount point
├── pingpong_counter.txt           # PingPong counter (managed by PingPong app)
└── output.log                     # Log output (managed by LogOutput app)
```

## Data Flow

1. **User Request → PingPong**
   - HTTP request to `/pingpong`
   - Counter incremented
   - Counter saved to `pingpong_counter.txt`

2. **LogOutput Writer Loop**
   - Reads `pingpong_counter.txt`
   - Generates formatted log entry
   - Writes to `output.log`

3. **User Request → LogOutput**
   - HTTP request to `/`
   - Reads `output.log`
   - Returns formatted content

## Deployment Workflow

```bash
# 1. Infrastructure setup
docker exec k3d-k3s-default-agent-0 mkdir -p /tmp/kube

# 2. Storage resources
kubectl apply -f storage/persistentvolume.yaml
kubectl apply -f storage/persistentvolumeclaim.yaml

# 3. Applications
kubectl apply -f ../PingPong/manifests/
kubectl apply -f manifests/
```

## Testing Persistence

The implementation includes comprehensive tests:

### Automated Test Suite
```bash
./test-persistent.sh
```

### Manual Verification
1. Make requests to PingPong (counter increments)
2. Check LogOutput (displays current count)
3. Delete PingPong pod
4. Wait for pod recreation
5. Make request to PingPong (counter continues from previous value)

## Production Considerations

⚠️ **Important**: This implementation uses local persistent volumes, which are:
- **Not suitable for production** (data tied to specific node)
- **Single point of failure** (if node fails, data is lost)
- **Not scalable** (can't move between nodes)

### Production Alternatives:
- **Cloud Storage**: AWS EBS, GCE Persistent Disk, Azure Disk
- **Network Storage**: NFS, Ceph, GlusterFS
- **Storage Classes**: Dynamic provisioning with CSI drivers
- **Distributed Storage**: Rook, OpenEBS, Longhorn

## Benefits Achieved

✅ **Data Persistence**: Counter survives pod restarts  
✅ **Cross-Application Sharing**: Two apps share data seamlessly  
✅ **Real-world Scenario**: Demonstrates practical persistent volume usage  
✅ **Kubernetes Native**: Uses standard Kubernetes storage primitives  

## Output Format Compliance

The implementation produces the exact format specified:
```
2020-03-30T12:15:17.705Z: 8523ecb1-c716-4cb6-a044-b9e83bb98e43.
Ping / Pongs: 3
```

This format is achieved by:
- LogOutput writer generates timestamp and UUID
- Reads ping-pong count from shared volume
- Formats output according to specification
- LogOutput reader serves the formatted content 