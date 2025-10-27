# K8S-TUI Test Resources

This document describes the test Kubernetes resources provided for manual testing of k8s-tui.

## Quick Start

### Apply Test Resources

```bash
# Apply all test resources
kubectl apply -f test-resources.yaml

# Wait for resources to be created
kubectl wait --for=condition=ready pod -l app=nginx -n k8s-tui-test --timeout=60s
```

### Run K8S-TUI

```bash
# Build the TUI
go build -o k8s-tui cmd/k8s-tui/main.go

# Run in the test namespace
./k8s-tui -n k8s-tui-test

# Or run in all namespaces (default)
./k8s-tui
```

### Clean Up

```bash
# Delete all test resources
kubectl delete -f test-resources.yaml

# Or delete just the namespace (cascading delete)
kubectl delete namespace k8s-tui-test
```

## Test Resources Overview

### Namespace
- **k8s-tui-test**: Dedicated namespace for all test resources

### Deployments (3)

| Name | Replicas | Status | Purpose |
|------|----------|--------|---------|
| `nginx-healthy` | 3 | Healthy | Test multi-replica healthy deployment with RollingUpdate strategy |
| `redis-single` | 1 | Healthy | Test single-replica deployment |
| `app-scaled-down` | 0 | Scaled to 0 | Test zero-replica deployment (all unavailable) |
| `failing-deployment` | 1 | ImagePullBackOff | Test error states and warning symbols |

### StatefulSets (2)

| Name | Replicas | Update Strategy | Purpose |
|------|----------|----------------|---------|
| `postgres-cluster` | 3 | RollingUpdate | Test multi-replica statefulset with persistent volumes |
| `redis-cluster` | 2 | OnDelete | Test statefulset with OnDelete strategy |

### Services (6)

| Name | Type | Ports | Purpose |
|------|------|-------|---------|
| `nginx-service` | ClusterIP | 80/TCP | Standard ClusterIP service |
| `nginx-nodeport` | NodePort | 80:30080/TCP | NodePort service (accessible externally) |
| `postgres-headless` | ClusterIP (None) | 5432/TCP | Headless service for StatefulSet |
| `redis-cluster-headless` | ClusterIP (None) | 6379/TCP | Headless service for StatefulSet |
| `redis-service` | ClusterIP | 6379/TCP | Standard ClusterIP for standalone pods |
| `multi-port-service` | ClusterIP | 80/TCP, 443/TCP | Multi-port service |

### Standalone Pods (3)

| Name | Containers | Status | Purpose |
|------|-----------|--------|---------|
| `multi-container-pod` | 3 (main-app, sidecar, logger) | Running | Test multi-container pod (3 containers) |
| `standalone-busybox` | 1 | Running | Simple standalone pod |
| `completed-job-pod` | 1 | Succeeded | Pod that completes successfully |

## Manual Testing Scenarios

### 1. Test Tab Navigation
**Goal**: Verify switching between resource types

1. Launch k8s-tui: `./k8s-tui -n k8s-tui-test`
2. Press `Tab` to cycle through tabs: Pods → Services → Deployments → StatefulSets
3. Press `Shift+Tab` to cycle backwards
4. Press `1`, `2`, `3`, `4` to jump directly to specific tabs

**Expected**:
- Tab highlighting changes
- Resource list updates to show the correct resource type
- Status symbols display correctly for each type

### 2. Test Pod List & Status Symbols
**Goal**: Verify pod listing and status indicators

1. Navigate to Pods tab (`1`)
2. Observe the different pod states:
   - ✓ (green) - Running pods from nginx-healthy deployment
   - ✓ (green) - Multi-container pod (all containers ready)
   - ✓ (green) - Succeeded pod (completed-job-pod)
   - ○ (yellow) - Pending/creating pods (if any are still starting)
   - ⚠ (warning) - ImagePullBackOff from failing-deployment

**Expected**:
- All pods visible in list
- Correct status symbols
- Ready count shows correctly (e.g., "3/3" for multi-container-pod)
- Age displays in human-readable format (e.g., "5m", "2h")

