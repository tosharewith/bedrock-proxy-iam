# Testing Guide - OpenAI-Compatible API with Bedrock Converse

## Prerequisites

Before testing, ensure you have:

1. **AWS Credentials** configured (one of):
   - AWS CLI configured: `aws configure`
   - Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
   - IAM role (for EKS/EC2)
   - IRSA (for EKS pods)

2. **AWS Bedrock Model Access**:
   - Go to AWS Console → Bedrock → Model access
   - Request access to Claude 3 models
   - Wait for approval (usually instant for Haiku/Sonnet)

3. **Tools**:
   - `curl` - for making HTTP requests
   - `jq` - for JSON formatting (optional but recommended)

## Quick Start

### 1. Build the Server

```bash
# Build the application
go build -v ./cmd/server

# The binary will be created as ./server
```

### 2. Start the Server

```bash
# Option A: Run without authentication (for testing)
export AWS_REGION=us-east-1
./server

# Option B: Run with API key authentication
export AWS_REGION=us-east-1
export AUTH_ENABLED=true
export AUTH_MODE=api_key
export BEDROCK_API_KEY_TEST=my-secret-key-123
./server
```

The server will start on port 8080 by default and display a startup banner with available endpoints.

### 3. Run the Test Suite

```bash
# Run all tests (without authentication)
./test-openai-api.sh

# Run tests with API key
API_KEY=my-secret-key-123 ./test-openai-api.sh

# Run against different URL
BASE_URL=http://localhost:8080 ./test-openai-api.sh
```

## Manual Testing Examples

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "service": "ai-gateway",
  "status": "healthy"
}
```

### 2. List Available Models

```bash
curl http://localhost:8080/v1/models | jq '.'
```

Expected response:
```json
{
  "object": "list",
  "data": [
    {
      "id": "claude-3-opus",
      "object": "model",
      "created": 1234567890,
      "owned_by": "bedrock"
    },
    ...
  ]
}
```

### 3. Chat Completion with Claude 3 Haiku

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ],
    "max_tokens": 100
  }' | jq '.'
```

Expected response:
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "claude-3-haiku",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "2 + 2 equals 4."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 8,
    "total_tokens": 18
  }
}
```

### 4. Chat Completion with System Message

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {"role": "system", "content": "You are a helpful coding assistant."},
      {"role": "user", "content": "Write a hello world in Python"}
    ],
    "max_tokens": 200,
    "temperature": 0.7
  }' | jq '.'
```

### 5. Multi-turn Conversation

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [
      {"role": "user", "content": "My favorite color is blue"},
      {"role": "assistant", "content": "That'\''s nice! Blue is a calming color."},
      {"role": "user", "content": "What is my favorite color?"}
    ],
    "max_tokens": 50
  }' | jq '.'
```

### 6. Using Different Claude Models

```bash
# Claude 3 Haiku (fastest, cheapest)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Hello!"}]
  }' | jq '.choices[0].message.content'

# Claude 3 Sonnet (balanced)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}]
  }' | jq '.choices[0].message.content'

# Claude 3 Opus (most capable, requires access approval)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus",
    "messages": [{"role": "user", "content": "Hello!"}]
  }' | jq '.choices[0].message.content'

# Claude 3.5 Sonnet (latest)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}]
  }' | jq '.choices[0].message.content'
```

## Testing with OpenAI SDK

You can use the official OpenAI Python SDK with your proxy:

```python
from openai import OpenAI

# Point to your local proxy
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed-if-auth-disabled"  # or your actual API key
)

# Use Claude models with OpenAI SDK!
response = client.chat.completions.create(
    model="claude-3-sonnet",  # Bedrock model through OpenAI API
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "What is the capital of France?"}
    ],
    max_tokens=100,
    temperature=0.7
)

print(response.choices[0].message.content)
```

## Testing with curl (with authentication)

If authentication is enabled:

```bash
# Set your API key
export API_KEY="your-api-key-here"

# Make requests with X-API-Key header
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Native Bedrock API Testing

You can still use the native Bedrock Converse API:

```bash
curl -X POST http://localhost:8080/providers/bedrock/model/anthropic.claude-3-haiku-20240307-v1:0/converse \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": [{"text": "Hello!"}]
      }
    ],
    "inferenceConfig": {
      "maxTokens": 100,
      "temperature": 0.7
    }
  }' | jq '.'
```

## Monitoring and Metrics

Check Prometheus metrics:

```bash
curl http://localhost:8080/metrics
```

Metrics include:
- `bedrock_proxy_requests_total` - Total requests
- `bedrock_proxy_request_duration_seconds` - Request latency
- `http_requests_total` - HTTP request count by method and status

## Troubleshooting

### Issue: "Model not found" error

**Solution**: Check that you have access to the model in AWS Bedrock:
```bash
aws bedrock list-foundation-models --region us-east-1 | grep -i claude
```

### Issue: "Authentication failed" error

**Possible causes**:
1. No AWS credentials configured
2. IAM role lacks `bedrock:InvokeModel` permission
3. Region mismatch

**Solution**: Check AWS credentials and permissions:
```bash
# Test AWS credentials
aws sts get-caller-identity

# Check IAM permissions (if using IAM role)
aws iam get-role-policy --role-name your-role-name --policy-name your-policy-name
```

### Issue: Server won't start - "Failed to load router config"

**Solution**: Ensure config file exists:
```bash
ls -la configs/model-mapping.yaml

# Or set a custom path
export MODEL_MAPPING_CONFIG=/path/to/your/config.yaml
```

### Issue: Rate limiting errors

**Solution**: Bedrock has quotas. Check your quotas:
```bash
aws service-quotas list-service-quotas \
  --service-code bedrock \
  --region us-east-1
```

## Expected Performance

- **Claude 3 Haiku**: ~1-2 seconds for short responses
- **Claude 3 Sonnet**: ~2-4 seconds for short responses
- **Claude 3 Opus**: ~4-8 seconds for short responses

Token limits:
- **Input**: Up to 200K tokens (Claude 3 family)
- **Output**: Configurable via `max_tokens` (default: 4096)

## Next Steps

1. ✅ Test OpenAI-compatible API with Claude models
2. ✅ Verify authentication works correctly
3. 🔜 Test streaming responses (coming soon)
4. 🔜 Add more providers (Azure, OpenAI, Anthropic Direct, Vertex AI)

## Support

If you encounter issues:
1. Check server logs
2. Verify AWS credentials and permissions
3. Ensure Bedrock model access is granted
4. Check the configuration file syntax
5. Review the architecture documentation: `docs/MULTI-PROVIDER-ARCHITECTURE.md`
