# 3 Users Setup Example

Complete guide to configure Bedrock Proxy for exactly 3 users with network access via ingress endpoint.

## 📋 Overview

This example shows how to:
1. **Configure the proxy** to allow only 3 specific users
2. **Set up ingress** for external/internal network access
3. **Configure client laptops** to use the proxy

---

## 🚀 Quick Setup (10 minutes)

### Step 1: Deploy the Proxy with Authentication

```bash
# 1. Run the automated setup script
chmod +x 3-users-setup.sh
./3-users-setup.sh

# This will:
# - Generate 3 unique API keys
# - Create Kubernetes secret
# - Enable authentication
# - Output API keys for Alice, Bob, and Charlie
```

### Step 2: Deploy Ingress

```bash
# Edit ingress file with your domain and certificate
vim ingress-3-users.yaml

# Replace these values:
# - bedrock-proxy.example.com → your actual domain
# - arn:aws:acm:...certificate/xxx → your ACM certificate ARN

# Apply ingress
kubectl apply -f ingress-3-users.yaml

# Get ingress endpoint
kubectl get ingress bedrock-proxy-ingress -n bedrock-system
```

### Step 3: Configure DNS

```bash
# Get ALB DNS name
ALB_DNS=$(kubectl get ingress bedrock-proxy-ingress -n bedrock-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "Create DNS record:"
echo "  Type: CNAME (or A if using Alias)"
echo "  Name: bedrock-proxy.example.com"
echo "  Value: $ALB_DNS"

# Or use external-dns (automatic)
kubectl annotate service bedrock-proxy-service \
  external-dns.alpha.kubernetes.io/hostname=bedrock-proxy.example.com \
  -n bedrock-system
```

### Step 4: Share API Keys with Users

```bash
# API keys are in api-keys-3-users.txt
cat api-keys-3-users.txt

# Share securely (Slack DM, 1Password, etc.)
# Each user gets their own unique key
```

---

## 💻 Client Setup (For Each User)

### Alice's Laptop

```bash
# 1. Save API key
mkdir -p ~/.bedrock-proxy
echo "export BEDROCK_API_KEY='bdrk_alice_key_here'" > ~/.bedrock-proxy/config
echo "source ~/.bedrock-proxy/config" >> ~/.bashrc
source ~/.bashrc

# 2. Test access
curl -H "X-API-Key: $BEDROCK_API_KEY" \
  https://bedrock-proxy.example.com/health

# Expected: {"status":"healthy","service":"bedrock-proxy"}
```

### Bob's Laptop

```bash
# Same process with Bob's API key
echo "export BEDROCK_API_KEY='bdrk_bob_key_here'" > ~/.bedrock-proxy/config
# ... rest same as Alice
```

### Charlie's Laptop

```bash
# Same process with Charlie's API key
echo "export BEDROCK_API_KEY='bdrk_charlie_key_here'" > ~/.bedrock-proxy/config
# ... rest same as Alice
```

---

## 🧪 Testing

### Test from Alice's Laptop

```bash
# Health check
curl -H "X-API-Key: $BEDROCK_API_KEY" \
  https://bedrock-proxy.example.com/health

# Invoke Claude
curl -X POST \
  https://bedrock-proxy.example.com/model/anthropic.claude-3-sonnet-20240229-v1:0/invoke \
  -H "X-API-Key: $BEDROCK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Test Invalid Key (Should Fail)

```bash
# This should return 401 Unauthorized
curl -H "X-API-Key: invalid_key" \
  https://bedrock-proxy.example.com/health

# Expected: {"error":"Invalid API key"}
```

### Test No Key (Should Fail)

```bash
# This should also return 401
curl https://bedrock-proxy.example.com/health

# Expected: {"error":"Missing API key"}
```

---

## 🔒 Security Options

### Option A: Network-Level Restriction (IP Whitelist)

Restrict to specific IP addresses:

```yaml
# In ingress-3-users.yaml
metadata:
  annotations:
    # AWS ALB
    alb.ingress.kubernetes.io/inbound-cidrs: "203.0.113.0/24,198.51.100.0/24"

    # Or NGINX
    nginx.ingress.kubernetes.io/whitelist-source-range: "203.0.113.0/24,198.51.100.0/24"
