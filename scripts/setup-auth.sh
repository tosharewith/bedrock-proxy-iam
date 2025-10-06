#!/bin/bash
# Copyright 2025 Bedrock Proxy Authors
# SPDX-License-Identifier: Apache-2.0

# Setup authorization for Bedrock Proxy
set -e

NAMESPACE="${NAMESPACE:-bedrock-system}"
AUTH_MODE="${AUTH_MODE:-api_key}"

echo "🔐 Setting up authorization for Bedrock Proxy"
echo "   Namespace: $NAMESPACE"
echo "   Auth Mode: $AUTH_MODE"
echo ""

# Create namespace if it doesn't exist
if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    echo "📦 Creating namespace: $NAMESPACE"
    kubectl create namespace $NAMESPACE
    kubectl label namespace $NAMESPACE bedrock-access=allowed
fi

# Setup based on auth mode
case $AUTH_MODE in
    api_key)
        echo "🔑 Setting up API Key authentication"

        # Generate API keys
        ADMIN_KEY=$(openssl rand -hex 32)
        APP1_KEY=$(openssl rand -hex 32)
        APP2_KEY=$(openssl rand -hex 32)

        # Create secret
        kubectl create secret generic bedrock-api-keys \
            --from-literal=API_KEY_ADMIN=$ADMIN_KEY \
            --from-literal=API_KEY_APP1=$APP1_KEY \
            --from-literal=API_KEY_APP2=$APP2_KEY \
            -n $NAMESPACE \
            --dry-run=client -o yaml | kubectl apply -f -

        echo "✅ API keys created:"
        echo "   ADMIN: $ADMIN_KEY"
        echo "   APP1:  $APP1_KEY"
        echo "   APP2:  $APP2_KEY"
        echo ""
        echo "⚠️  Save these keys securely! They won't be shown again."
        echo ""

        # Save to file (remove in production!)
        cat > api-keys.txt <<EOF
Bedrock Proxy API Keys
Generated: $(date)

ADMIN: $ADMIN_KEY
APP1:  $APP1_KEY
APP2:  $APP2_KEY

Usage:
curl -H "X-API-Key: $ADMIN_KEY" https://bedrock-proxy/model/anthropic.claude-3-sonnet/invoke
EOF
        echo "📝 Keys saved to: api-keys.txt"
        ;;

    basic)
        echo "🔑 Setting up Basic authentication"

        # Generate passwords
        ADMIN_PASS=$(openssl rand -hex 16)
        USER1_PASS=$(openssl rand -hex 16)

        # Create secret
        kubectl create secret generic bedrock-basic-auth \
            --from-literal=credentials="admin:$ADMIN_PASS,user1:$USER1_PASS" \
            -n $NAMESPACE \
            --dry-run=client -o yaml | kubectl apply -f -

        echo "✅ Basic auth credentials created:"
        echo "   admin: $ADMIN_PASS"
        echo "   user1: $USER1_PASS"
        echo ""
        echo "Usage: curl -u admin:$ADMIN_PASS https://bedrock-proxy/..."
        ;;

    service_account)
        echo "🔑 Setting up Service Account authentication"

        # Apply RBAC
        kubectl apply -f deployments/kubernetes/rbac.yaml

        # Create example service account
        kubectl create serviceaccount bedrock-client-sa -n $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

        echo "✅ Service Account authentication configured"
        echo "   Allowed: $NAMESPACE/bedrock-client-sa"
        echo ""
        echo "To allow other namespaces:"
        echo "   kubectl label namespace <namespace> bedrock-access=allowed"
        ;;

    *)
        echo "❌ Unknown auth mode: $AUTH_MODE"
        echo "   Supported: api_key, basic, service_account"
        exit 1
        ;;
esac

# Apply auth config
echo "📋 Creating auth configuration..."
kubectl create configmap bedrock-auth-config \
    --from-literal=auth_mode=$AUTH_MODE \
    --from-literal=rate_limit_enabled=true \
    --from-literal=rate_limit_requests_per_minute=100 \
    --from-literal=allowed_service_accounts="$NAMESPACE/bedrock-client-sa" \
    -n $NAMESPACE \
    --dry-run=client -o yaml | kubectl apply -f -

# Update deployment to enable auth
echo "🚀 Enabling authentication in deployment..."
kubectl set env deployment/bedrock-proxy \
    AUTH_ENABLED=true \
    AUTH_MODE=$AUTH_MODE \
    RATE_LIMIT_ENABLED=true \
    RATE_LIMIT_RPM=100 \
    -n $NAMESPACE 2>/dev/null || echo "⚠️  Deployment not found. Deploy first with: kubectl apply -f deployments/kubernetes/"

echo ""
echo "✅ Authorization setup complete!"
echo ""
echo "Next steps:"
echo "1. Deploy the proxy: kubectl apply -f deployments/kubernetes/deployment-with-auth.yaml"
echo "2. Test authentication: curl -H \"X-API-Key: <your-key>\" http://<proxy-url>/health"
echo "3. Review logs: kubectl logs -n $NAMESPACE deployment/bedrock-proxy"
echo ""
echo "📚 Full documentation: docs/AUTHORIZATION.md"
