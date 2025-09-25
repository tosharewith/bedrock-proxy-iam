# Bedrock Proxy Kustomize Deployments

This directory contains Kustomize configurations for deploying the Bedrock IAM Proxy across different environments.

## Structure

```
kustomize/
├── base/                    # Base Kubernetes manifests
├── overlays/
│   ├── development/         # Development environment
│   ├── staging/             # Staging environment
│   └── production/          # Production environment with Vault
└── README.md
```

## Prerequisites

- kustomize v4.0+
- kubectl
- AWS EKS cluster with IRSA configured
- Vault (for production environment)

## Quick Start

### Development Environment
```bash
# Apply development configuration
kubectl apply -k deployments/kustomize/overlays/development

# Check deployment status
kubectl get all -n bedrock-development
```

### Staging Environment
```bash
# Apply staging configuration
kubectl apply -k deployments/kustomize/overlays/staging

# Monitor rollout
kubectl rollout status deployment/staging-bedrock-proxy -n bedrock-staging
```

### Production Environment
```bash
# Apply production configuration (with Vault)
kubectl apply -k deployments/kustomize/overlays/production

# Verify deployment
kubectl get all -n bedrock-system
kubectl get hpa,vpa,pdb -n bedrock-system
```

## Environment Configurations

### Development
- **Namespace**: `bedrock-development`
- **Replicas**: 1
- **Resources**: Minimal (100m CPU, 128Mi RAM)
- **Mode**: Debug logging, NodePort service
- **Scaling**: HPA disabled, VPA disabled
- **Image**: `dev-latest`

```bash
# Build and view development manifest
kustomize build deployments/kustomize/overlays/development

# Apply with dry-run
kubectl apply -k deployments/kustomize/overlays/development --dry-run=client
```

### Staging
- **Namespace**: `bedrock-staging`
- **Replicas**: 2-10 (HPA managed)
- **Resources**: Moderate (150m CPU, 192Mi RAM)
- **Mode**: Info logging, internal LoadBalancer
- **Scaling**: HPA enabled, VPA in Initial mode
- **Image**: `staging-latest`

```bash
# Deploy to staging
kubectl apply -k deployments/kustomize/overlays/staging

# Monitor scaling
kubectl get hpa staging-bedrock-proxy-hpa -n bedrock-staging -w
```

### Production
- **Namespace**: `bedrock-system`
- **Replicas**: 5-50 (HPA managed)
- **Resources**: Optimized (300m CPU, 384Mi RAM)
- **Mode**: Warn logging, NLB with health checks
- **Scaling**: Full HPA/VPA with aggressive scaling
- **Security**: Vault integration, enhanced monitoring
- **Image**: Specific version tags

```bash
# Deploy to production
kubectl apply -k deployments/kustomize/overlays/production

# Verify Vault integration
kubectl get pods -n bedrock-system -o yaml | grep vault
```

## Customization

### Adding New Overlays

1. Create a new overlay directory:
```bash
mkdir -p deployments/kustomize/overlays/my-env
```

2. Create `kustomization.yaml`:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: bedrock-my-env
resources:
  - ../../base

namePrefix: myenv-

patches:
  # Your custom patches here
```

### Environment Variables

Use `configMapGenerator` to customize environment variables:

```yaml
configMapGenerator:
  - name: bedrock-proxy-config
    behavior: merge
    literals:
      - AWS_REGION=eu-west-1
      - LOG_LEVEL=debug
      - CUSTOM_VAR=value
```

### Resource Limits

Patch resource limits for different environments:

```yaml
patches:
  - target:
      kind: Deployment
      name: bedrock-proxy
    patch: |-
      - op: replace
        path: /spec/template/spec/containers/0/resources
        value:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
```

## Vault Integration (Production)

### Prerequisites
1. **Vault Server**: Running and accessible from the cluster
2. **Kubernetes Auth**: Vault Kubernetes auth method configured
3. **Secrets**: TLS certificates stored in Vault
4. **CSI Driver**: Secrets Store CSI Driver installed

### Vault Setup

1. Enable Kubernetes auth in Vault:
```bash
vault auth enable kubernetes

vault write auth/kubernetes/config \
    kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
    kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
```

2. Create Vault policy:
```bash
vault policy write bedrock-proxy-prod - <<EOF
path "secret/data/bedrock-proxy/tls" {
  capabilities = ["read"]
}
path "secret/data/bedrock-proxy/config" {
  capabilities = ["read"]
}
EOF
```

3. Create Vault role:
```bash
vault write auth/kubernetes/role/bedrock-proxy-prod \
    bound_service_account_names=bedrock-proxy \
    bound_service_account_namespaces=bedrock-system \
    policies=bedrock-proxy-prod \
    ttl=24h
