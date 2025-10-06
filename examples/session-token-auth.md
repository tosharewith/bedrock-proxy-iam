# Session Token Authentication - 12+ Hour Access

**Simple authentication: Login once with API Key + TOTP, get a long-lived token for 12+ hours**

---

## 🎯 The Problem

**Before (Annoying):**
```bash
# Every request needs TOTP code from phone
curl -H "X-API-Key: bdrk_abc123" \
     -H "X-TOTP-Code: 123456" \    # Look at phone every time!
     https://...
```

**After (Easy):**
```bash
# Login once → Get token → Use for 12 hours
curl -H "X-Session-Token: bdrk_sess_xyz..." \
     https://...    # No phone needed!
```

---

## 🚀 How It Works

```
Step 1: Login (once per 12h)
  Send: API Key + TOTP Code
  Get:  Session Token (valid 12h)

Step 2-∞: Use Session Token
  Send: Session Token
  No TOTP needed!
```

---

## 📋 Setup (Server Side)

### Enable Session Token Authentication

```bash
# Deploy with session support
kubectl set env deployment/bedrock-proxy \
  AUTH_MODE=session \
  SESSION_DURATION=12h \
  -n bedrock-system

# Or edit deployment:
env:
- name: AUTH_MODE
  value: "session"
- name: SESSION_DURATION
  value: "12h"  # or 24h, 7d, etc.
```

---

## 💻 Client Usage

### Step 1: Login (Once)

```bash
# Login with API key + TOTP to get session token
curl -X POST https://bedrock-proxy.example.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "bdrk_abc123...",
    "totp_code": "123456"
  }'

# Response:
{
  "session_token": "bdrk_sess_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
  "expires_at": "2025-01-07T18:30:00Z",
  "expires_in": 43200,
  "user": "Alice",
  "message": "Authenticated successfully. Use this token for 12 hours."
}
```

### Step 2: Save Token

```bash
# Save token to environment
export SESSION_TOKEN="bdrk_sess_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"

# Or save to file
echo "$SESSION_TOKEN" > ~/.bedrock-session
chmod 600 ~/.bedrock-session
```

### Step 3: Use Token (for next 12 hours)

```bash
# All requests just use the session token - no TOTP!
curl -H "X-Session-Token: $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health

curl -X POST \
  -H "X-Session-Token: $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  https://bedrock-proxy.example.com/model/claude-3-sonnet/invoke \
  -d '{"messages":[{"role":"user","content":"Hello!"}]}'

# Works for 12 hours without looking at phone!
```

### Alternative: Bearer Token

```bash
# Also works with Authorization header
curl -H "Authorization: Bearer $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health
```

---

## 🐍 Python Example

```python
import os
import requests
import pyotp
from datetime import datetime

# Configuration
API_KEY = os.getenv('BEDROCK_API_KEY')
TOTP_SECRET = os.getenv('TOTP_SECRET')
BASE_URL = 'https://bedrock-proxy.example.com'

class BedrockClient:
    def __init__(self):
        self.session_token = None
        self.expires_at = None

    def login(self):
        """Login once to get session token"""
        # Generate TOTP code
        totp = pyotp.TOTP(TOTP_SECRET)
        totp_code = totp.now()

        # Login
        response = requests.post(
            f'{BASE_URL}/auth/login',
            json={
                'api_key': API_KEY,
                'totp_code': totp_code
            }
        )

        if response.status_code == 200:
            data = response.json()
            self.session_token = data['session_token']
            self.expires_at = data['expires_at']
            print(f"✅ Logged in! Token valid until {self.expires_at}")
            return True
        else:
            print(f"❌ Login failed: {response.text}")
            return False

    def call_bedrock(self, message):
        """Make request with session token (no TOTP needed)"""
        if not self.session_token:
            print("⚠️  Not logged in. Calling login()...")
            self.login()

        headers = {
            'X-Session-Token': self.session_token,
            'Content-Type': 'application/json'
        }

        response = requests.post(
            f'{BASE_URL}/model/claude-3-sonnet/invoke',
            headers=headers,
            json={
                'messages': [{'role': 'user', 'content': message}],
                'max_tokens': 100
            }
        )

        return response.json()

# Usage
client = BedrockClient()
client.login()  # Login once

# Use for hours without TOTP!
result1 = client.call_bedrock("Hello!")
result2 = client.call_bedrock("How are you?")
result3 = client.call_bedrock("Tell me a joke")
# ... no TOTP needed for any of these!
```

---

## 🌐 JavaScript Example

```javascript
const axios = require('axios');
const speakeasy = require('speakeasy');

class BedrockClient {
  constructor(apiKey, totpSecret, baseUrl) {
    this.apiKey = apiKey;
    this.totpSecret = totpSecret;
    this.baseUrl = baseUrl;
    this.sessionToken = null;
  }

  async login() {
    // Generate TOTP code
    const totpCode = speakeasy.totp({
      secret: this.totpSecret,
      encoding: 'base32'
    });

    // Login
    const response = await axios.post(`${this.baseUrl}/auth/login`, {
      api_key: this.apiKey,
      totp_code: totpCode
    });

    this.sessionToken = response.data.session_token;
    console.log(`✅ Logged in! Token valid until ${response.data.expires_at}`);
  }

  async callBedrock(message) {
    if (!this.sessionToken) {
      await this.login();
    }

    const response = await axios.post(
      `${this.baseUrl}/model/claude-3-sonnet/invoke`,
      {
        messages: [{ role: 'user', content: message }],
        max_tokens: 100
      },
      {
        headers: {
          'X-Session-Token': this.sessionToken
        }
      }
    );

    return response.data;
  }
}

// Usage
const client = new BedrockClient(
  process.env.BEDROCK_API_KEY,
  process.env.TOTP_SECRET,
  'https://bedrock-proxy.example.com'
);

(async () => {
  await client.login();  // Login once

  // Use for hours without TOTP!
  const result1 = await client.callBedrock('Hello!');
  const result2 = await client.callBedrock('How are you?');
  // ... no TOTP needed!
})();
```

