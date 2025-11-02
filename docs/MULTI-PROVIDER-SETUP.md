# Multi-Provider Setup Guide

This guide explains how to configure and use the AI Gateway with multiple cloud providers.

## Table of Contents

- [Overview](#overview)
- [Supported Providers](#supported-providers)
- [Provider Configuration](#provider-configuration)
- [Environment Variables](#environment-variables)
- [Model Routing](#model-routing)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

The AI Gateway supports **7 major cloud AI providers**, allowing you to:

- **Unified API**: Use OpenAI-compatible API across all providers
- **Automatic Routing**: Requests are routed to the appropriate provider based on model name
- **Fallback Support**: Automatically try alternative providers if one fails
- **Cost Optimization**: Route to cost-effective providers
- **Load Balancing**: Distribute requests across multiple providers (future)

---

## Supported Providers

| Provider | Description | Authentication | Status |
|----------|-------------|----------------|--------|
| **AWS Bedrock** | Claude, Titan, Llama, Mistral models | AWS IAM/IRSA | ✅ Production |
| **Azure OpenAI** | GPT-3.5, GPT-4 deployments | API Key | ✅ Production |
| **OpenAI Direct** | GPT-3.5, GPT-4, GPT-4-turbo | API Key | ✅ Production |
| **Anthropic** | Claude 3 Opus, Sonnet, Haiku | API Key | ✅ Production |
| **Google Vertex AI** | Gemini, PaLM 2 models | OAuth2/Service Account | ✅ Production |
| **IBM Watson** | Granite, Llama 3, Mixtral | API Key | ✅ Production |
| **Oracle Cloud** | Cohere, Llama models | Auth Token | ✅ Production |

---

## Provider Configuration

### 1. AWS Bedrock

**Models**: Claude 3, Titan, Llama 2, Mistral

**Environment Variables**:
```bash
export AWS_REGION=us-east-1
# AWS credentials via IAM role or:
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
```

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

### 2. Azure OpenAI

**Models**: GPT-3.5-turbo, GPT-4, custom deployments

**Environment Variables**:
```bash
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
export AZURE_OPENAI_API_KEY=your-azure-api-key
export AZURE_API_VERSION=2024-02-15-preview  # Optional
```

**Model Mapping**:
Azure uses deployment names instead of model names. Configure in `configs/model-mapping.yaml`:

```yaml
model_mappings:
  gpt-4:
    providers:
      azure:
        deployment: gpt-4-deployment-name
```

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

### 3. OpenAI Direct

**Models**: gpt-3.5-turbo, gpt-4, gpt-4-turbo-preview

**Environment Variables**:
```bash
export OPENAI_API_KEY=sk-your-openai-api-key
export OPENAI_BASE_URL=https://api.openai.com/v1  # Optional
```

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

### 4. Anthropic Direct

**Models**: Claude 3 Opus, Sonnet, Haiku, Claude 3.5

**Environment Variables**:
```bash
export ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
export ANTHROPIC_BASE_URL=https://api.anthropic.com/v1  # Optional
```

**Features**:
- Full function/tool calling support
- System messages
- Vision capabilities (via API)

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 1024
  }'
```

**Important**: Anthropic requires `max_tokens` parameter. The gateway automatically sets it to 4096 if not provided.

---

### 5. Google Vertex AI

**Models**: Gemini 1.5 Pro, Gemini 1.5 Flash, PaLM 2

**Environment Variables**:
```bash
export GCP_PROJECT_ID=your-gcp-project-id
export GCP_LOCATION=us-central1  # Optional, default: us-central1
export GCP_ACCESS_TOKEN=your-access-token  # Or use Application Default Credentials
```

**Authentication Options**:

1. **Access Token** (quick testing):
   ```bash
   export GCP_ACCESS_TOKEN=$(gcloud auth print-access-token)
   ```

2. **Application Default Credentials** (production):
   ```bash
   gcloud auth application-default login
   # Token will be automatically obtained
   ```

3. **Service Account** (Kubernetes):
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
   ```

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-1.5-pro",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 1024
  }'
```

---

### 6. IBM Watson (watsonx.ai)

**Models**: Granite, Llama 3, Mixtral, Flan-UL2

**Environment Variables**:
```bash
export IBM_API_KEY=your-ibm-api-key
export IBM_PROJECT_ID=your-project-id
export IBM_BASE_URL=https://us-south.ml.cloud.ibm.com  # Optional
```

**Setup Steps**:

1. Create IBM Cloud account
2. Create a watsonx.ai instance
3. Create a project and note the Project ID
4. Generate an API key from IBM Cloud IAM

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "granite-13b-chat",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 512
  }'
```

**Note**: IBM Watson uses a simplified generation API. Multi-turn conversations are flattened into a single prompt.

---

### 7. Oracle Cloud AI

**Models**: Cohere Command R+, Cohere Command R, Llama

**Environment Variables**:
```bash
export ORACLE_ENDPOINT=https://inference.generativeai.us-chicago-1.oci.oraclecloud.com
export ORACLE_AUTH_TOKEN=your-oci-auth-token
export ORACLE_COMPARTMENT_ID=ocid1.compartment.oc1..xxxxx
```

**Setup Steps**:

1. Create Oracle Cloud account
2. Enable Generative AI service
3. Create a compartment or use existing one
4. Generate auth token or use OCI CLI config

**Example Request**:
```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "cohere-command-r-plus",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 1024
  }'
```

---

## Environment Variables Reference

### Complete List

```bash
# Server Configuration
export PORT=8090
export GIN_MODE=release
export AUTH_ENABLED=false
export TLS_ENABLED=false

# AWS Bedrock
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=...  # Optional if using IAM role
export AWS_SECRET_ACCESS_KEY=...

# Azure OpenAI
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
export AZURE_OPENAI_API_KEY=...
export AZURE_API_VERSION=2024-02-15-preview

# OpenAI
export OPENAI_API_KEY=sk-...
export OPENAI_BASE_URL=https://api.openai.com/v1

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-...
export ANTHROPIC_BASE_URL=https://api.anthropic.com/v1

# Google Vertex AI
export GCP_PROJECT_ID=...
export GCP_LOCATION=us-central1
export GCP_ACCESS_TOKEN=...  # Or use Application Default Credentials

# IBM Watson
export IBM_API_KEY=...
export IBM_PROJECT_ID=...
export IBM_BASE_URL=https://us-south.ml.cloud.ibm.com

# Oracle Cloud AI
export ORACLE_ENDPOINT=...
export ORACLE_AUTH_TOKEN=...
export ORACLE_COMPARTMENT_ID=...

# Model Routing
export MODEL_MAPPING_CONFIG=configs/model-mapping.yaml
```

---

## Model Routing

The gateway automatically routes requests to the appropriate provider based on the model name.

### Routing Rules

```yaml
# In configs/model-mapping.yaml
routing:
  patterns:
    - pattern: "^gpt-"
      default_provider: openai

    - pattern: "^claude-"
      default_provider: bedrock

    - pattern: "^gemini-"
      default_provider: vertex

    - pattern: "^granite-"
      default_provider: ibm
```

### Multi-Provider Models

Some models are available on multiple providers. You can specify a preferred provider:

```bash
# Use Claude on Bedrock (default)
curl -X POST http://localhost:8090/v1/chat/completions \
  -d '{"model": "claude-3-sonnet", ...}'

# Use Claude on Anthropic Direct
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "X-Provider: anthropic" \
  -d '{"model": "claude-3-sonnet-20240229", ...}'
```

### Fallback Behavior

If a provider fails, the gateway can automatically try alternatives:

```yaml
routing:
  fallback:
    enabled: true
    providers:
      - bedrock
      - anthropic
      - openai
    max_attempts: 2
```

---

## Examples

### Using Python OpenAI SDK

```python
from openai import OpenAI

# Point to the gateway
client = OpenAI(
    base_url="http://localhost:8090/v1",
    api_key="not-needed"  # Unless auth is enabled
)

# Use any supported model
response = client.chat.completions.create(
    model="claude-3-sonnet",  # Bedrock
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)

print(response.choices[0].message.content)

# Or use GPT-4
response = client.chat.completions.create(
    model="gpt-4",  # OpenAI or Azure
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)
```

### Using curl with Different Providers

```bash
# Claude on Bedrock
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Gemini on Vertex AI
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-1.5-flash",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 1024
  }'

# GPT-4 on OpenAI
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# IBM Granite
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "granite-13b-chat",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Function Calling (Supported Providers)

Function calling is supported on:
- **AWS Bedrock** (Claude models)
- **Anthropic Direct** (all Claude models)
- **OpenAI Direct** (all models)
- **Azure OpenAI** (GPT-3.5, GPT-4)
- **Google Vertex AI** (Gemini models)

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {"role": "user", "content": "What is the weather in San Francisco?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get weather for a location",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {"type": "string"}
            },
            "required": ["location"]
          }
        }
      }
    ],
    "tool_choice": "auto"
  }'
```

---

## Troubleshooting

### Provider Not Initializing

**Symptom**: "No providers initialized" error

**Solutions**:
1. Check environment variables are set correctly
2. Verify API keys are valid
3. Check logs for specific error messages
4. Ensure at least one provider is configured

### Model Not Found

**Symptom**: "Model not found" or "No provider for model"

**Solutions**:
1. Check model name spelling
2. Verify model is in `configs/model-mapping.yaml`
3. Check provider is initialized
4. Use `/v1/models` endpoint to list available models

### Authentication Errors

**AWS Bedrock**:
```bash
# Check AWS credentials
aws sts get-caller-identity

# Check region
echo $AWS_REGION
```

**Azure OpenAI**:
```bash
# Verify endpoint format
curl "$AZURE_OPENAI_ENDPOINT/openai/deployments?api-version=2024-02-15-preview" \
  -H "api-key: $AZURE_OPENAI_API_KEY"
```

**OpenAI**:
```bash
# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

**Google Vertex AI**:
```bash
# Test access token
gcloud auth print-access-token

# List models
curl "https://us-central1-aiplatform.googleapis.com/v1/projects/$GCP_PROJECT_ID/locations/us-central1/publishers/google/models" \
  -H "Authorization: Bearer $(gcloud auth print-access-token)"
```

### Rate Limiting

Different providers have different rate limits:

| Provider | Rate Limit | Solution |
|----------|------------|----------|
| OpenAI | Tier-based (TPM/RPM) | Upgrade tier or use fallback |
| Anthropic | API key-based | Contact support for increase |
| Bedrock | Region/model-based | Use multiple regions |
| Vertex AI | Quota-based | Request quota increase |
| Azure | Deployment-based | Scale deployment |

**Enable Fallback**:
```yaml
# configs/model-mapping.yaml
routing:
  fallback:
    enabled: true
```

### Monitoring

Check gateway metrics:
```bash
# Prometheus metrics
curl http://localhost:8090/metrics

# Health check
curl http://localhost:8090/health

# Provider health
curl http://localhost:8090/ready
```

---

## Best Practices

1. **Use Fallback**: Enable automatic fallback for high availability
2. **Monitor Costs**: Different providers have different pricing
3. **Region Selection**: Choose regions close to your users
4. **Authentication**: Use IAM roles/service accounts in production (not API keys in environment)
5. **Rate Limiting**: Implement client-side rate limiting
6. **Caching**: Cache responses when possible
7. **Model Selection**: Use appropriate models for tasks (cost vs. capability)

---

## Next Steps

- [Architecture Documentation](MULTI-PROVIDER-ARCHITECTURE.md)
- [Model Mapping Configuration](../configs/model-mapping.yaml)
- [Transformation Configuration](../configs/transformations.yaml)
- [Quick Start Guide](QUICKSTART-OPENAI-API.md)

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/bedrock-proxy/bedrock-iam-proxy/issues
- Documentation: https://docs.example.com
