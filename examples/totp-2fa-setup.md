# TOTP/2FA Setup with Google Authenticator

Complete guide to enable Time-based One-Time Password (TOTP) authentication for your Bedrock proxy.

## 📱 What is TOTP/2FA?

**TOTP (Time-based One-Time Password)** generates 6-digit codes that change every 30 seconds:
- Works with Google Authenticator, Authy, 1Password, etc.
- Adds second layer of security beyond API keys
- No internet required (works offline)
- Industry standard (used by Google, GitHub, AWS, etc.)

## 🎯 Two Options

### Option 1: API Key + TOTP (Recommended)
```
User provides:
1. API Key (X-API-Key: bdrk_abc123...)
2. TOTP Code (X-TOTP-Code: 123456)

Both required for access
```

### Option 2: TOTP Only
```
User provides:
1. TOTP Code (X-TOTP-Code: 123456)

No API key needed
```

---

## 🚀 Setup: Enable TOTP for 3 Users

### Step 1: Enable 2FA on Proxy

```bash
# Enable 2FA requirement
kubectl set env deployment/bedrock-proxy \
  AUTH_ENABLED=true \
  AUTH_MODE=api_key \
  REQUIRE_2FA=true \
  -n bedrock-system

# Restart deployment
kubectl rollout restart deployment/bedrock-proxy -n bedrock-system
```

### Step 2: Generate TOTP Secrets for Each User

Create a script to generate TOTP for each user:

```bash
#!/bin/bash
# generate-totp-for-users.sh

# This script generates TOTP secrets and QR codes for 3 users

cat > setup-2fa.go <<'EOF'
package main

import (
    "fmt"
    "github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
)

func main() {
    users := []struct{
        name  string
        email string
    }{
        {"Alice", "alice@example.com"},
        {"Bob", "bob@example.com"},
        {"Charlie", "charlie@example.com"},
    }

    for _, user := range users {
        // Generate TOTP key
        key, _ := totp.Generate(totp.GenerateOpts{
            Issuer:      "Bedrock Proxy",
            AccountName: user.email,
        })

        // Generate QR code
        qrcode.WriteFile(key.String(), qrcode.Medium, 256,
            fmt.Sprintf("qr-%s.png", user.name))

        fmt.Printf("\n=== %s (%s) ===\n", user.name, user.email)
        fmt.Printf("Secret: %s\n", key.Secret())
        fmt.Printf("QR Code: qr-%s.png\n", user.name)
        fmt.Printf("URL: %s\n", key.String())
    }
}
EOF

# Run it
go run setup-2fa.go

# Output will be:
# === Alice (alice@example.com) ===
# Secret: JBSWY3DPEHPK3PXP
# QR Code: qr-Alice.png
# URL: otpauth://totp/Bedrock%20Proxy:alice@example.com?secret=...
```

### Step 3: Share QR Codes with Users

```bash
# QR codes are saved as images
ls -la qr-*.png

# Share with users via:
# - Slack DM
# - Email (encrypted)
# - Secure file transfer
# - In-person display

# IMPORTANT: Delete QR codes after sharing!
shred -u qr-*.png
```

---

## 📱 User Setup (Alice's Example)

### Step 1: Install Authenticator App

Choose one:
- **Google Authenticator** (iOS/Android) - Free
- **Authy** (iOS/Android/Desktop) - Free, cloud backup
- **1Password** (Paid, but integrates password manager)
- **Microsoft Authenticator** (iOS/Android) - Free

### Step 2: Scan QR Code

1. Open authenticator app
2. Click "Add account" or "+"
3. Choose "Scan QR code"
4. Scan the QR code image sent by admin

**Or manually enter:**
1. Choose "Enter setup key"
2. Account name: `Bedrock Proxy (alice@example.com)`
3. Secret key: `JBSWY3DPEHPK3PXP`
4. Type: Time-based
5. Save

### Step 3: Save Backup Codes

Admin will provide 10 backup codes:
```
ABCD-1234
EFGH-5678
IJKL-9012
...
```

**Store securely** (password manager, encrypted file)
Use if you lose your phone!

---

## 🧪 Testing TOTP

### Test from Alice's Laptop

