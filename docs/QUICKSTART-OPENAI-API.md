# Quick Start - OpenAI-Compatible API with Bedrock

This guide shows you how to use the OpenAI-compatible API to access AWS Bedrock models (Claude) using the familiar OpenAI API format.

## What You Get

✅ **OpenAI-compatible endpoints** - Use Claude models with OpenAI SDK
✅ **Bedrock Converse API** - Latest AWS Bedrock unified API
✅ **Multiple Claude models** - Haiku, Sonnet, Opus, 3.5 Sonnet
✅ **Drop-in replacement** - Swap `base_url` and use Bedrock models
✅ **Same authentication** - Your existing auth layer still works

## Architecture

```
Your App/OpenAI SDK
       ↓
   OpenAI API format
   POST /v1/chat/completions
   { "model": "claude-3-sonnet", ... }
       ↓
   AI Gateway (Translation Layer)
   • Translate OpenAI → Bedrock Converse
   • Route to Bedrock provider
   • Authenticate with AWS (IRSA/IAM)
       ↓
   AWS Bedrock Converse API
   • Invoke Claude models
   • Return response
       ↓
   AI Gateway (Translation Layer)
   • Translate Bedrock → OpenAI format
       ↓
   Your App receives OpenAI-compatible response
```

## Setup (2 minutes)

### 1. Ensure AWS Access

```bash
# Test AWS credentials
aws sts get-caller-identity

# Verify Bedrock model access
aws bedrock list-foundation-models --region us-east-1 | grep -i claude
```

### 2. Build and Run

```bash
# Build
go build -v ./cmd/server

# Run (no auth for testing)
export AWS_REGION=us-east-1
./server
```

You should see:
```
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║              🚀 Multi-Provider AI Gateway                   ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝

Configuration:
  • HTTP Port:         8080
  • Authentication:    false
  • Enabled Providers: bedrock

API Endpoints:
  • OpenAI-compatible: http://localhost:8080/v1/chat/completions
  • List models:       http://localhost:8080/v1/models
  • Native Bedrock:    http://localhost:8080/providers/bedrock/...
  • Health check:      http://localhost:8080/health
  • Metrics:           http://localhost:8080/metrics

🎯 Ready to accept requests!
```

### 3. Test It

```bash
# Simple test
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [
      {"role": "user", "content": "Say hello in one sentence"}
    ]
  }' | jq '.choices[0].message.content'
```

## Usage Examples

### Python with OpenAI SDK

```python
from openai import OpenAI

# Point to your local proxy
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed"  # Unless you enable auth
)

# Use Claude models!
response = client.chat.completions.create(
    model="claude-3-sonnet",
    messages=[
        {"role": "user", "content": "What is 2+2?"}
    ],
    max_tokens=100
)

print(response.choices[0].message.content)
# Output: "2 + 2 equals 4."
```

### JavaScript/TypeScript with OpenAI SDK

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'not-needed'
});

const response = await client.chat.completions.create({
  model: 'claude-3-haiku',
  messages: [
    { role: 'user', content: 'Hello!' }
  ],
  max_tokens: 100
});

console.log(response.choices[0].message.content);
```

### LangChain

```python
from langchain_openai import ChatOpenAI

# Use with LangChain
llm = ChatOpenAI(
    model="claude-3-sonnet",
    openai_api_base="http://localhost:8080/v1",
    openai_api_key="not-needed"
)

response = llm.invoke("What is the capital of France?")
print(response.content)
```

### curl

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Explain quantum physics in simple terms"}
    ],
    "max_tokens": 200,
    "temperature": 0.7
  }'
```

## Available Models

| Model Name | Bedrock Model | Speed | Cost | Best For |
|------------|---------------|-------|------|----------|
| `claude-3-haiku` | anthropic.claude-3-haiku-20240307-v1:0 | ⚡ Fastest | $ Cheapest | Simple tasks, quick responses |
| `claude-3-sonnet` | anthropic.claude-3-sonnet-20240229-v1:0 | ⚡⚡ Fast | $$ Moderate | Balanced tasks, most use cases |
| `claude-3-opus` | anthropic.claude-3-opus-20240229-v1:0 | ⚡⚡⚡ Slow | $$$ Expensive | Complex reasoning, analysis |
| `claude-3-5-sonnet` | anthropic.claude-3-5-sonnet-20240620-v1:0 | ⚡⚡ Fast | $$ Moderate | Latest model, enhanced capabilities |