### 3. Test Service List & Types
**Goal**: Verify service listing and type differentiation

1. Navigate to Services tab (`2`)
2. Observe different service types:
   - ClusterIP services (nginx-service, redis-service, multi-port-service)
   - NodePort service (nginx-nodeport) - shows NodePort in ports column
   - Headless services (postgres-headless, redis-cluster-headless) - ClusterIP shows "None"

**Expected**:
- All 6 services visible
- Type column shows correct service type
- Ports column shows formatted ports (e.g., "80/TCP", "80:30080/TCP")
- External IP shows "<none>" for services without external IPs

### 4. Test Deployment List & Replica Status
**Goal**: Verify deployment listing and replica counts

1. Navigate to Deployments tab (`3`)
2. Observe different deployment states:
   - ✓ `nginx-healthy` (3/3 ready, 3 up-to-date, 3 available)
   - ✓ `redis-single` (1/1 ready)
   - ✗ `app-scaled-down` (0/0 - no replicas)
   - ○ or ✗ `failing-deployment` (0/1 - not ready due to image pull error)

**Expected**:
- Ready column shows correct format "X/Y"
- Up-to-date and Available columns show correct counts
- Status symbols reflect health (✓ = all ready, ○ = partial, ✗ = none ready)
- Strategy column shows "RollingUpdate" or "Recreate"

### 5. Test StatefulSet List & Update Strategies
**Goal**: Verify statefulset listing

1. Navigate to StatefulSets tab (`4`)
2. Observe statefulsets:
   - ✓ or ○ `postgres-cluster` (may take time to reach 3/3 ready due to PVCs)
   - ✓ `redis-cluster` (2/2 ready)

**Expected**:
- Ready column shows "X/Y" format
- Strategy column shows "RollingUpdate" or "OnDelete"
- Status symbols reflect health

### 6. Test Detail View
**Goal**: Verify detailed resource information display

1. Select any resource (use arrow keys to navigate)
2. Press `Enter` to view details
3. Press `Esc` to return to list view
4. Test with different resource types (Pods, Services, Deployments, StatefulSets)

**Expected**:
- Detail view shows comprehensive information
- Key-value pairs are properly formatted
- For Pods: shows containers, node, IP, etc.
- For Services: shows cluster IP, ports, selectors
- For Deployments: shows replicas, strategy, conditions
- For StatefulSets: shows replicas, update strategy

### 7. Test Namespace Switching
**Goal**: Verify namespace selector functionality

1. Press `n` to open namespace selector
2. Use arrow keys to navigate namespaces
3. Press `Enter` to switch to a different namespace (e.g., "default", "kube-system")
4. Press `n` again and switch back to "k8s-tui-test"
5. Press `Esc` to cancel namespace selection

**Expected**:
- Modal overlay appears when pressing `n`
- Namespaces list is visible and navigable
- Switching namespace updates the resource list
- Resources from new namespace are displayed
- Cancel works correctly

### 8. Test Search/Filter
**Goal**: Verify search functionality

1. Press `/` to enter search mode
2. Type "nginx" - should filter to nginx-related resources
3. Clear filter and try "redis"
4. Try filtering by status: "running", "pending", "failed"
5. Try filtering by namespace: "k8s-tui-test"
6. Press `Esc` to clear filter

**Expected**:
- Only matching resources displayed
- Search is case-insensitive
- Search works across name, namespace, status fields
- Clearing filter shows all resources again

### 9. Test Navigation & Scrolling
**Goal**: Verify list navigation controls

1. Use `↑`/`↓` arrow keys to move through list
2. Use `Page Up`/`Page Down` for page scrolling
3. Use `Home` to jump to top
4. Use `End` to jump to bottom
5. Try navigation with different resource types

