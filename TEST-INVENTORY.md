# 🧪 Multi-Provider AI Gateway - Test Inventory

Comprehensive testing checklist for all 7 cloud providers and gateway features.

---

## 📋 Test Status Legend

- ⏳ **Not Started** - Test not yet run
- ✅ **Passed** - Test passed successfully
- ❌ **Failed** - Test failed, needs investigation
- ⚠️ **Partial** - Test partially passed (some issues)
- 🔧 **Blocked** - Test blocked by missing credentials/setup
- ⏭️ **Skipped** - Test intentionally skipped

---

## 🏗️ Infrastructure Tests

### Server Startup

| Test | Status | Command | Expected Result |
|------|--------|---------|-----------------|
| Build compiles | ⏳ | `go build ./cmd/server` | No errors |
| Server starts | ⏳ | `./server` | Server running on port 8090 |
| Health check | ⏳ | `curl http://localhost:8090/health` | `{"status":"healthy"}` |
| Ready check | ⏳ | `curl http://localhost:8090/ready` | `{"status":"ready"}` |
| Metrics endpoint | ⏳ | `curl http://localhost:8090/metrics` | Prometheus metrics |

### Provider Initialization

| Provider | Status | Environment Variables Required | Expected Log |
|----------|--------|--------------------------------|--------------|
| Bedrock | ⏳ | `AWS_REGION` | ✓ Bedrock provider initialized |
| Azure OpenAI | ⏳ | `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_API_KEY` | ✓ Azure OpenAI provider initialized |
| OpenAI | ⏳ | `OPENAI_API_KEY` | ✓ OpenAI provider initialized |
| Anthropic | ⏳ | `ANTHROPIC_API_KEY` | ✓ Anthropic provider initialized |
| Vertex AI | ⏳ | `GCP_PROJECT_ID`, `GCP_ACCESS_TOKEN` | ✓ Google Vertex AI provider initialized |
| IBM Watson | ⏳ | `IBM_API_KEY`, `IBM_PROJECT_ID` | ✓ IBM Watson provider initialized |
| Oracle Cloud | ⏳ | `ORACLE_ENDPOINT`, `ORACLE_AUTH_TOKEN`, `ORACLE_COMPARTMENT_ID` | ✓ Oracle Cloud AI provider initialized |

---

## 🎯 Provider-Specific Tests

### 1. AWS Bedrock Tests

#### Setup
```bash
export AWS_REGION=us-east-1
# Ensure AWS credentials are configured
aws sts get-caller-identity
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **1.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns Bedrock models |
| **1.2** Claude 3 Haiku chat | ⏳ | See [Test 1.2](#test-12-claude-3-haiku) | Response with text |
| **1.3** Claude 3 Sonnet chat | ⏳ | See [Test 1.3](#test-13-claude-3-sonnet) | Response with text |
| **1.4** Claude 3 Opus chat | ⏳ | See [Test 1.4](#test-14-claude-3-opus) | Response with text |
| **1.5** Claude 3.5 Sonnet | ⏳ | See [Test 1.5](#test-15-claude-35-sonnet) | Response with text |
| **1.6** Claude with system message | ⏳ | See [Test 1.6](#test-16-system-message) | System message honored |
| **1.7** Claude function calling | ⏳ | See [Test 1.7](#test-17-function-calling) | Returns tool_calls |
| **1.8** Claude multi-turn | ⏳ | See [Test 1.8](#test-18-multi-turn) | Context maintained |
| **1.9** Titan Text Express | ⏳ | See [Test 1.9](#test-19-titan) | Response with text |
| **1.10** Llama 2 70B | ⏳ | See [Test 1.10](#test-110-llama) | Response with text |
| **1.11** Mistral 8x7B | ⏳ | See [Test 1.11](#test-111-mistral) | Response with text |
| **1.12** Native Bedrock endpoint | ⏳ | `curl http://localhost:8090/providers/bedrock/...` | Native API works |
| **1.13** Legacy endpoint | ⏳ | `curl http://localhost:8090/bedrock/...` | Backward compatible |

