# Authorization & Access Control Guide

This guide explains how to secure your Bedrock proxy with multi-layer authorization.

## 🔐 Security Architecture

```
┌─────────────────────────────────────────────────┐
│  Layer 1: Network (VPC, Security Groups)        │
├─────────────────────────────────────────────────┤
│  Layer 2: Kubernetes (RBAC, Network Policies)   │
├─────────────────────────────────────────────────┤
│  Layer 3: Ingress/Gateway (OAuth, mTLS)         │
├─────────────────────────────────────────────────┤
│  Layer 4: Application (API Keys, JWT)           │
├─────────────────────────────────────────────────┤
│  Layer 5: AWS IAM (IRSA, Bedrock Access)        │
└─────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Option 1: API Key Authentication (Recommended for Most Use Cases)

**Step 1: Generate API keys**
```bash
# Generate secure API keys
openssl rand -hex 32  # For admin
openssl rand -hex 32  # For app1
openssl rand -hex 32  # For app2
```

**Step 2: Create Kubernetes secret**
```bash
kubectl create secret generic bedrock-api-keys \
  --from-literal=API_KEY_ADMIN=<admin-key> \
  --from-literal=API_KEY_APP1=<app1-key> \
  --from-literal=API_KEY_APP2=<app2-key> \
  -n bedrock-system
```

**Step 3: Enable auth in deployment**
```bash
kubectl set env deployment/bedrock-proxy \
  AUTH_ENABLED=true \
  AUTH_MODE=api_key \
  -n bedrock-system
```

**Step 4: Use the proxy**
```bash
# With X-API-Key header
curl -H "X-API-Key: <your-api-key>" \
  https://bedrock-proxy/model/anthropic.claude-3-sonnet/invoke

# With Bearer token
curl -H "Authorization: Bearer <your-api-key>" \
  https://bedrock-proxy/model/anthropic.claude-3-sonnet/invoke
```

---

## 📋 Authorization Methods

### 1. API Key Authentication

**Pros**: Simple, works with any client, per-user tracking
**Use case**: Internal services, trusted applications

```yaml
# deployment-with-auth.yaml
env:
- name: AUTH_ENABLED
  value: "true"
- name: AUTH_MODE
  value: "api_key"
envFrom:
- secretRef:
    name: bedrock-api-keys
```

**Client usage**:
```python
import requests

headers = {
    "X-API-Key": "your-api-key-here",
    "Content-Type": "application/json"
}

response = requests.post(
    "https://bedrock-proxy/model/anthropic.claude-3-sonnet/invoke",
    headers=headers,
    json={"prompt": "Hello"}
)
```

---

### 2. Basic Authentication

**Pros**: Built into HTTP, simple
**Use case**: Quick setup, testing

```bash
# Create credentials
kubectl create secret generic bedrock-basic-auth \
  --from-literal=credentials="admin:strong_password,user1:user1_pass" \
  -n bedrock-system

# Enable
kubectl set env deployment/bedrock-proxy \
  AUTH_MODE=basic \
  BASIC_AUTH_CREDENTIALS=admin:pass,user:pass \
  -n bedrock-system
```

**Client usage**:
```bash
curl -u admin:strong_password \
  https://bedrock-proxy/model/anthropic.claude-3-sonnet/invoke
```

---

### 3. Kubernetes Service Account (mTLS/Istio)

**Pros**: Zero config for clients, K8s native, automatic rotation
**Use case**: Service-to-service within cluster

```yaml
# rbac.yaml - Already created
---
# Client pod
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: my-app-namespace
  labels:
    bedrock-client: "true"
spec:
  serviceAccountName: my-app-sa  # Must be in allowed list
  containers:
  - name: app
    image: myapp:latest
```

**Network Policy** restricts access:
```yaml
# Only pods with label bedrock-client=true can access
spec:
  ingress:
  - from:
    - podSelector:
        matchLabels:
          bedrock-client: "true"
```

**Configure**:
```bash
kubectl label namespace my-app-namespace bedrock-access=allowed

kubectl set env deployment/bedrock-proxy \
  AUTH_MODE=service_account \
  ALLOWED_SERVICE_ACCOUNTS=my-app-namespace/my-app-sa,other-ns/other-sa \
  -n bedrock-system
```

---

### 4. AWS IAM (IRSA-based)

**Pros**: AWS native, fine-grained permissions
**Use case**: Cross-account access, AWS-native apps

```yaml
# Client pod with IRSA
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-with-bedrock-access
  namespace: my-namespace
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/bedrock-client-role

---
# IAM Trust Policy for client role
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::ACCOUNT:oidc-provider/oidc.eks.REGION.amazonaws.com/id/XXXXXX"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "oidc.eks.REGION.amazonaws.com/id/XXXXXX:sub":
          "system:serviceaccount:my-namespace:app-with-bedrock-access"
      }
    }
  }]
}
```

---

## 🌐 Advanced: OAuth2/OIDC with AWS Cognito

**Best for**: External users, web applications, SSO integration

### Setup AWS Cognito

```bash
# Create user pool
aws cognito-idp create-user-pool \
  --pool-name bedrock-users \
  --auto-verified-attributes email

# Create app client
aws cognito-idp create-user-pool-client \
  --user-pool-id us-east-1_XXXXX \
  --client-name bedrock-proxy \
  --generate-secret