```

4. Store secrets in Vault:
```bash
# TLS certificates
vault kv put secret/bedrock-proxy/tls \
    cert=@tls.crt \
    key=@tls.key

# Application config (optional)
vault kv put secret/bedrock-proxy/config \
    api_key="secret-api-key" \
    db_password="secret-password"
```

### Vault Secret Management

The production overlay uses the Secrets Store CSI Driver to mount Vault secrets:

```yaml
# vault-secret.yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: bedrock-proxy-vault-secrets
spec:
  provider: vault
  parameters:
    vaultAddress: "https://vault.company.internal:8200"
    roleName: "bedrock-proxy-prod"
```

## Validation and Testing

### Kustomize Validation
```bash
# Validate base configuration
kustomize build deployments/kustomize/base

# Validate all overlays
for env in development staging production; do
  echo "Validating $env..."
  kustomize build deployments/kustomize/overlays/$env > /dev/null
done
```

### Kubernetes Validation
```bash
# Dry-run apply
kubectl apply -k deployments/kustomize/overlays/production --dry-run=server

# Validate against cluster
kubectl apply -k deployments/kustomize/overlays/production --validate=true --dry-run=client
```

### Security Scanning
```bash
# Scan manifests with kube-score
kustomize build deployments/kustomize/overlays/production | kube-score score -

# Scan with polaris
kustomize build deployments/kustomize/overlays/production | polaris validate --format=pretty
```

## Monitoring and Observability

### Prometheus Queries for Each Environment

```promql
# Development metrics
sum(rate(bedrock_proxy_requests_total{environment="development"}[5m]))

# Staging metrics
histogram_quantile(0.95,
  sum(rate(bedrock_proxy_request_duration_seconds_bucket{environment="staging"}[5m])) by (le)
)

# Production metrics
sum(rate(bedrock_proxy_requests_total{environment="production",status=~"5.."}[5m])) /
sum(rate(bedrock_proxy_requests_total{environment="production"}[5m]))
```

### Grafana Dashboard Variables
```json
{
  "environment": {
    "query": "label_values(bedrock_proxy_requests_total, environment)",
    "type": "query"
  }
}
```

## Migration Between Environments

### Promote from Development to Staging
```bash
# Get development image tag
DEV_IMAGE=$(kubectl get deployment dev-bedrock-proxy -n bedrock-development -o jsonpath='{.spec.template.spec.containers[0].image}')

# Update staging kustomization
sed -i "s/newTag:.*/newTag: \"$DEV_IMAGE\"/" deployments/kustomize/overlays/staging/kustomization.yaml

# Apply staging update
kubectl apply -k deployments/kustomize/overlays/staging
```

### Promote from Staging to Production
```bash
# Tag staging image for production
STAGING_IMAGE=$(kubectl get deployment staging-bedrock-proxy -n bedrock-staging -o jsonpath='{.spec.template.spec.containers[0].image}')

# Update production with specific version
sed -i "s/newTag:.*/newTag: \"v1.2.3\"/" deployments/kustomize/overlays/production/kustomization.yaml

# Apply production update
kubectl apply -k deployments/kustomize/overlays/production
```

## Troubleshooting

### Common Issues

1. **Namespace Not Found**
```bash
# Create namespace manually
kubectl create namespace bedrock-system

# Or use kustomize with namespace creation
kubectl apply -k deployments/kustomize/overlays/production --recursive
```

2. **Image Pull Errors**
```bash
# Check image name in overlay
kustomize build deployments/kustomize/overlays/production | grep image:

# Verify image exists
docker pull <image-name>
```

3. **Vault Integration Issues**
```bash
# Check Vault connectivity
kubectl exec -it deployment/bedrock-proxy -n bedrock-system -- \
  wget -qO- http://vault.company.internal:8200/v1/sys/health

# Check Vault agent logs
kubectl logs -f deployment/bedrock-proxy -c vault-agent -n bedrock-system
```

### Debug Commands

```bash
# View final manifests
kustomize build deployments/kustomize/overlays/production

# Compare environments
diff <(kustomize build deployments/kustomize/overlays/staging) \
     <(kustomize build deployments/kustomize/overlays/production)

# Validate patches
kustomize build deployments/kustomize/overlays/production | kubectl apply --dry-run=client -f -
```

## Best Practices

1. **Version Control**: Always commit kustomization changes
2. **Testing**: Test overlays in development first
3. **Secrets**: Never store secrets in kustomization files
4. **Images**: Use specific tags in production, not `latest`
5. **Resources**: Set appropriate resource requests and limits
6. **Monitoring**: Deploy with monitoring from the start
7. **Security**: Use security contexts and network policies
8. **Backup**: Backup configurations before major changes