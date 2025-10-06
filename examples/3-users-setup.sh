#!/bin/bash
# Setup for exactly 3 users with API key authentication

set -e

NAMESPACE="bedrock-system"

echo "🔐 Setting up Bedrock Proxy for 3 users"
echo ""

# Generate API keys for 3 users
USER1_KEY=$(openssl rand -hex 32)
USER2_KEY=$(openssl rand -hex 32)
USER3_KEY=$(openssl rand -hex 32)

# Create Kubernetes secret with the 3 API keys
kubectl create secret generic bedrock-api-keys \
  --from-literal=API_KEY_ALICE=$USER1_KEY \
  --from-literal=API_KEY_BOB=$USER2_KEY \
  --from-literal=API_KEY_CHARLIE=$USER3_KEY \
  -n $NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# Enable authentication on the deployment
kubectl set env deployment/bedrock-proxy \
  AUTH_ENABLED=true \
  AUTH_MODE=api_key \
  -n $NAMESPACE

echo ""
echo "✅ Setup complete! Share these API keys securely:"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "👤 User: Alice"
echo "🔑 API Key: $USER1_KEY"
echo ""
echo "👤 User: Bob"
echo "🔑 API Key: $USER2_KEY"
echo ""
echo "👤 User: Charlie"
echo "🔑 API Key: $USER3_KEY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Save to file (IMPORTANT: Share securely and delete this file!)
cat > api-keys-3-users.txt <<EOF
Bedrock Proxy - API Keys
Generated: $(date)

User: Alice
API Key: $USER1_KEY

User: Bob
API Key: $USER2_KEY

User: Charlie
API Key: $USER3_KEY

Usage (curl):
curl -H "X-API-Key: $USER1_KEY" https://your-ingress-url/health

Usage (Python):
import requests
headers = {"X-API-Key": "$USER1_KEY"}
response = requests.get("https://your-ingress-url/health", headers=headers)

Usage (JavaScript):
const headers = { "X-API-Key": "$USER1_KEY" };
fetch("https://your-ingress-url/health", { headers });
EOF

echo "📝 API keys saved to: api-keys-3-users.txt"
echo "⚠️  Share this file securely with your users, then DELETE it!"
echo ""
echo "🧪 Test access:"
echo "curl -H \"X-API-Key: $USER1_KEY\" https://your-ingress-url/health"