**Expected**:
- Selection indicator moves correctly
- Viewport scrolls when selection moves off-screen
- Boundary conditions work (can't scroll past top/bottom)

### 10. Test Auto-Refresh
**Goal**: Verify 5-second auto-refresh functionality

1. Open k8s-tui in one terminal
2. In another terminal, scale a deployment:
   ```bash
   kubectl scale deployment nginx-healthy -n k8s-tui-test --replicas=5
   ```
3. Watch the TUI update automatically within 5 seconds
4. Scale back down:
   ```bash
   kubectl scale deployment nginx-healthy -n k8s-tui-test --replicas=3
   ```

**Expected**:
- TUI updates automatically without manual refresh
- Changes appear within ~5 seconds
- Selection position is maintained where possible

### 11. Test Multi-Container Pods
**Goal**: Verify multi-container pod display

1. Navigate to Pods tab
2. Select `multi-container-pod`
3. Press `Enter` for detail view
4. Verify all 3 containers are shown
5. Check ready status shows "3/3"

**Expected**:
- Ready count is "3/3"
- Detail view lists all containers
- Container names and images are visible

### 12. Test Error States
**Goal**: Verify error handling and display

1. Navigate to Pods tab
2. Find pod from `failing-deployment` (should be in ImagePullBackOff)
3. Observe status symbol (⚠ or ✗)
4. Press `Enter` to view details
5. Check if error reason is visible

**Expected**:
- Error status is clearly indicated
- Status symbol reflects error state
- Detail view provides useful information

### 13. Test Resource Counts
**Goal**: Verify correct resource counts in header

1. Switch between tabs and observe header
2. Compare counts with kubectl:
   ```bash
   kubectl get pods -n k8s-tui-test
   kubectl get services -n k8s-tui-test
   kubectl get deployments -n k8s-tui-test
   kubectl get statefulsets -n k8s-tui-test
   ```

**Expected**:
- Counts match kubectl output
- Header updates when switching tabs

## Expected Resource Counts

When all resources are successfully created:

| Resource Type | Expected Count (k8s-tui-test namespace) |
|---------------|----------------------------------------|
| Pods | ~13-15 (3 nginx, 1 redis, 3 postgres, 2 redis-cluster, 3 standalone, 1 failing) |
| Services | 6 |
| Deployments | 4 |
| StatefulSets | 2 |

**Note**: Pod count may vary as some pods complete or fail.

## Troubleshooting

### Pods Stuck in Pending
- **Cause**: Insufficient cluster resources or storage provisioner issues
- **Solution**: Check node resources or reduce replica counts

### StatefulSets Not Ready
- **Cause**: No default StorageClass or PVC provisioning failure
- **Solution**: Check if your cluster has a default storage class:
  ```bash
  kubectl get storageclass
  ```

### Failing Deployment Always in Error
- **Expected behavior**: This deployment intentionally uses a bad image to test error states

### Services Show No Endpoints
- **Check**: Verify matching pods exist with correct labels
  ```bash
  kubectl get endpoints -n k8s-tui-test
  ```

## Advanced Testing

### Test with Different Namespaces

```bash
# Create resources in default namespace too
kubectl apply -f test-resources.yaml -n default

# Run TUI without namespace flag to see all namespaces
./k8s-tui
```

### Test with kubectl Port-Forward

```bash
# Test that TUI shows correct status while port-forwarding
kubectl port-forward -n k8s-tui-test service/nginx-service 8080:80

# Access in browser: http://localhost:8080
```

### Stress Test with Many Resources

```bash
# Scale up to create more pods
kubectl scale deployment nginx-healthy -n k8s-tui-test --replicas=10

# Test TUI performance with more resources
```

## Useful kubectl Commands

```bash
# Watch resources being created
kubectl get all -n k8s-tui-test -w

# Check pod logs (useful for debugging)
kubectl logs -n k8s-tui-test <pod-name>

# Describe a resource
kubectl describe deployment nginx-healthy -n k8s-tui-test

# Force delete stuck resources
kubectl delete namespace k8s-tui-test --grace-period=0 --force
```

## Feedback & Issues

When testing, note:
- Any UI rendering issues
- Incorrect status symbols
- Navigation problems
- Performance with many resources
- Any crashes or errors

Report findings in the project issue tracker!
