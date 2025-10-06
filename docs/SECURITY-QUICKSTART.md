# 🔐 Security & Authorization Quick Start

Choose the authorization method that fits your use case:

## 🎯 Choose Your Method

| Method | Setup Time | Security Level | Best For |
|--------|-----------|---------------|----------|
| **API Keys** | 5 min | ⭐⭐⭐ | Internal services, most common |
| **Service Account** | 10 min | ⭐⭐⭐⭐ | K8s-native apps, zero client config |
| **Basic Auth** | 3 min | ⭐⭐ | Testing, quick demos |
| **OAuth2/OIDC** | 30 min | ⭐⭐⭐⭐⭐ | External users, SSO required |
| **AWS IAM** | 15 min | ⭐⭐⭐⭐⭐ | AWS-native, cross-account |

---

## ⚡ Quick Setup (API Keys - Recommended)

```bash
# 1. Run automated setup
./scripts/setup-auth.sh

# 2. Deploy with auth enabled
kubectl apply -f deployments/kubernetes/deployment-with-auth.yaml

# 3. Get your API key from output
API_KEY="<from-setup-output>"

# 4. Test it
curl -H "X-API-Key: $API_KEY" http://bedrock-proxy/health
```

**That's it!** Your proxy is now secured with API key authentication.

---

## 📋 Manual Setup Options

### Option 1: API Key (Most Popular)

**When to use**: Internal services, microservices, simple apps

```bash
# Generate keys
ADMIN_KEY=$(openssl rand -hex 32)
APP_KEY=$(openssl rand -hex 32)

# Create secret
kubectl create secret generic bedrock-api-keys \
  --from-literal=API_KEY_ADMIN=$ADMIN_KEY \
  --from-literal=API_KEY_APP=$APP_KEY \
  -n bedrock-system

# Enable auth
kubectl set env deployment/bedrock-proxy \
  AUTH_ENABLED=true \
  AUTH_MODE=api_key \
  -n bedrock-system

# Use it
curl -H "X-API-Key: $ADMIN_KEY" https://bedrock-proxy/model/claude-3-sonnet/invoke
```

---

### Option 2: Kubernetes Service Account (K8s Native)

**When to use**: Service-to-service within cluster, zero client config

```bash
# 1. Label your namespace
kubectl label namespace my-app bedrock-access=allowed

# 2. Label your pods
kubectl label pod my-app bedrock-client=true -n my-app

# 3. Apply RBAC
kubectl apply -f deployments/kubernetes/rbac.yaml

# 4. Enable auth
kubectl set env deployment/bedrock-proxy \
  AUTH_MODE=service_account \
  ALLOWED_SERVICE_ACCOUNTS=my-app/my-app-sa \
  -n bedrock-system
```

**Client pod automatically authenticated** - no keys needed!

---

### Option 3: Basic Auth (Simple)

**When to use**: Testing, demos, quick prototypes

```bash
# Create credentials
kubectl create secret generic bedrock-basic-auth \
  --from-literal=credentials="admin:strongpass123,user:userpass" \
  -n bedrock-system

# Enable auth
kubectl set env deployment/bedrock-proxy \
  AUTH_MODE=basic \
  BASIC_AUTH_CREDENTIALS=admin:pass,user:pass \
  -n bedrock-system

# Use it
curl -u admin:strongpass123 https://bedrock-proxy/model/claude-3-sonnet/invoke
```

---

### Option 4: AWS IAM + IRSA (AWS Native)

**When to use**: AWS-native apps, cross-account access, maximum security

```bash
# 1. Create IAM role for client
aws iam create-role --role-name bedrock-client-role \
  --assume-role-policy-document file://trust-policy.json

# 2. Attach to service account
kubectl annotate serviceaccount my-app-sa \
  eks.amazonaws.com/role-arn=arn:aws:iam::ACCOUNT:role/bedrock-client-role \
  -n my-app

# 3. Client pods automatically use IAM role
```

---

### Option 5: OAuth2 + AWS Cognito (Enterprise SSO)

**When to use**: External users, web apps, corporate SSO

```bash
# 1. Create Cognito User Pool
aws cognito-idp create-user-pool --pool-name bedrock-users

# 2. Deploy OAuth2 proxy sidecar
kubectl apply -f deployments/kubernetes/oauth2-proxy.yaml

# 3. Users authenticate via browser
# 4. Proxy validates JWT tokens
```

