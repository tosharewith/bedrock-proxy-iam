

# Session Tokens with Simple Clients & HTTP Proxies

**Works with ANY client: curl, wget, browsers, proxies, scripts!**

---

## ✅ Yes, It Works With Everything!

### Simple HTTP Headers (Works Everywhere)

```bash
# Just add ONE header
X-Session-Token: bdrk_sess_abc123...

# Or use standard Authorization header
Authorization: Bearer bdrk_sess_abc123...
```

**That's it!** Works with:
- ✅ curl
- ✅ wget
- ✅ Any programming language
- ✅ Through HTTP proxies
- ✅ Through corporate firewalls
- ✅ Postman, Insomnia, etc.
- ✅ Browsers (JavaScript fetch/axios)

---

## 🌐 Works Through HTTP Proxies

### Corporate Proxy Example

```bash
# 1. Login through proxy to get token
curl -x http://corporate-proxy:8080 \
  -X POST https://bedrock-proxy.example.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"api_key":"bdrk_abc","totp_code":"123456"}'

# Response: {"session_token":"bdrk_sess_xyz..."}

# 2. Use token through same proxy
export SESSION_TOKEN="bdrk_sess_xyz..."
export http_proxy="http://corporate-proxy:8080"
export https_proxy="http://corporate-proxy:8080"

# 3. All requests work through proxy
curl -H "X-Session-Token: $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health
```

### SOCKS Proxy Example

```bash
# Through SOCKS5 proxy
curl --socks5 localhost:1080 \
  -H "X-Session-Token: $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health
```

---

## 🔄 Hybrid Mode (Both Methods Work)

### Option A: API Key + TOTP (Every Request)

```bash
# Traditional method - need phone every time
curl -H "X-API-Key: bdrk_abc123" \
     -H "X-TOTP-Code: 123456" \
     https://bedrock-proxy.example.com/health
```

### Option B: Session Token (Once Per 12h)

```bash
# Modern method - login once, use token for 12h
curl -H "X-Session-Token: bdrk_sess_xyz" \
     https://bedrock-proxy.example.com/health
```

**User chooses which method to use!**

---

## 📱 Simple Client Examples

### 1. wget (No Dependencies)

```bash
# Login
wget --post-data='{"api_key":"bdrk_abc","totp_code":"123456"}' \
  --header='Content-Type: application/json' \
  -O login.json \
  https://bedrock-proxy.example.com/auth/login

# Extract token
SESSION_TOKEN=$(cat login.json | grep -o '"session_token":"[^"]*' | cut -d'"' -f4)

# Use token
wget --header="X-Session-Token: $SESSION_TOKEN" \
  https://bedrock-proxy.example.com/health
```

### 2. Python requests (Simplest)

```python
import requests

# Login once
response = requests.post(
    'https://bedrock-proxy.example.com/auth/login',
    json={
        'api_key': 'bdrk_abc123',
        'totp_code': '123456'
    }
)

session_token = response.json()['session_token']

# Save token
with open('.session_token', 'w') as f:
    f.write(session_token)

# Use for 12 hours
headers = {'X-Session-Token': session_token}
requests.get('https://bedrock-proxy.example.com/health', headers=headers)
```

### 3. Browser JavaScript (fetch)

```javascript
// Login once (in console or script)
const login = await fetch('https://bedrock-proxy.example.com/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    api_key: 'bdrk_abc123',
    totp_code: '123456'
  })
});

const { session_token } = await login.json();

// Save to localStorage
localStorage.setItem('bedrock_session', session_token);

// Use in all requests
const response = await fetch('https://bedrock-proxy.example.com/health', {
  headers: {
    'X-Session-Token': localStorage.getItem('bedrock_session')
  }
});
```

### 4. Shell Script (Portable)

```bash
#!/bin/bash
# bedrock-client.sh - Simple client with session management

SESSION_FILE="$HOME/.bedrock-session"
API_KEY="${BEDROCK_API_KEY}"
TOTP_SECRET="${TOTP_SECRET}"
BASE_URL="https://bedrock-proxy.example.com"

# Function to login
login() {
    # Generate TOTP (requires oathtool)
    TOTP_CODE=$(oathtool --totp -b "$TOTP_SECRET")

    # Login and get token
    SESSION_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"api_key\":\"$API_KEY\",\"totp_code\":\"$TOTP_CODE\"}" \
        | jq -r '.session_token')

    # Save token
    echo "$SESSION_TOKEN" > "$SESSION_FILE"
    chmod 600 "$SESSION_FILE"

    echo "✅ Logged in! Session saved to $SESSION_FILE"
}

# Function to make request
request() {
    # Load token
    if [ ! -f "$SESSION_FILE" ]; then
        echo "⚠️  No session found. Logging in..."
        login
    fi

    SESSION_TOKEN=$(cat "$SESSION_FILE")

    # Make request
    curl -H "X-Session-Token: $SESSION_TOKEN" "$@"
}

# Usage
case "$1" in
    login)
        login
        ;;
    *)
        request "$@"
        ;;
esac

# Examples:
# ./bedrock-client.sh login
# ./bedrock-client.sh https://bedrock-proxy.example.com/health
# ./bedrock-client.sh -X POST https://bedrock-proxy.example.com/model/claude-3-sonnet/invoke
```

---

## 🔌 Through Different Network Setups

### 1. Corporate Network (HTTP Proxy)