```

### Option B: VPN/Private Access Only

Make ingress internal (VPC-only):

```yaml
metadata:
  annotations:
    alb.ingress.kubernetes.io/scheme: internal  # instead of internet-facing
```

Then users connect via:
- Corporate VPN
- AWS VPN/Direct Connect
- Bastion/Jump host

### Option C: Add 2FA (Google Authenticator)

Enable TOTP for each user:

```bash
# Enable 2FA requirement
kubectl set env deployment/bedrock-proxy \
  REQUIRE_2FA=true \
  -n bedrock-system

# Each user sets up Google Authenticator
# They'll need both API key + TOTP code to access
```

---

## 📊 Monitoring Access

### View Audit Logs

```bash
# Check who's accessing the proxy
kubectl logs -n bedrock-system deployment/bedrock-proxy | grep "auth_success"

# Output shows:
# {"user":"Alice","action":"auth_success","ip":"203.0.113.5"}
# {"user":"Bob","action":"auth_success","ip":"198.51.100.23"}
```

### Check Active Keys

```bash
# Get list of all API keys and usage
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  sqlite3 /data/apikeys.db "SELECT name, email, last_used_at FROM api_keys;"
```

### Revoke Access (e.g., Charlie leaves)

```bash
# Get Charlie's key ID
KEY_ID=$(kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  sqlite3 /data/apikeys.db "SELECT id FROM api_keys WHERE name='Charlie';")

# Revoke the key
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  sqlite3 /data/apikeys.db "UPDATE api_keys SET is_active=0 WHERE id=$KEY_ID;"

# Charlie can no longer access
```

---

## 🔄 Adding More Users

To add a 4th user (David):

```bash
# Generate new key
DAVID_KEY=$(openssl rand -hex 32)

# Add to secret
kubectl patch secret bedrock-api-keys -n bedrock-system \
  -p "{\"data\":{\"API_KEY_DAVID\":\"$(echo -n bdrk_$DAVID_KEY | base64)\"}}"

# Share key with David
echo "David's API key: bdrk_$DAVID_KEY"
```

---

## 🛠️ Troubleshooting

### Issue: "401 Unauthorized" with valid key

```bash
# 1. Check if auth is enabled
kubectl get deployment bedrock-proxy -n bedrock-system -o yaml | grep AUTH_ENABLED

# 2. Verify key exists in secret
kubectl get secret bedrock-api-keys -n bedrock-system -o yaml

# 3. Check proxy logs
kubectl logs -n bedrock-system deployment/bedrock-proxy | tail -20
```

### Issue: Can't reach ingress endpoint

```bash
# 1. Check ingress status
kubectl get ingress -n bedrock-system
kubectl describe ingress bedrock-proxy-ingress -n bedrock-system

# 2. Test DNS
nslookup bedrock-proxy.example.com

# 3. Test direct connection
curl -v https://bedrock-proxy.example.com/health
```

### Issue: SSL/TLS errors

```bash
# Check certificate
kubectl get ingress bedrock-proxy-ingress -n bedrock-system -o yaml | grep certificate-arn

# Verify cert-manager (if using)
kubectl get certificate -n bedrock-system
```

---

## 📁 Files in This Example

- **`3-users-setup.sh`** - Automated setup script
- **`ingress-3-users.yaml`** - Ingress configuration
- **`client-laptop-config.md`** - Client setup guide
- **`README.md`** - This file

---

## 🎯 Summary

**What You Did:**
1. ✅ Generated 3 unique API keys
2. ✅ Configured proxy with authentication
3. ✅ Deployed ingress with HTTPS
4. ✅ Configured DNS
5. ✅ Set up client laptops

**What Users Need:**
- Their API key (from api-keys-3-users.txt)
- Ingress URL (bedrock-proxy.example.com)
- Client config (see client-laptop-config.md)

**Security:**
- ✅ Only 3 users have access (via unique API keys)
- ✅ HTTPS encrypted communication
- ✅ All requests are audited
- ✅ Keys can be revoked anytime
- ✅ Optional: IP whitelist, 2FA, VPN-only access

Your Bedrock proxy is now accessible to exactly 3 users! 🎉