<details>
<summary><strong>Test 1.2: Claude 3 Haiku</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Say hello in one sentence."}],
    "max_tokens": 100
  }' | jq '.'
```

**Expected**:
```json
{
  "id": "...",
  "object": "chat.completion",
  "model": "claude-3-haiku",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello! How can I assist you today?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 8,
    "total_tokens": 23
  }
}
```
</details>

<details>
<summary><strong>Test 1.7: Claude Function Calling</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "What is the weather in San Francisco?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather for a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string", "description": "City name"}
          },
          "required": ["location"]
        }
      }
    }],
    "tool_choice": "auto",
    "max_tokens": 500
  }' | jq '.'
```

**Expected**:
```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "tool_calls": [{
        "id": "...",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\"location\":\"San Francisco\"}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }]
}
```
</details>

---

### 2. Azure OpenAI Tests

#### Setup
```bash
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
export AZURE_OPENAI_API_KEY=your-key
export AZURE_API_VERSION=2024-02-15-preview
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **2.1** List deployments | ⏳ | `curl http://localhost:8090/v1/models` | Returns Azure deployments |
| **2.2** GPT-4 chat | ⏳ | See [Test 2.2](#test-22-gpt-4) | Response with text |
| **2.3** GPT-3.5-turbo chat | ⏳ | See [Test 2.3](#test-23-gpt-35) | Response with text |
| **2.4** GPT-4 with system message | ⏳ | Similar to 2.2 with system | System message honored |
| **2.5** GPT-4 function calling | ⏳ | See [Test 2.5](#test-25-functions) | Returns function_call |
| **2.6** GPT-4 with temperature | ⏳ | Request with temperature=0.9 | More creative response |
| **2.7** GPT-4 with max_tokens | ⏳ | Request with max_tokens=50 | Shorter response |
| **2.8** Native Azure endpoint | ⏳ | `curl http://localhost:8090/providers/azure/...` | Native API works |

<details>
<summary><strong>Test 2.2: GPT-4 Chat</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Explain AI in one sentence."}],
    "max_tokens": 100
  }' | jq '.'
```
</details>

---

### 3. OpenAI Direct Tests

#### Setup
```bash
export OPENAI_API_KEY=sk-your-key-here
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **3.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns OpenAI models |
| **3.2** GPT-4-turbo chat | ⏳ | See [Test 3.2](#test-32-gpt4-turbo) | Response with text |
| **3.3** GPT-3.5-turbo chat | ⏳ | See [Test 3.3](#test-33-gpt35) | Response with text |
| **3.4** Function calling | ⏳ | Similar to Bedrock test 1.7 | Returns tool_calls |
| **3.5** Vision (GPT-4V) | ⏳ | Request with image_url | Describes image |
| **3.6** JSON mode | ⏳ | Request with response_format | Returns JSON |
| **3.7** Streaming | ⏳ | Request with stream=true | SSE stream |
| **3.8** Native OpenAI endpoint | ⏳ | `curl http://localhost:8090/providers/openai/...` | Native API works |

---

### 4. Anthropic Direct Tests

#### Setup
```bash
export ANTHROPIC_API_KEY=sk-ant-your-key
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **4.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns Claude models |
| **4.2** Claude 3 Opus | ⏳ | See [Test 4.2](#test-42-opus) | Response with text |
| **4.3** Claude 3 Sonnet | ⏳ | See [Test 4.3](#test-43-sonnet) | Response with text |
| **4.4** Claude 3 Haiku | ⏳ | See [Test 4.4](#test-44-haiku) | Response with text |
| **4.5** Claude 3.5 Sonnet | ⏳ | See [Test 4.5](#test-45-35sonnet) | Response with text |
| **4.6** System message | ⏳ | Request with system message | System honored |
| **4.7** Function calling | ⏳ | Similar to test 1.7 | Returns tool_calls |
| **4.8** Multi-turn conversation | ⏳ | Multiple messages | Context maintained |
| **4.9** max_tokens required | ⏳ | Request without max_tokens | Auto-set to 4096 |
| **4.10** Native Anthropic endpoint | ⏳ | `curl http://localhost:8090/providers/anthropic/...` | Native API works |

<details>
<summary><strong>Test 4.2: Claude 3 Opus</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Explain quantum computing in simple terms."}
    ],
    "max_tokens": 200
  }' | jq '.'
```
</details>

---

### 5. Google Vertex AI Tests

#### Setup
```bash
export GCP_PROJECT_ID=your-project-id
export GCP_LOCATION=us-central1
export GCP_ACCESS_TOKEN=$(gcloud auth print-access-token)
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **5.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns Gemini models |
| **5.2** Gemini 1.5 Pro | ⏳ | See [Test 5.2](#test-52-gemini-pro) | Response with text |
| **5.3** Gemini 1.5 Flash | ⏳ | See [Test 5.3](#test-53-gemini-flash) | Response with text |
| **5.4** Gemini Pro | ⏳ | See [Test 5.4](#test-54-gemini) | Response with text |
| **5.5** System instruction | ⏳ | Request with system message | System honored |
| **5.6** Function calling | ⏳ | Similar to test 1.7 | Returns functionCall |
| **5.7** Multi-turn conversation | ⏳ | Multiple messages | Context maintained |
| **5.8** Role mapping (assistant→model) | ⏳ | Check response | Roles correctly mapped |
| **5.9** Native Vertex endpoint | ⏳ | `curl http://localhost:8090/providers/vertex/...` | Native API works |

<details>
<summary><strong>Test 5.2: Gemini 1.5 Pro</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-1.5-pro",
    "messages": [{"role": "user", "content": "Write a haiku about AI."}],
    "max_tokens": 100
  }' | jq '.'
```
</details>

---

### 6. IBM Watson Tests

#### Setup
```bash
export IBM_API_KEY=your-ibm-key
export IBM_PROJECT_ID=your-project-id
export IBM_BASE_URL=https://us-south.ml.cloud.ibm.com
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **6.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns IBM models |
| **6.2** Granite 13B Chat | ⏳ | See [Test 6.2](#test-62-granite) | Response with text |
| **6.3** Granite 13B Instruct | ⏳ | See [Test 6.3](#test-63-granite-instruct) | Response with text |
| **6.4** Llama 3 70B Instruct | ⏳ | See [Test 6.4](#test-64-llama3) | Response with text |
| **6.5** Mixtral 8x7B | ⏳ | See [Test 6.5](#test-65-mixtral) | Response with text |
| **6.6** Multi-turn (flattened) | ⏳ | Multiple messages | Flattened to prompt |
| **6.7** Parameter mapping | ⏳ | Request with OpenAI params | Mapped to IBM params |
| **6.8** Native IBM endpoint | ⏳ | `curl http://localhost:8090/providers/ibm/...` | Native API works |

<details>
<summary><strong>Test 6.2: Granite 13B Chat</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "granite-13b-chat",
    "messages": [{"role": "user", "content": "Tell me a fun fact about IBM."}],
    "max_tokens": 150
  }' | jq '.'
```
</details>

---

### 7. Oracle Cloud AI Tests

#### Setup
```bash
export ORACLE_ENDPOINT=https://inference.generativeai.us-chicago-1.oci.oraclecloud.com
export ORACLE_AUTH_TOKEN=your-oci-token
export ORACLE_COMPARTMENT_ID=ocid1.compartment.oc1..xxxxx
```

#### Tests

| Test | Status | Command | Expected |
|------|--------|---------|----------|
| **7.1** List models | ⏳ | `curl http://localhost:8090/v1/models` | Returns Oracle models |
| **7.2** Cohere Command R Plus | ⏳ | See [Test 7.2](#test-72-cohere-plus) | Response with text |
| **7.3** Cohere Command R | ⏳ | See [Test 7.3](#test-73-cohere) | Response with text |
| **7.4** Llama 3 70B | ⏳ | See [Test 7.4](#test-74-llama) | Response with text |
| **7.5** Llama 2 70B | ⏳ | See [Test 7.5](#test-75-llama2) | Response with text |
| **7.6** Role uppercase mapping | ⏳ | Check request format | Roles in UPPERCASE |
| **7.7** Multi-turn conversation | ⏳ | Multiple messages | Context maintained |
| **7.8** Native Oracle endpoint | ⏳ | `curl http://localhost:8090/providers/oracle/...` | Native API works |

<details>
<summary><strong>Test 7.2: Cohere Command R Plus</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "cohere-command-r-plus",
    "messages": [{"role": "user", "content": "Explain what Oracle Cloud is."}],
    "max_tokens": 150
  }' | jq '.'
```
</details>

---

## 🔀 Routing & Fallback Tests

### Model Routing

| Test | Status | Model Pattern | Expected Provider |
|------|--------|---------------|-------------------|
| **R.1** GPT models | ⏳ | `gpt-*` | OpenAI or Azure |
| **R.2** Claude models | ⏳ | `claude-*` | Bedrock or Anthropic |
| **R.3** Gemini models | ⏳ | `gemini-*` | Vertex AI |
| **R.4** Granite models | ⏳ | `granite-*` | IBM Watson |
| **R.5** Cohere models | ⏳ | `cohere-*` | Oracle Cloud |
| **R.6** Titan models | ⏳ | `amazon-titan-*` | Bedrock |
| **R.7** Llama models | ⏳ | `llama*` | Bedrock, IBM, or Oracle |
| **R.8** Mistral models | ⏳ | `mistral*` | Bedrock |

### Fallback Tests

<details>
<summary><strong>Test R.9: Automatic Fallback</strong></summary>

**Setup**: Configure fallback in `configs/model-mapping.yaml`
```yaml
routing:
  fallback:
    enabled: true
    providers: [bedrock, anthropic, openai]
```

**Test**: Request Claude when Bedrock is down
```bash
# Stop Bedrock (remove AWS credentials temporarily)
unset AWS_ACCESS_KEY_ID
unset AWS_SECRET_ACCESS_KEY

curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Expected**: Request automatically routed to Anthropic Direct
</details>

| Test | Status | Scenario | Expected |
|------|--------|----------|----------|
| **R.9** Automatic fallback | ⏳ | Bedrock fails → Anthropic | Uses fallback |
| **R.10** Max attempts | ⏳ | All providers fail | Error after max attempts |
| **R.11** Preferred provider | ⏳ | Header `X-Provider: anthropic` | Uses specified provider |
| **R.12** Model not found | ⏳ | Unknown model name | 404 error |

---

## 🔧 Transformation Tests

### Special Models

| Test | Status | Model | Expected Transformation |
|------|--------|-------|------------------------|
| **T.1** GPT-OSS Harmony | ⏳ | `gpt-oss-harmony` | Custom system message injected |
| **T.2** Azure deployment mapping | ⏳ | `gpt-4` on Azure | Mapped to deployment name |
| **T.3** Claude max_tokens | ⏳ | Claude without max_tokens | Auto-set to 4096 |
| **T.4** Vertex role mapping | ⏳ | `assistant` role | Mapped to `model` |
| **T.5** Oracle role uppercase | ⏳ | `user` role | Mapped to `USER` |
| **T.6** IBM prompt flattening | ⏳ | Multi-turn messages | Flattened to single prompt |

<details>
<summary><strong>Test T.1: GPT-OSS Harmony</strong></summary>

```bash
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-oss-harmony",
    "messages": [{"role": "user", "content": "Hello, who are you?"}],
    "max_tokens": 100
  }' | jq '.'
```

**Expected**: Response indicates custom system prompt about open-source collaboration
</details>

---

## 🛠️ Function Calling Comparison Tests

Test function calling works consistently across all supporting providers.

**Test Function**:
```json
{
  "name": "get_weather",
  "description": "Get current weather for a location",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {"type": "string", "description": "City name"},
      "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
    },
    "required": ["location"]
  }
}
```

| Provider | Test Status | Tool Format | Response Format |
|----------|-------------|-------------|-----------------|
| **Bedrock (Claude)** | ⏳ | OpenAI → Converse tools | Converse → OpenAI tool_calls |
| **Azure OpenAI** | ⏳ | OpenAI native | OpenAI native |
| **OpenAI Direct** | ⏳ | OpenAI native | OpenAI native |
| **Anthropic Direct** | ⏳ | OpenAI → Anthropic tools | Anthropic → OpenAI tool_calls |
| **Vertex AI (Gemini)** | ⏳ | OpenAI → Vertex functions | Vertex → OpenAI tool_calls |
| **IBM Watson** | ⏳ | Not supported | N/A |
| **Oracle Cloud** | ⏳ | Not supported | N/A |

---

## ⚠️ Error Handling Tests

| Test | Status | Scenario | Expected Response |
|------|--------|----------|-------------------|
| **E.1** Invalid API key | ⏳ | Wrong OpenAI key | 401 Unauthorized |
| **E.2** Model not found | ⏳ | Unknown model | 404 Not Found |
| **E.3** Missing max_tokens (Claude) | ⏳ | Anthropic without max_tokens | Auto-set or error |
| **E.4** Rate limiting | ⏳ | Exceed provider limit | 429 Too Many Requests |
| **E.5** Network timeout | ⏳ | Slow/dead provider | 504 Gateway Timeout |
| **E.6** Invalid request format | ⏳ | Malformed JSON | 400 Bad Request |
| **E.7** Provider unavailable | ⏳ | All providers down | 503 Service Unavailable |
| **E.8** Context length exceeded | ⏳ | Too many tokens | 400 Bad Request |

---

## 🔄 Integration Tests

### Python OpenAI SDK

<details>
<summary><strong>Test I.1: Python SDK Integration</strong></summary>

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8090/v1",
    api_key="not-needed"
)

# Test with Claude
response = client.chat.completions.create(
    model="claude-3-haiku",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)

# Test with GPT-4
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)

# Test function calling
response = client.chat.completions.create(
    model="claude-3-sonnet",
    messages=[{"role": "user", "content": "What's the weather in NYC?"}],
    tools=[{
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get weather",
            "parameters": {
                "type": "object",
                "properties": {"location": {"type": "string"}},
                "required": ["location"]
            }
        }
    }]
)
print(response.choices[0].message.tool_calls)
```

**Expected**: All requests work seamlessly
</details>

| Test | Status | SDK | Providers Tested |
|------|--------|-----|------------------|
| **I.1** Python OpenAI SDK | ⏳ | Python | Bedrock, OpenAI, Anthropic |
| **I.2** Go client | ⏳ | Go | All examples in examples/go-client |
| **I.3** TypeScript/Node SDK | ⏳ | TypeScript | Bedrock, OpenAI, Vertex |
| **I.4** curl (raw HTTP) | ⏳ | curl | All 7 providers |

---

## 📊 Performance Tests

| Test | Status | Metric | Target |
|------|--------|--------|--------|
| **P.1** Response latency (Haiku) | ⏳ | Time to first token | < 2s |
| **P.2** Response latency (GPT-3.5) | ⏳ | Time to first token | < 2s |
| **P.3** Response latency (Gemini Flash) | ⏳ | Time to first token | < 2s |
| **P.4** Concurrent requests | ⏳ | 100 parallel requests | No errors |
| **P.5** Memory usage | ⏳ | Idle memory | < 100MB |
| **P.6** Memory usage (load) | ⏳ | 100 concurrent | < 500MB |

---

## 🔐 Security Tests

| Test | Status | Feature | Expected |
|------|--------|---------|----------|
| **S.1** API key authentication | ⏳ | `AUTH_ENABLED=true` | Requires valid key |
| **S.2** 2FA/TOTP | ⏳ | TOTP enabled | Requires TOTP code |
| **S.3** Service account auth (K8s) | ⏳ | K8s service account | IRSA works |
| **S.4** TLS/HTTPS | ⏳ | `TLS_ENABLED=true` | HTTPS working |
| **S.5** AWS SigV4 signing | ⏳ | Bedrock requests | Properly signed |
| **S.6** Rate limiting | ⏳ | Multiple requests | Rate limited |

---

## 📈 Monitoring Tests

| Test | Status | Endpoint | Expected |
|------|--------|----------|----------|
| **M.1** Prometheus metrics | ⏳ | `/metrics` | Metrics exported |
| **M.2** Request counter | ⏳ | Make requests, check metrics | Counter increments |
| **M.3** Error counter | ⏳ | Trigger errors, check metrics | Error counter increments |
| **M.4** Latency histogram | ⏳ | Check metrics | Latency tracked |
| **M.5** Provider health | ⏳ | `/ready` | Shows provider status |

---

## 📝 Test Execution Plan

### Phase 1: Infrastructure (Day 1)
- [ ] Server startup tests
- [ ] Health/ready endpoints
- [ ] Provider initialization for available credentials

### Phase 2: Provider Basics (Day 1-2)
- [ ] Bedrock tests (if AWS credentials available)
- [ ] OpenAI tests (if API key available)
- [ ] One chat completion per available provider

### Phase 3: Advanced Features (Day 2-3)
- [ ] Function calling on all supporting providers
- [ ] Multi-turn conversations
- [ ] System messages
- [ ] Parameter variations (temperature, max_tokens)

### Phase 4: Routing & Transformations (Day 3)
- [ ] Model routing tests
- [ ] Fallback behavior
- [ ] Special model transformations (GPT-OSS Harmony)

### Phase 5: Integration & Performance (Day 4)
- [ ] SDK integration tests (Python, Go)
- [ ] Concurrent request handling
- [ ] Error scenarios
- [ ] Performance benchmarks

### Phase 6: Security & Monitoring (Day 5)
- [ ] Authentication tests
- [ ] TLS/HTTPS
- [ ] Metrics and monitoring
- [ ] Final validation

---

## 🎯 Success Criteria

### Must Pass (Critical)
- ✅ Server builds and starts
- ✅ At least 1 provider works end-to-end
- ✅ OpenAI-compatible API format verified
- ✅ Function calling works on at least 1 provider
- ✅ Model routing works correctly

### Should Pass (Important)
- ✅ 3+ providers working
- ✅ Python SDK integration works
- ✅ Fallback mechanism works
- ✅ Error handling returns proper HTTP codes

### Nice to Have
- ✅ All 7 providers tested
- ✅ Performance targets met
- ✅ All SDKs tested
- ✅ Transformations verified

---

## 📋 Test Results Template

Use this template to record results:

```markdown
## Test Results - [Date]

### Environment
- Server version: [git commit hash]
- Test date: [YYYY-MM-DD]
- Tester: [Name]

### Providers Available
- [ ] Bedrock
- [ ] Azure OpenAI
- [ ] OpenAI
- [ ] Anthropic
- [ ] Vertex AI
- [ ] IBM Watson
- [ ] Oracle Cloud

### Test Results Summary

#### Infrastructure: X/5 passed
#### Bedrock: X/13 passed
#### Azure: X/8 passed
#### OpenAI: X/8 passed
#### Anthropic: X/10 passed
#### Vertex AI: X/9 passed
#### IBM Watson: X/8 passed
#### Oracle Cloud: X/8 passed
#### Routing: X/12 passed
#### Transformations: X/6 passed
#### Integration: X/4 passed

### Issues Found
1. [Issue description]
2. [Issue description]

### Notes
[Any additional observations]
```

---

## 🚀 Quick Start Testing

If you only have time for basic testing, run these:

```bash
# 1. Build and start
go build ./cmd/server && ./server

# 2. Basic health check
curl http://localhost:8090/health

# 3. List models (shows available providers)
curl http://localhost:8090/v1/models | jq '.data[].id'

# 4. Simple chat (use any available model)
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Say hello"}],
    "max_tokens": 50
  }' | jq '.choices[0].message.content'

# 5. Function calling test
curl -X POST http://localhost:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "What is the weather in SF?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather",
        "parameters": {
          "type": "object",
          "properties": {"location": {"type": "string"}},
          "required": ["location"]
        }
      }
    }],
    "max_tokens": 500
  }' | jq '.choices[0].message.tool_calls'
```

---

**Next Steps**: Use this inventory to systematically test each provider and feature. Update status as you go!