```bash
# Set proxy environment variables
export http_proxy="http://proxy.corp.com:8080"
export https_proxy="http://proxy.corp.com:8080"
export no_proxy="localhost,127.0.0.1"

# Login (goes through proxy automatically)
SESSION_TOKEN=$(curl -X POST https://bedrock-proxy.example.com/auth/login \
    -H "Content-Type: application/json" \
    -d '{"api_key":"bdrk_abc","totp_code":"123456"}' \
    | jq -r '.session_token')

# All requests use proxy + session token
curl -H "X-Session-Token: $SESSION_TOKEN" \
    https://bedrock-proxy.example.com/health
```

### 2. VPN Connection

```bash
# Connect to VPN first
# Then use session token normally
curl -H "X-Session-Token: $SESSION_TOKEN" \
    https://bedrock-proxy.example.com/health
```

### 3. SSH Tunnel

```bash
# Create SSH tunnel
ssh -L 8888:bedrock-proxy.example.com:443 jumphost

# Use through tunnel
curl -H "X-Session-Token: $SESSION_TOKEN" \
    https://localhost:8888/health
```

### 4. Direct Internet

```bash
# Works without any proxy
curl -H "X-Session-Token: $SESSION_TOKEN" \
    https://bedrock-proxy.example.com/health
```

---

## 🔄 Auto-Refresh Pattern

### Keep Session Alive Forever

```bash
#!/bin/bash
# auto-refresh-session.sh

SESSION_FILE="$HOME/.bedrock-session"
BASE_URL="https://bedrock-proxy.example.com"

while true; do
    # Load current token
    SESSION_TOKEN=$(cat "$SESSION_FILE" 2>/dev/null)

    if [ -z "$SESSION_TOKEN" ]; then
        echo "No session, manual login required"
        exit 1
    fi

    # Refresh token (extends for another 12h)
    NEW_TOKEN=$(curl -s -X POST "$BASE_URL/auth/refresh" \
        -H "X-Session-Token: $SESSION_TOKEN" \
        | jq -r '.session_token')

    if [ "$NEW_TOKEN" != "null" ]; then
        echo "$NEW_TOKEN" > "$SESSION_FILE"
        echo "✅ Session refreshed at $(date)"
    else
        echo "❌ Refresh failed, re-login required"
        exit 1
    fi

    # Sleep for 11 hours (refresh before expiration)
    sleep $((11 * 60 * 60))
done

# Run in background:
# ./auto-refresh-session.sh &
```

---

## 🎯 Configuration Summary

### Server Side (Choose Mode)

```yaml
# Option 1: Session Only (simplest for users)
AUTH_MODE: "session"
SESSION_DURATION: "12h"

# Option 2: API Key + TOTP Only (most secure)
AUTH_MODE: "api_key"
REQUIRE_2FA: "true"

# Option 3: Hybrid (both work - let users choose!)
AUTH_MODE: "hybrid"
SESSION_DURATION: "12h"
REQUIRE_2FA: "true"
```

### Client Side (Both Methods Work)

```bash
# Method 1: Traditional (secure, annoying)
curl -H "X-API-Key: bdrk_abc" \
     -H "X-TOTP-Code: 123456" \
     https://...

# Method 2: Session (convenient, still secure)
curl -H "X-Session-Token: bdrk_sess_xyz" \
     https://...

# Both work in hybrid mode!
```

---

## ✅ Compatibility Matrix

| Client/Tool | Session Token | Through Proxy | Notes |
|-------------|---------------|---------------|-------|
| **curl** | ✅ | ✅ | `-H "X-Session-Token: ..."` |
| **wget** | ✅ | ✅ | `--header="X-Session-Token: ..."` |
| **Python requests** | ✅ | ✅ | `headers={'X-Session-Token': ...}` |
| **JavaScript fetch** | ✅ | ✅ | `headers: {'X-Session-Token': ...}` |
| **Go net/http** | ✅ | ✅ | `req.Header.Set("X-Session-Token", ...)` |
| **Postman** | ✅ | ✅ | Add header in Headers tab |
| **Insomnia** | ✅ | ✅ | Add header in Headers section |
| **Browser** | ✅ | ✅ | Use fetch/axios with headers |
| **HTTP Proxy** | ✅ | ✅ | Standard HTTP headers pass through |
| **SOCKS Proxy** | ✅ | ✅ | Works with `--socks5` |
| **VPN** | ✅ | ✅ | Transparent to client |
| **SSH Tunnel** | ✅ | ✅ | Works through tunnels |

---

## 🎯 Best Practice Recommendation

**Use Hybrid Mode:**

```yaml
# Server config
AUTH_MODE: "hybrid"
SESSION_DURATION: "12h"
REQUIRE_2FA: "true"
```

**Why?**
- ✅ Power users: API Key + TOTP (max security)
- ✅ Automation/scripts: Session tokens (convenience)
- ✅ Both work simultaneously
- ✅ Users choose based on their needs

**User Experience:**

```bash
# Initial setup (once)
SESSION_TOKEN=$(curl -X POST .../auth/login \
    -d '{"api_key":"bdrk_abc","totp_code":"123456"}' \
    | jq -r '.session_token')

echo "export SESSION_TOKEN='$SESSION_TOKEN'" >> ~/.bashrc

# Daily usage (simple!)
curl -H "X-Session-Token: $SESSION_TOKEN" https://...

# Auto-refresh every 11h (optional)
crontab -e
0 */11 * * * curl -X POST .../auth/refresh -H "X-Session-Token: $SESSION_TOKEN"
```

**Perfect for any client, any network setup, any use case!** 🎉