---

## 🔄 Token Management

### Refresh Token (Extend Session)

```bash
# Refresh before expiration to extend session
curl -X POST https://bedrock-proxy.example.com/auth/refresh \
  -H "X-Session-Token: $SESSION_TOKEN"

# Response: New token with extended expiration
{
  "session_token": "bdrk_sess_new_token...",
  "expires_at": "2025-01-08T06:30:00Z",
  "expires_in": 43200
}
```

### Logout (Revoke Token)

```bash
# Logout to invalidate session
curl -X POST https://bedrock-proxy.example.com/auth/logout \
  -H "X-Session-Token: $SESSION_TOKEN"

# Response:
{
  "message": "Logged out successfully"
}
```

### List Active Sessions

```bash
# See all your active sessions
curl -X GET https://bedrock-proxy.example.com/auth/sessions \
  -H "X-Session-Token: $SESSION_TOKEN"

# Response:
{
  "sessions": [
    {
      "id": 1,
      "created_at": "2025-01-06T18:30:00Z",
      "expires_at": "2025-01-07T06:30:00Z",
      "last_used_at": "2025-01-06T19:45:00Z",
      "ip_address": "203.0.113.5",
      "user_agent": "curl/7.68.0"
    }
  ]
}
```

---

## 🔒 Security Features

| Feature | Benefit |
|---------|---------|
| **12+ hour validity** | Login once per day (or week) |
| **Automatic expiration** | Tokens expire after set duration |
| **IP tracking** | Session tied to IP address |
| **Revocable** | Can logout to invalidate immediately |
| **Audit logging** | All session usage tracked |
| **Secure tokens** | 32-byte random, base64-encoded |

---

## ⚙️ Configuration Options

### Session Duration

```yaml
# Short sessions (more secure)
SESSION_DURATION: "1h"

# Standard (balanced)
SESSION_DURATION: "12h"

# Long sessions (convenience)
SESSION_DURATION: "24h"
SESSION_DURATION: "7d"
SESSION_DURATION: "30d"
```

### Hybrid Mode (Both Methods Work)

```yaml
# Accept both session tokens AND API key + TOTP
AUTH_MODE: "hybrid"

# Users can choose:
# - Session token (convenient, no TOTP)
# - API key + TOTP (more secure, each request)
```

---

## 🧪 Testing

### Test Login

```bash
# Get TOTP code
TOTP_CODE=$(oathtool --totp -b "$TOTP_SECRET")

# Login
RESPONSE=$(curl -s -X POST https://bedrock-proxy.example.com/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"api_key\":\"$API_KEY\",\"totp_code\":\"$TOTP_CODE\"}")

# Extract token
SESSION_TOKEN=$(echo $RESPONSE | jq -r '.session_token')

echo "Session token: $SESSION_TOKEN"
```

### Test Session Token

```bash
# Health check with session token
curl -H "X-Session-Token: $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health

# Should work without TOTP!
```

### Test Expiration

```bash
# Try using expired token (after 12h)
curl -H "X-Session-Token: $OLD_TOKEN" \
  https://bedrock-proxy.example.com/health

# Expected: 401 Unauthorized - "Session token expired"
```

---

## 📊 Comparison

### Traditional (API Key + TOTP)
```bash
✅ Most secure (2FA every request)
❌ Annoying (need phone every time)
❌ Hard to automate
❌ Slow (lookup phone)
```

### Session Token (12h)
```bash
✅ Convenient (login once)
✅ Easy to automate
✅ Fast (no phone needed)
✅ Still secure (2FA at login)
⚠️  Less secure if token leaked (but expires in 12h)
```

---

## 🎯 Best Practices

1. **Set appropriate duration**
   - Development: 24h
   - Production: 12h
   - High security: 1h

2. **Rotate regularly**
   - Auto-refresh before expiration
   - Or re-login daily

3. **Secure storage**
   - Save token in secure location
   - Use environment variables
   - Don't commit to git

4. **Monitor sessions**
   - List active sessions regularly
   - Revoke suspicious sessions
   - Check audit logs

---

## ✅ Summary

**What You Get:**
- 🔐 Login once with API Key + TOTP
- 🎟️ Get session token (valid 12h+)
- 🚀 Use token for all requests (no TOTP!)
- 🔄 Refresh to extend
- 🚪 Logout to revoke

**User Experience:**
```bash
# Morning: Login once
SESSION_TOKEN=$(curl -X POST .../auth/login ...)

# All day: Just use token
curl -H "X-Session-Token: $SESSION_TOKEN" ...
curl -H "X-Session-Token: $SESSION_TOKEN" ...
curl -H "X-Session-Token: $SESSION_TOKEN" ...

# No phone needed until tomorrow!
```

**Perfect for automation, scripts, and long-running applications!** 🎉