```

### Use OAuth2 Proxy as Sidecar

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bedrock-proxy
spec:
  template:
    spec:
      containers:
      # Main proxy container
      - name: bedrock-proxy
        image: bedrock-proxy:latest
        ports:
        - containerPort: 8080

      # OAuth2 Proxy sidecar
      - name: oauth2-proxy
        image: quay.io/oauth2-proxy/oauth2-proxy:latest
        args:
        - --provider=oidc
        - --provider-display-name=AWS Cognito
        - --oidc-issuer-url=https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXX
        - --upstream=http://localhost:8080
        - --http-address=0.0.0.0:4180
        - --email-domain=*
        env:
        - name: OAUTH2_PROXY_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: oauth-credentials
              key: client-id
        - name: OAUTH2_PROXY_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: oauth-credentials
              key: client-secret
        ports:
        - containerPort: 4180
```

---

## 🔒 Production Best Practices

### 1. Use External Secrets Operator

Instead of plain Kubernetes secrets:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: bedrock-api-keys
  namespace: bedrock-system
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: bedrock-api-keys
  data:
  - secretKey: API_KEY_ADMIN
    remoteRef:
      key: bedrock/api-keys
      property: admin
  - secretKey: API_KEY_APP1
    remoteRef:
      key: bedrock/api-keys
      property: app1
```

### 2. Implement Rate Limiting

```bash
# Enable per-user rate limiting
kubectl set env deployment/bedrock-proxy \
  RATE_LIMIT_ENABLED=true \
  RATE_LIMIT_REQUESTS_PER_MINUTE=100 \
  -n bedrock-system
```

### 3. Audit Logging

All authenticated requests are logged with user information:

```json
{
  "timestamp": "2025-01-06T10:30:00Z",
  "user": "app1",
  "auth_method": "api_key",
  "request_id": "abc123",
  "path": "/model/claude-3-sonnet/invoke",
  "status": 200
}
```

### 4. Rotate API Keys Regularly

```bash
# Generate new key
NEW_KEY=$(openssl rand -hex 32)

# Update secret
kubectl patch secret bedrock-api-keys \
  -p "{\"data\":{\"API_KEY_APP1\":\"$(echo -n $NEW_KEY | base64)\"}}" \
  -n bedrock-system

# Notify client to update
echo "New API key for APP1: $NEW_KEY"
```

---

## 📊 Authorization Matrix

| Method | Complexity | Security | Use Case |
|--------|-----------|----------|----------|
| **API Key** | ⭐ Low | ⭐⭐⭐ Medium | Internal services, simple apps |
| **Basic Auth** | ⭐ Low | ⭐⭐ Low | Testing, quick demos |
| **Service Account** | ⭐⭐ Medium | ⭐⭐⭐⭐ High | K8s services, zero-config |
| **IAM (IRSA)** | ⭐⭐⭐ High | ⭐⭐⭐⭐⭐ Highest | AWS-native, cross-account |
| **OAuth2/OIDC** | ⭐⭐⭐⭐ Very High | ⭐⭐⭐⭐⭐ Highest | External users, SSO |

---

## 🧪 Testing Authorization

### Test API Key Auth
```bash
# Valid key
curl -H "X-API-Key: valid-key-here" http://localhost:8080/health
# Expected: 200 OK

# Invalid key
curl -H "X-API-Key: invalid-key" http://localhost:8080/health
# Expected: 401 Unauthorized

# Missing key
curl http://localhost:8080/health
# Expected: 401 Unauthorized (if auth enabled)
```

### Test Network Policy
```bash
# From allowed namespace
kubectl run test -n my-app-namespace --rm -it --image=curlimages/curl -- \
  curl bedrock-proxy-service.bedrock-system/health
# Expected: 200 OK

# From unauthorized namespace
kubectl run test -n default --rm -it --image=curlimages/curl -- \
  curl bedrock-proxy-service.bedrock-system/health
# Expected: Timeout (network policy blocks)
```

---

## 🚨 Troubleshooting

### Issue: 401 Unauthorized with valid API key

**Check:**
1. API key is correctly set in secret
2. Secret is mounted to pod
3. AUTH_ENABLED=true and AUTH_MODE=api_key
4. No typos in header name (X-API-Key)

```bash
# Verify secret
kubectl get secret bedrock-api-keys -n bedrock-system -o yaml

# Check pod env vars
kubectl exec -n bedrock-system deployment/bedrock-proxy -- env | grep API_KEY

# Check logs
kubectl logs -n bedrock-system deployment/bedrock-proxy | grep -i auth
```

### Issue: Network policy blocks legitimate traffic

**Check:**
1. Namespace has correct label: `bedrock-access=allowed`
2. Pod has correct label: `bedrock-client=true`
3. Network policy is applied to correct namespace

```bash
# Verify namespace label
kubectl get namespace my-app-namespace --show-labels

# Verify pod label
kubectl get pod -n my-app-namespace --show-labels

# Check network policy
kubectl describe networkpolicy bedrock-proxy-ingress -n bedrock-system
```

---

## 📚 Next Steps

1. ✅ Choose auth method based on your use case
2. ✅ Create and deploy secrets securely
3. ✅ Enable authentication in deployment
4. ✅ Test with authorized and unauthorized requests
5. ✅ Set up monitoring and alerts
6. ✅ Implement key rotation schedule

For questions or issues, check the [troubleshooting guide](./TROUBLESHOOTING.md) or open an issue.
