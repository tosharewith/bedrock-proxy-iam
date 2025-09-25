# Kubernetes Deployment with VPA & HPA

This directory contains Kubernetes manifests for deploying the Bedrock IAM Proxy with advanced autoscaling capabilities.

## Autoscaling Strategy

### Horizontal Pod Autoscaler (HPA)
- **Scales pods horizontally** based on CPU, memory, and custom metrics
- **Range**: 3-20 replicas
- **Triggers**:
  - CPU utilization > 70%
  - Memory utilization > 80%
  - Response time > 2 seconds
  - Request rate > 50 RPS per pod

### Vertical Pod Autoscaler (VPA)
- **Scales pod resources vertically** (CPU/memory requests and limits)
- **Mode**: Auto (automatically applies recommendations)
- **Range**: 100m-2000m CPU, 128Mi-4Gi memory
- **Updates**: Both requests and limits

## Deployment Order

```bash
# 1. Create namespace and basic resources
kubectl apply -f namespace.yaml
kubectl apply -f serviceaccount.yaml
kubectl apply -f configmap.yaml

# 2. Deploy the application
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 3. Configure autoscaling
kubectl apply -f hpa.yaml
kubectl apply -f vpa.yaml
kubectl apply -f pdb.yaml

# 4. Set up monitoring
kubectl apply -f servicemonitor.yaml
```

## Monitoring and Metrics

### Required Metrics for HPA
- `bedrock_proxy_request_duration_seconds` - Response time
- `bedrock_proxy_requests_per_second` - Request rate
- Standard CPU/memory metrics

### Prometheus Configuration
```yaml
# Example Prometheus rule for custom metrics
groups:
- name: bedrock-proxy-hpa
  rules:
  - record: bedrock_proxy_requests_per_second
    expr: rate(bedrock_proxy_requests_total[1m])
```

## Prerequisites

### 1. Metrics Server
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### 2. VPA Components
```bash
kubectl apply -f https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/vpa-rbac.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/recommender-deployment.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/updater-deployment.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/admission-controller-deployment.yaml
```

### 3. Prometheus Operator (for custom metrics)
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

### 4. Prometheus Adapter (for HPA custom metrics)
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus-adapter prometheus-community/prometheus-adapter \
  --set prometheus.url=http://prometheus-server.monitoring.svc \
  --set prometheus.port=80
```

## Verification

### Check HPA Status
```bash
kubectl get hpa -n bedrock-system
kubectl describe hpa bedrock-proxy-hpa -n bedrock-system
```

### Check VPA Status
```bash
kubectl get vpa -n bedrock-system
kubectl describe vpa bedrock-proxy-vpa -n bedrock-system
```

### Monitor Scaling Events
```bash
kubectl get events -n bedrock-system --sort-by='.lastTimestamp'
```

### Check Resource Usage
```bash
kubectl top pods -n bedrock-system
kubectl top nodes
```

## Troubleshooting

### HPA Not Scaling
1. Check metrics server: `kubectl get apiservice v1beta1.metrics.k8s.io`
2. Verify custom metrics: `kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1`
3. Check HPA metrics: `kubectl describe hpa bedrock-proxy-hpa -n bedrock-system`

### VPA Not Working
1. Check VPA components: `kubectl get pods -n kube-system | grep vpa`
2. Verify VPA CRDs: `kubectl get crd | grep autoscaling`
3. Check VPA recommendations: `kubectl describe vpa bedrock-proxy-vpa -n bedrock-system`

### Pod Disruption Issues
1. Check PDB status: `kubectl get pdb -n bedrock-system`
2. Verify disruption events: `kubectl get events -n bedrock-system`

## Configuration Options

### HPA Tuning
- Adjust `minReplicas`/`maxReplicas` based on expected load
- Modify target utilization percentages
- Add/remove custom metrics as needed
- Tune scaling behavior policies

### VPA Tuning
- Change `updateMode` to "Off" or "Initial" for safer updates
- Adjust `minAllowed`/`maxAllowed` resource bounds
- Modify `controlledValues` to update only requests or limits

### PDB Configuration
- Use `minAvailable` for absolute numbers
- Use `maxUnavailable` for percentage-based limits
- Create multiple PDBs for different disruption scenarios

## Production Recommendations

1. **Start Conservative**: Begin with HPA only, add VPA gradually
2. **Monitor Closely**: Watch scaling behavior for the first week
3. **Set Resource Limits**: Always define reasonable upper bounds
4. **Test Disruptions**: Verify PDBs work during maintenance
5. **Use Node Affinity**: Spread pods across availability zones
6. **Monitor Costs**: Track resource usage and scaling costs

## Example Scaling Scenarios

### High Load (Black Friday)
- HPA scales to 20 pods based on high RPS
- VPA increases CPU/memory per pod based on actual usage
- PDB ensures minimum availability during any disruptions

### Low Load (Weekend)
- HPA scales down to 3 pods based on low metrics
- VPA reduces resource requests to save costs
- PDB still maintains minimum availability

### Memory Leak Detected
- VPA detects increasing memory usage pattern
- VPA increases memory limits to prevent OOM kills
- Alerts should still fire for investigation