## Request Parameters

All standard OpenAI parameters are supported:

```json
{
  "model": "claude-3-sonnet",           // Required
  "messages": [...],                     // Required
  "max_tokens": 4096,                   // Optional (default: 4096)
  "temperature": 1.0,                   // Optional (0.0 - 2.0, default: 1.0)
  "top_p": 1.0,                         // Optional (0.0 - 1.0)
  "stop": ["stop", "sequences"],        // Optional
  "stream": false                       // Optional (streaming coming soon)
}
```

## Response Format

Standard OpenAI format:

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "claude-3-sonnet",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "The response text here..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 25,
    "total_tokens": 40
  }
}
```

## With Authentication

Enable authentication for production use:

```bash
# Generate API key
export API_KEY=$(openssl rand -hex 32)

# Run with auth
export AWS_REGION=us-east-1
export AUTH_ENABLED=true
export AUTH_MODE=api_key
export BEDROCK_API_KEY_USER1=$API_KEY
./server

# Use API key in requests
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"model": "claude-3-haiku", "messages": [...]}'
```

## Comparison: OpenAI API vs Native Bedrock

### OpenAI-Compatible (Recommended for most use cases)

```bash
# Simple, familiar API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Pros:**
- ✅ Familiar OpenAI format
- ✅ Works with OpenAI SDKs
- ✅ Easy migration from OpenAI
- ✅ Framework compatible (LangChain, etc.)

**Cons:**
- ⚠️ Slight translation overhead (~10ms)

### Native Bedrock Converse API

```bash
# More verbose, provider-specific
curl -X POST http://localhost:8080/providers/bedrock/model/anthropic.claude-3-sonnet-20240229-v1:0/converse \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": [{"text": "Hello"}]}],
    "inferenceConfig": {"maxTokens": 100}
  }'
```

**Pros:**
- ✅ Native API (no translation)
- ✅ Access to provider-specific features

**Cons:**
- ⚠️ Verbose model IDs
- ⚠️ Provider-specific format
- ⚠️ Harder to migrate

## Testing

Run the comprehensive test suite:

```bash
./test-openai-api.sh
```

This tests:
- ✅ Health checks
- ✅ Model listing
- ✅ Chat completions
- ✅ System messages
- ✅ Multi-turn conversations
- ✅ Error handling

## Migration from OpenAI

**Before (OpenAI):**
```python
from openai import OpenAI

client = OpenAI(api_key="sk-...")
response = client.chat.completions.create(
    model="gpt-4",  # OpenAI model
    messages=[...]
)
```

**After (Bedrock via Gateway):**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",  # ← Only change
    api_key="your-gateway-api-key"        # ← Your gateway key
)
response = client.chat.completions.create(
    model="claude-3-sonnet",  # ← Use Claude instead
    messages=[...]            # ← Same format!
)
```

That's it! Just change 2 lines.

## Next Steps

1. ✅ **Test the API** - Run `./test-openai-api.sh`
2. ✅ **Try different models** - Haiku for speed, Opus for quality
3. 🔜 **Enable authentication** - See `docs/AUTHORIZATION.md`
4. 🔜 **Deploy to production** - See `deployments/kubernetes/`
5. 🔜 **Add more providers** - Azure, OpenAI, Anthropic Direct, Vertex (coming soon)

## Troubleshooting

### "Model not found"
- Ensure you have Bedrock model access in AWS Console
- Check: AWS Console → Bedrock → Model access

### "Authentication failed"
- Verify AWS credentials: `aws sts get-caller-identity`
- Check IAM permissions include `bedrock:InvokeModel`

### "Connection refused"
- Ensure server is running on correct port
- Check: `curl http://localhost:8080/health`

## More Information

- **Full Architecture**: `docs/MULTI-PROVIDER-ARCHITECTURE.md`
- **Testing Guide**: `TESTING.md`
- **Authorization Setup**: `docs/AUTHORIZATION.md`
- **Security Guide**: `docs/SECURITY-QUICKSTART.md`

---

**You're ready to use OpenAI SDK with AWS Bedrock! 🚀**