```bash
# 1. Set API key (as before)
export BEDROCK_API_KEY='bdrk_alice_key'

# 2. Get TOTP code from Google Authenticator app
# (e.g., 123456)

# 3. Make request with both API key and TOTP
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: 123456" \
     https://bedrock-proxy.example.com/health

# Expected: {"status":"healthy"}
```

### Test with Invalid TOTP (Should Fail)

```bash
# Old/invalid code
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: 000000" \
     https://bedrock-proxy.example.com/health

# Expected: {"error":"Invalid TOTP code"}
```

### Test with Backup Code

```bash
# Use one of the backup codes
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: ABCD-1234" \
     https://bedrock-proxy.example.com/health

# ✅ Works once, then code is consumed
# ❌ Using same backup code again will fail
```

---

## 💻 Client Code Examples

### Python with TOTP

```python
import os
import pyotp
import requests

# Load API key
API_KEY = os.getenv('BEDROCK_API_KEY')

# Load TOTP secret (saved during setup)
TOTP_SECRET = os.getenv('TOTP_SECRET')  # e.g., 'JBSWY3DPEHPK3PXP'

# Generate current TOTP code
totp = pyotp.TOTP(TOTP_SECRET)
totp_code = totp.now()  # e.g., '123456'

print(f"Current TOTP code: {totp_code}")

# Make request with both API key and TOTP
headers = {
    'X-API-Key': API_KEY,
    'X-TOTP-Code': totp_code,
    'Content-Type': 'application/json'
}

response = requests.post(
    'https://bedrock-proxy.example.com/model/claude-3-sonnet/invoke',
    headers=headers,
    json={
        "messages": [{"role": "user", "content": "Hello!"}],
        "max_tokens": 100
    }
)

print(response.json())
```

### JavaScript/Node.js with TOTP

```javascript
const speakeasy = require('speakeasy');
const fetch = require('node-fetch');

const API_KEY = process.env.BEDROCK_API_KEY;
const TOTP_SECRET = process.env.TOTP_SECRET;

// Generate TOTP code
const totpCode = speakeasy.totp({
  secret: TOTP_SECRET,
  encoding: 'base32'
});

console.log(`Current TOTP code: ${totpCode}`);

// Make request
fetch('https://bedrock-proxy.example.com/health', {
  headers: {
    'X-API-Key': API_KEY,
    'X-TOTP-Code': totpCode
  }
})
  .then(res => res.json())
  .then(data => console.log(data))
  .catch(err => console.error(err));
```

### Go with TOTP

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/pquerna/otp/totp"
)

func main() {
    apiKey := os.Getenv("BEDROCK_API_KEY")
    totpSecret := os.Getenv("TOTP_SECRET")

    // Generate TOTP code
    code, _ := totp.GenerateCode(totpSecret, time.Now())
    fmt.Printf("Current TOTP code: %s\n", code)

    // Make request
    req, _ := http.NewRequest("GET",
        "https://bedrock-proxy.example.com/health", nil)

    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("X-TOTP-Code", code)

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()

    // Handle response
    fmt.Println("Status:", resp.Status)
}
```

### Bash Script with Auto-Generated TOTP

```bash
#!/bin/bash
# bedrock-request.sh - Makes authenticated request with auto-generated TOTP

BEDROCK_API_KEY="bdrk_your_key_here"
TOTP_SECRET="JBSWY3DPEHPK3PXP"

# Generate TOTP code (requires oathtool)
# Install: brew install oath-toolkit (macOS) or apt-get install oathtool (Linux)
TOTP_CODE=$(oathtool --totp -b "$TOTP_SECRET")

echo "Using TOTP code: $TOTP_CODE"

# Make request
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: $TOTP_CODE" \
     https://bedrock-proxy.example.com/health
```

---

## 🔧 Advanced Configuration

### Per-User 2FA Control

Enable 2FA only for specific users:

```sql
-- Enable 2FA for Alice
UPDATE api_key_2fa
SET is_enabled = 1
WHERE api_key_id = (SELECT id FROM api_keys WHERE name = 'Alice');

