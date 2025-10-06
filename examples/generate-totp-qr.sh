#!/bin/bash
# Generate TOTP secrets and QR codes for 3 users
# Requires: qrencode (brew install qrencode or apt-get install qrencode)

set -e

NAMESPACE="${NAMESPACE:-bedrock-system}"
ISSUER="Bedrock Proxy"

echo "🔐 Generating TOTP/2FA for 3 users"
echo ""

# Check if qrencode is installed
if ! command -v qrencode &> /dev/null; then
    echo "❌ qrencode not found. Install it:"
    echo "   macOS:  brew install qrencode"
    echo "   Ubuntu: sudo apt-get install qrencode"
    exit 1
fi

# Generate TOTP for each user
generate_totp() {
    local name=$1
    local email=$2

    # Generate random secret (base32)
    secret=$(openssl rand -base32 20 | tr -d '=' | tr '[:lower:]' '[:upper:]')

    # Create OTP auth URL
    url="otpauth://totp/${ISSUER}:${email}?secret=${secret}&issuer=${ISSUER}"

    # Generate QR code (terminal)
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "👤 User: ${name} (${email})"
    echo "🔑 Secret: ${secret}"
    echo ""
    echo "📱 Scan this QR code with Google Authenticator:"
    echo ""
    qrencode -t ANSIUTF8 "${url}"
    echo ""
    echo "Or manually enter:"
    echo "  Account: ${ISSUER} (${email})"
    echo "  Secret:  ${secret}"
    echo "  Type:    Time-based"
    echo ""

    # Save QR code as PNG
    qrencode -o "qr-${name}.png" "${url}"
    echo "💾 QR code saved to: qr-${name}.png"

    # Generate backup codes
    echo "🔐 Backup codes (save securely!):"
    for i in {1..10}; do
        code=$(openssl rand -hex 4 | tr '[:lower:]' '[:upper:]')
        echo "   ${code:0:4}-${code:4:4}"
    done

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Save to database (if running in K8s)
    if kubectl get deployment bedrock-proxy -n $NAMESPACE &>/dev/null; then
        echo "📊 Saving to database..."

        # Get API key ID for this user
        KEY_ID=$(kubectl exec -n $NAMESPACE deployment/bedrock-proxy -- \
            sqlite3 /data/apikeys.db \
            "SELECT id FROM api_keys WHERE name='${name}' LIMIT 1;" 2>/dev/null || echo "")

        if [ -n "$KEY_ID" ]; then
            # Insert TOTP secret
            kubectl exec -n $NAMESPACE deployment/bedrock-proxy -- \
                sqlite3 /data/apikeys.db \
                "INSERT OR REPLACE INTO api_key_2fa (api_key_id, totp_secret, is_enabled)
                 VALUES (${KEY_ID}, '${secret}', 1);" 2>/dev/null || true
            echo "✅ TOTP saved for ${name}"
        else
            echo "⚠️  API key for ${name} not found. Create it first with 3-users-setup.sh"
        fi
        echo ""
    fi
}

# Generate for 3 users
generate_totp "Alice" "alice@example.com"
sleep 1

generate_totp "Bob" "bob@example.com"
sleep 1

generate_totp "Charlie" "charlie@example.com"

echo ""
echo "✅ TOTP setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Share QR codes with users (send PNG files securely)"
echo "2. Users scan with Google Authenticator app"
echo "3. Enable 2FA requirement:"
echo ""
echo "   kubectl set env deployment/bedrock-proxy REQUIRE_2FA=true -n $NAMESPACE"
echo ""
echo "4. Test with:"
echo ""
echo '   curl -H "X-API-Key: $KEY" -H "X-TOTP-Code: 123456" https://...'
echo ""
echo "⚠️  IMPORTANT: Delete QR code files after sharing!"
echo "   shred -u qr-*.png"
echo ""