Full setup: See [docs/AUTHORIZATION.md](./AUTHORIZATION.md)

---

## 🛡️ Multi-Layer Security Stack

```
Internet/VPC
    ↓
┌─────────────────────────────┐
│ AWS Security Groups          │ ← Network-level filtering
├─────────────────────────────┤
│ Kubernetes Network Policies  │ ← Pod-to-pod restrictions
├─────────────────────────────┤
│ Application Auth Middleware  │ ← API keys, JWT, etc.
├─────────────────────────────┤
│ AWS IAM (IRSA)              │ ← Bedrock access control
└─────────────────────────────┘
         ↓
    AWS Bedrock
```

---

## 🔒 Production Checklist

- [ ] ✅ Enable authentication (`AUTH_ENABLED=true`)
- [ ] 🔑 Use strong API keys (32+ chars)
- [ ] 🔐 Store secrets in AWS Secrets Manager or Vault
- [ ] 🌐 Restrict network access (VPC, Security Groups)
- [ ] 📊 Enable audit logging
- [ ] 🔄 Set up key rotation schedule
- [ ] 🚦 Configure rate limiting
- [ ] 📈 Monitor authentication failures
- [ ] 🔔 Set up alerts for unauthorized access
- [ ] 📝 Document who has access

---

## 🧪 Testing Your Setup

### Test API Key Auth
```bash
# Should succeed
curl -H "X-API-Key: valid-key" http://bedrock-proxy/health
# Expected: 200 OK

# Should fail
curl -H "X-API-Key: invalid-key" http://bedrock-proxy/health
# Expected: 401 Unauthorized

# Should fail
curl http://bedrock-proxy/health
# Expected: 401 Unauthorized
```

### Test Network Policies
```bash
# From allowed namespace (should work)
kubectl run test -n my-app --rm -it --image=curlimages/curl -- \
  curl bedrock-proxy-service.bedrock-system/health

# From unauthorized namespace (should timeout)
kubectl run test -n default --rm -it --image=curlimages/curl -- \
  curl bedrock-proxy-service.bedrock-system/health
```

---

## 🚨 Troubleshooting

### "401 Unauthorized" with valid key

```bash
# Check secret exists
kubectl get secret bedrock-api-keys -n bedrock-system

# Verify environment variables
kubectl exec deployment/bedrock-proxy -n bedrock-system -- env | grep API_KEY

# Check logs
kubectl logs deployment/bedrock-proxy -n bedrock-system | grep -i auth
```

### Network policy blocking traffic

```bash
# Verify namespace label
kubectl get ns my-app --show-labels | grep bedrock-access

# Verify pod label
kubectl get pod -n my-app --show-labels | grep bedrock-client

# Check network policy rules
kubectl describe networkpolicy bedrock-proxy-ingress -n bedrock-system
```

### IAM permissions issues

```bash
# Verify service account annotation
kubectl get sa my-app-sa -n my-app -o yaml | grep eks.amazonaws.com/role-arn

# Check IAM role trust policy
aws iam get-role --role-name bedrock-client-role

# Test credentials
kubectl exec deployment/my-app -n my-app -- env | grep AWS
```

---

## 📚 Learn More

- **Full Documentation**: [docs/AUTHORIZATION.md](./AUTHORIZATION.md)
- **Configuration Examples**: `deployments/kubernetes/auth-*.yaml`
- **Middleware Code**: `internal/middleware/authorization.go`
- **Setup Script**: `scripts/setup-auth.sh`

---

## 🆘 Need Help?

1. Check the [troubleshooting guide](./AUTHORIZATION.md#troubleshooting)
2. Review logs: `kubectl logs -n bedrock-system deployment/bedrock-proxy`
3. Verify configuration: `kubectl describe deployment bedrock-proxy -n bedrock-system`
4. Open an issue with logs and configuration

---

## 🔐 Security Best Practices Summary

1. **Always enable authentication in production** (`AUTH_ENABLED=true`)
2. **Use external secret management** (AWS Secrets Manager, Vault)
3. **Implement network policies** to restrict pod access
4. **Enable rate limiting** to prevent abuse
5. **Rotate keys regularly** (every 90 days)
6. **Monitor and alert** on authentication failures
7. **Use HTTPS/TLS** for all traffic
8. **Audit all access** with comprehensive logging

Your Bedrock proxy is now secure! 🎉
