# User Quick Start - API Key + 2FA

**Simple guide for Alice, Bob, and Charlie to use Bedrock Proxy with 2FA**

---

## 📱 One-Time Setup (5 minutes)

### Step 1: Install Google Authenticator

Download on your phone:
- **iOS**: App Store → "Google Authenticator"
- **Android**: Play Store → "Google Authenticator"

### Step 2: Scan Your QR Code

Admin will send you a QR code image (e.g., `qr-Alice.png`)

1. Open Google Authenticator app
2. Tap **"+"** (Add account)
3. Choose **"Scan QR code"**
4. Scan the image
5. You'll see: **Bedrock Proxy (your-email@example.com)**
6. Below it: A 6-digit code that changes every 30 seconds

✅ Done! You're ready to use 2FA

### Step 3: Save Your API Key

```bash
# On your laptop - save API key
echo "export BEDROCK_API_KEY='bdrk_your_key_here'" >> ~/.bashrc
source ~/.bashrc
```

---

## 🚀 Daily Usage

### Every Request Needs TWO Things:

1. **API Key** (saved in Step 3)
2. **TOTP Code** (from Google Authenticator app)

### Example: Health Check

```bash
# 1. Look at Google Authenticator app on your phone
#    Current code: 123456

# 2. Make request with BOTH headers
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: 123456" \
     https://bedrock-proxy.example.com/health
```

### Example: Call Claude

```bash
# 1. Get code from Google Authenticator: 456789

# 2. Make request
curl -X POST \
  -H "X-API-Key: $BEDROCK_API_KEY" \
  -H "X-TOTP-Code: 456789" \
  -H "Content-Type: application/json" \
  https://bedrock-proxy.example.com/model/claude-3-sonnet/invoke \
  -d '{"messages":[{"role":"user","content":"Hello!"}]}'
```

---

## 💻 Code Examples

### Python

```python
import os
import pyotp
import requests

# One-time setup: Save TOTP secret
# (Ask admin for your secret, or extract from QR code)
TOTP_SECRET = 'JBSWY3DPEHPK3PXP'  # Your secret
API_KEY = os.getenv('BEDROCK_API_KEY')

# Generate current code
totp = pyotp.TOTP(TOTP_SECRET)
code = totp.now()  # e.g., '123456'

# Make request
headers = {
    'X-API-Key': API_KEY,
    'X-TOTP-Code': code
}

response = requests.get(
    'https://bedrock-proxy.example.com/health',
    headers=headers
)
print(response.json())
```

**Install required package:**
```bash
pip install pyotp
```

### JavaScript (Node.js)

```javascript
const speakeasy = require('speakeasy');
const fetch = require('node-fetch');

// One-time setup
const TOTP_SECRET = 'JBSWY3DPEHPK3PXP';  // Your secret
const API_KEY = process.env.BEDROCK_API_KEY;

// Generate current code
const totpCode = speakeasy.totp({
  secret: TOTP_SECRET,
  encoding: 'base32'
});

// Make request
fetch('https://bedrock-proxy.example.com/health', {
  headers: {
    'X-API-Key': API_KEY,
    'X-TOTP-Code': totpCode
  }
})
  .then(res => res.json())
  .then(data => console.log(data));
```

**Install required package:**
```bash
npm install speakeasy
```

---

## ❓ FAQ

### Q: The code changed before I could use it!

**A:** Codes change every 30 seconds. If you're typing manually, be quick! Or use the code examples above to auto-generate.

### Q: I get "Invalid TOTP code" error

**A:** Common causes:
- Using an old code (they expire after 30s)
- Phone time not synced (Settings → Date & Time → Automatic)
- Wrong secret configured

### Q: I lost my phone!

**A:** Use backup codes! Admin gave you 10 backup codes during setup. Each can be used once:
```bash
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: ABCD-1234" \
     https://bedrock-proxy.example.com/health
```

Then contact admin to generate new QR code.

### Q: Can I use Authy instead of Google Authenticator?

**A:** Yes! Any TOTP app works:
- Google Authenticator
- Authy
- Microsoft Authenticator
- 1Password
- Bitwarden

### Q: Where do I find my TOTP secret?

**A:**
1. If you scanned QR code: It's embedded (you don't need it)
2. For code automation: Ask admin for your secret string

### Q: Do I need internet for TOTP?

**A:** No! TOTP works offline. The app generates codes based on time, not network.

---

## 🔒 Security Tips

✅ **DO:**
- Keep your phone secure (lock screen)
- Save backup codes in password manager
- Use latest version of authenticator app
- Contact admin if phone is lost

❌ **DON'T:**
- Share your TOTP secret with anyone
- Screenshot QR codes (delete after scanning)
- Use the same backup code twice (they're single-use)
- Ignore "Invalid TOTP" errors (might be attack)

---

## 🆘 Quick Troubleshooting

| Error | Fix |
|-------|-----|
| "Missing API key" | Add `-H "X-API-Key: $BEDROCK_API_KEY"` |
| "Invalid API key" | Check your key is correct |
| "2FA required" | Add `-H "X-TOTP-Code: 123456"` |
| "Invalid TOTP code" | Use fresh code from app |
| "Time sync error" | Enable automatic time on phone |

---

## 📞 Need Help?

Contact your admin if:
- Lost phone or can't generate codes
- Backup codes not working
- Need new QR code
- Having persistent errors

---

## ✅ Quick Command Reference

### With Manual TOTP (look at phone)
```bash
# Get code from Google Authenticator, then:
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: 123456" \
     https://bedrock-proxy.example.com/health
```

### With Auto TOTP (using oathtool)
```bash
# Install: brew install oath-toolkit
CODE=$(oathtool --totp -b "YOUR_SECRET")
curl -H "X-API-Key: $BEDROCK_API_KEY" \
     -H "X-TOTP-Code: $CODE" \
     https://bedrock-proxy.example.com/health
```

### Save for Easy Access
```bash
# Add to ~/.bashrc or ~/.zshrc
bedrock() {
  local code=$(oathtool --totp -b "$TOTP_SECRET")
  curl -H "X-API-Key: $BEDROCK_API_KEY" \
       -H "X-TOTP-Code: $code" \
       "$@"
}

# Then use:
bedrock https://bedrock-proxy.example.com/health
```

---

**You're all set! 🎉**

Remember:
1. API Key = Your identity
2. TOTP Code = Proof you have your phone (something you have)
3. Together = Strong 2-factor authentication
