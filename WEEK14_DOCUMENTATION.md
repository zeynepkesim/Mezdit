# Week 14 - Production Documentation (Frontend)

## Production Deployment

### Staging vs Production Differences

#### Staging Environment (`emzops-staging`)
- **Replicas**: 1
- **Resources**: Lower limits (100m CPU, 256Mi memory requests)
- **Autoscaling**: Disabled
- **Affinity**: Not configured

#### Production Environment (`emzops-prod`)
- **Replicas**: 2 (minimum for high availability)
- **Resources**: Higher limits (200m CPU, 512Mi memory requests; 500m CPU, 1Gi memory limits)
- **Autoscaling**: Enabled (2-5 replicas)
- **Affinity**: Pod anti-affinity configured

### Deployment Configuration

**Production uses `values-prod.yaml` for overrides:**
```bash
helm upgrade --install devops-frontend ./chart \
  --namespace emzops-prod \
  --values ./chart/values-prod.yaml \
  --set image.tag=$IMAGE_TAG
```

## Automatic Scaling

### HPA Configuration

**Metrics:**
- CPU Utilization: 30% target (low threshold for easy demo)
- Memory Utilization: 50% target (low threshold for easy demo)

**Scaling Range:** 2-5 replicas

**Why Frontend is Scaled:**
- Serves static files (can handle high traffic)
- Stateless (perfect for horizontal scaling)
- Benefits from multiple replicas during traffic spikes

## Affinity and Topology Constraints

### Pod Anti-Affinity

Pods are configured to prefer different nodes:
```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - devops-frontend
        topologyKey: kubernetes.io/hostname
```

**Result:** Single node failure does not take frontend offline.

## Summary

✅ **Production Deployment**: Configured with production values  
✅ **Automatic Scaling**: HPA enabled  
✅ **Affinity Constraints**: Pod anti-affinity configured  
✅ **Chaos Ready**: Zero downtime configuration