-- Disable 2FA for Bob
UPDATE api_key_2fa
SET is_enabled = 0
WHERE api_key_id = (SELECT id FROM api_keys WHERE name = 'Bob');
```

### Custom TOTP Settings

Modify TOTP parameters:

```go
// internal/auth/totp.go - Customize these values
key, err := totp.Generate(totp.GenerateOpts{
    Issuer:      "Bedrock Proxy",
    AccountName: accountName,
    Period:      30,              // Code valid for 30 seconds (default)
    Digits:      otp.DigitsSix,   // 6-digit codes (can use DigitsEight)
    Algorithm:   otp.AlgorithmSHA1, // SHA1 (or SHA256/SHA512)
})
```

### Backup Code Management

```bash
# Regenerate backup codes for Alice
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  /app/admin regenerate-backup-codes --user alice

# List remaining backup codes (admin only)
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  /app/admin list-backup-codes --user alice

# Output: 7 backup codes remaining
```

---

## 🔍 Monitoring & Audit

### View 2FA Usage

```bash
# Check 2FA authentication logs
kubectl logs -n bedrock-system deployment/bedrock-proxy | grep "2fa"

# Output examples:
# {"user":"Alice","action":"2fa_success","ip":"203.0.113.5"}
# {"user":"Bob","action":"2fa_failed","ip":"198.51.100.23","error":"invalid_code"}
# {"user":"Charlie","action":"backup_code_used","code":"ABCD-1234"}
```

### 2FA Statistics

```bash
# Query audit database
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  sqlite3 /data/apikeys.db "
    SELECT
      name,
      COUNT(*) as attempts,
      SUM(CASE WHEN action = '2fa_success' THEN 1 ELSE 0 END) as success,
      SUM(CASE WHEN action = '2fa_failed' THEN 1 ELSE 0 END) as failed
    FROM api_key_audit a
    JOIN api_keys k ON a.api_key_id = k.id
    WHERE action LIKE '2fa%'
    GROUP BY name;
  "

# Output:
# Alice|150|148|2
# Bob|89|89|0
# Charlie|120|115|5
```

---

## 🚨 Troubleshooting

### "Invalid TOTP code" Error

**Possible causes:**

1. **Time drift** - Phone/server clocks not synced
   ```bash
   # Check server time
   kubectl exec -n bedrock-system deployment/bedrock-proxy -- date

   # Sync phone time (Settings > Date & Time > Automatic)
   ```

2. **Wrong secret** - Using incorrect TOTP secret
   ```bash
   # Verify secret in database
   kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
     sqlite3 /data/apikeys.db "SELECT totp_secret FROM api_key_2fa WHERE api_key_id = 1;"
   ```

3. **Code already used** - TOTP codes are single-use within 30s window
   ```bash
   # Wait for next code (max 30 seconds)
   ```

### Lost Phone (Can't Generate TOTP)

Use backup codes:

```bash
# Use any remaining backup code
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: EFGH-5678" \
     https://bedrock-proxy.example.com/health

# Then generate new QR code
```

### Reset 2FA for User

```bash
# Disable 2FA for user (emergency access)
kubectl exec -n bedrock-system deployment/bedrock-proxy -- \
  sqlite3 /data/apikeys.db "
    UPDATE api_key_2fa
    SET is_enabled = 0
    WHERE api_key_id = (SELECT id FROM api_keys WHERE name = 'Alice');
  "

# User can now access with API key only
# Then re-enable with new QR code
```

---

## 📊 Security Benefits

| Feature | Benefit |
|---------|---------|
| **Time-based** | Code changes every 30s |
| **Offline** | Works without internet |
| **Device-bound** | Tied to specific device |
| **Phishing-resistant** | Code can't be reused |
| **Backup codes** | Recovery if device lost |
| **Audit trail** | All 2FA attempts logged |

---

## 📋 Summary

**What You Need:**
1. ✅ Enable 2FA on proxy (`REQUIRE_2FA=true`)
2. ✅ Generate TOTP secrets for users
3. ✅ Share QR codes securely
4. ✅ Users scan with authenticator app

**What Users Do:**
1. Save API key (as before)
2. Scan QR code with Google Authenticator
3. Include TOTP code in every request:
   ```bash
   curl -H "X-API-Key: $KEY" \
        -H "X-TOTP-Code: 123456" \
        https://...
   ```

**Authentication Flow:**
```
User provides: API Key + TOTP Code (6 digits)
         ↓
Proxy validates:
  1. API key in database ✓
  2. TOTP code matches secret ✓
  3. Code not already used ✓
         ↓
Access granted!
```

**Time-based security with zero server-side secrets!** 🔐
