# Client Configuration - Laptop Setup

This guide shows how to configure your laptop to access the Bedrock proxy with your API key.

## 🔧 Quick Setup for Your Laptop

### 1. Save Your API Key

Create a config file with your API key:

```bash
# Create config directory
mkdir -p ~/.bedrock-proxy

# Save your API key (replace with your actual key)
echo "export BEDROCK_API_KEY='bdrk_your_api_key_here'" > ~/.bedrock-proxy/config

# Add to your shell profile
echo "source ~/.bedrock-proxy/config" >> ~/.bashrc  # or ~/.zshrc
source ~/.bashrc
```

### 2. Test Access

```bash
# Test with your API key
curl -H "X-API-Key: $BEDROCK_API_KEY" \
  https://bedrock-proxy.example.com/health

# Expected response:
# {"status":"healthy","service":"bedrock-proxy"}
```

---

## 📱 Usage Examples

### Command Line (curl)

```bash
# Health check
curl -H "X-API-Key: $BEDROCK_API_KEY" \
  https://bedrock-proxy.example.com/health

# Invoke Bedrock model
curl -X POST https://bedrock-proxy.example.com/model/anthropic.claude-3-sonnet-20240229-v1:0/invoke \
  -H "X-API-Key: $BEDROCK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello, Claude!"}
    ]
  }'
```

### Python

```python
import os
import requests

# Load API key from environment
API_KEY = os.getenv('BEDROCK_API_KEY')
BASE_URL = 'https://bedrock-proxy.example.com'

# Configure session with API key
session = requests.Session()
session.headers.update({
    'X-API-Key': API_KEY,
    'Content-Type': 'application/json'
})

# Health check
response = session.get(f'{BASE_URL}/health')
print(response.json())

# Invoke model
payload = {
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 1024,
    "messages": [
        {"role": "user", "content": "Hello, Claude!"}
    ]
}

response = session.post(
    f'{BASE_URL}/model/anthropic.claude-3-sonnet-20240229-v1:0/invoke',
    json=payload
)
print(response.json())
```

### JavaScript/Node.js

```javascript
// config.js
require('dotenv').config();

const BEDROCK_API_KEY = process.env.BEDROCK_API_KEY;
const BASE_URL = 'https://bedrock-proxy.example.com';

async function callBedrock(message) {
  const response = await fetch(`${BASE_URL}/model/anthropic.claude-3-sonnet-20240229-v1:0/invoke`, {
    method: 'POST',
    headers: {
      'X-API-Key': BEDROCK_API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      anthropic_version: 'bedrock-2023-05-31',
      max_tokens: 1024,
      messages: [
        { role: 'user', content: message }
      ]
    })
  });

  return response.json();
}

// Usage
callBedrock('Hello, Claude!')
  .then(result => console.log(result))
  .catch(error => console.error(error));
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

func main() {
    apiKey := os.Getenv("BEDROCK_API_KEY")
    baseURL := "https://bedrock-proxy.example.com"

    // Create request
    payload := map[string]interface{}{
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens": 1024,
        "messages": []map[string]string{
            {"role": "user", "content": "Hello, Claude!"},
        },
    }

    jsonData, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST",
        baseURL+"/model/anthropic.claude-3-sonnet-20240229-v1:0/invoke",
        bytes.NewBuffer(jsonData))

    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // Handle response
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result)
}
```

---

## 🔐 With 2FA/TOTP (Optional)

If 2FA is enabled, you'll need to provide a TOTP code:

### Command Line
```bash
# Get TOTP code from Google Authenticator app
TOTP_CODE=123456

curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: $TOTP_CODE" \
     https://bedrock-proxy.example.com/health
```

### Python with 2FA
```python
import pyotp
import requests

API_KEY = os.getenv('BEDROCK_API_KEY')
TOTP_SECRET = os.getenv('TOTP_SECRET')  # From setup

# Generate TOTP code
totp = pyotp.TOTP(TOTP_SECRET)
totp_code = totp.now()

# Make request with both API key and TOTP
headers = {
    'X-API-Key': API_KEY,
    'X-TOTP-Code': totp_code,
    'Content-Type': 'application/json'
}

response = requests.get('https://bedrock-proxy.example.com/health', headers=headers)
print(response.json())
```

---

## 🌐 Browser Access (with OAuth2)

If using OAuth2/Cognito authentication:

1. **Open browser**: `https://bedrock-proxy.example.com`
2. **Redirected to AWS Cognito** login page
3. **Enter credentials** (email/password or social login)
4. **Redirected back** with session cookie
5. **Authenticated** - no API key needed in requests!

---

## 🔍 Troubleshooting

### "401 Unauthorized"

```bash
# Check if API key is set
echo $BEDROCK_API_KEY

# Verify API key format (should start with 'bdrk_')
# bdrk_a1b2c3d4e5f6...

# Test with explicit key
curl -H "X-API-Key: bdrk_your_actual_key" \
  https://bedrock-proxy.example.com/health
```

### "Connection refused"

```bash
# Check ingress endpoint
kubectl get ingress -n bedrock-system

# Test DNS resolution
nslookup bedrock-proxy.example.com

# Test connectivity
curl -v https://bedrock-proxy.example.com/health
```

### "2FA required" error

You need to provide TOTP code if 2FA is enabled:
```bash
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: 123456" \
     https://bedrock-proxy.example.com/health
```

---

## 📋 Environment Variables Summary

Add these to your `~/.bashrc` or `~/.zshrc`:

```bash
# Bedrock Proxy Configuration
export BEDROCK_API_KEY='bdrk_your_api_key_here'
export BEDROCK_PROXY_URL='https://bedrock-proxy.example.com'
export TOTP_SECRET='your_totp_secret'  # Only if 2FA enabled
```

Reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc
```

---

## 🔒 Security Best Practices

1. ✅ **Never commit API keys** to git repositories
2. ✅ **Use environment variables** or secret managers
3. ✅ **Rotate keys regularly** (every 90 days recommended)
4. ✅ **Use HTTPS only** (never HTTP)
5. ✅ **Enable 2FA** for sensitive operations
6. ✅ **Monitor usage** via audit logs
