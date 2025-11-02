#!/bin/bash

# Test script for OpenAI-compatible API with Bedrock Converse

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-}"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                                                              ║"
echo "║          OpenAI-Compatible API Test Suite                   ║"
echo "║          (Bedrock Converse Backend)                         ║"
echo "║                                                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Base URL: $BASE_URL"
echo ""

# Function to make authenticated requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    if [ -n "$API_KEY" ]; then
        if [ -n "$data" ]; then
            curl -s -X "$method" "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                -H "X-API-Key: $API_KEY" \
                -d "$data"
        else
            curl -s -X "$method" "$BASE_URL$endpoint" \
                -H "X-API-Key: $API_KEY"
        fi
    else
        if [ -n "$data" ]; then
            curl -s -X "$method" "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data"
        else
            curl -s -X "$method" "$BASE_URL$endpoint"
        fi
    fi
}

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
echo "GET /health"
response=$(curl -s "$BASE_URL/health")
if echo "$response" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Health check failed${NC}"
    echo "$response"
fi
echo ""

# Test 2: List Models
echo -e "${YELLOW}Test 2: List Models${NC}"
echo "GET /v1/models"
response=$(make_request "GET" "/v1/models" "")
if echo "$response" | grep -q "object"; then
    echo -e "${GREEN}✓ List models successful${NC}"
    echo "$response" | jq '.data[] | {id, owned_by}' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ List models failed${NC}"
    echo "$response"
fi
echo ""

# Test 3: Get Specific Model
echo -e "${YELLOW}Test 3: Get Model Info${NC}"
echo "GET /v1/models/claude-3-sonnet"
response=$(make_request "GET" "/v1/models/claude-3-sonnet" "")
if echo "$response" | grep -q "claude-3-sonnet"; then
    echo -e "${GREEN}✓ Get model info successful${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Get model info failed${NC}"
    echo "$response"
fi
echo ""

# Test 4: Simple Chat Completion (Claude 3 Sonnet)
echo -e "${YELLOW}Test 4: Chat Completion - Claude 3 Sonnet${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "model": "claude-3-sonnet",
  "messages": [
    {"role": "user", "content": "Say hello in one sentence"}
  ],
  "max_tokens": 100,
  "temperature": 0.7
}'
echo "Request:"
echo "$request_data" | jq '.' 2>/dev/null || echo "$request_data"
echo ""
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "choices"; then
    echo -e "${GREEN}✓ Chat completion successful${NC}"
    echo "Response:"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Chat completion failed${NC}"
    echo "$response"
fi
echo ""

# Test 5: Chat Completion with System Message
echo -e "${YELLOW}Test 5: Chat Completion with System Message${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "model": "claude-3-haiku",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant that speaks like a pirate."},
    {"role": "user", "content": "What is the weather like today?"}
  ],
  "max_tokens": 150,
  "temperature": 0.8
}'
echo "Request:"
echo "$request_data" | jq '.' 2>/dev/null || echo "$request_data"
echo ""
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "choices"; then
    echo -e "${GREEN}✓ Chat completion with system message successful${NC}"
    echo "Response:"
    echo "$response" | jq '.choices[0].message.content' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Chat completion with system message failed${NC}"
    echo "$response"
fi
echo ""

# Test 6: Claude 3 Opus (if available)
echo -e "${YELLOW}Test 6: Chat Completion - Claude 3 Opus${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "model": "claude-3-opus",
  "messages": [
    {"role": "user", "content": "Explain quantum computing in one sentence."}
  ],
  "max_tokens": 100
}'
echo "Request:"
echo "$request_data" | jq '.' 2>/dev/null || echo "$request_data"
echo ""
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "choices"; then
    echo -e "${GREEN}✓ Chat completion with Opus successful${NC}"
    echo "Response:"
    echo "$response" | jq '.choices[0].message.content' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Chat completion with Opus failed (may not have access)${NC}"
    echo "$response"
fi
echo ""

# Test 7: Multi-turn Conversation
echo -e "${YELLOW}Test 7: Multi-turn Conversation${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "model": "claude-3-haiku",
  "messages": [
    {"role": "user", "content": "My name is Alice"},
    {"role": "assistant", "content": "Nice to meet you, Alice! How can I help you today?"},
    {"role": "user", "content": "What is my name?"}
  ],
  "max_tokens": 50
}'
echo "Request:"
echo "$request_data" | jq '.' 2>/dev/null || echo "$request_data"
echo ""
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "choices"; then
    echo -e "${GREEN}✓ Multi-turn conversation successful${NC}"
    echo "Response:"
    echo "$response" | jq '.choices[0].message.content' 2>/dev/null || echo "$response"
    if echo "$response" | grep -qi "Alice"; then
        echo -e "${GREEN}✓ Model remembered the name!${NC}"
    fi
else
    echo -e "${RED}✗ Multi-turn conversation failed${NC}"
    echo "$response"
fi
echo ""

# Test 8: Error Handling - Invalid Model
echo -e "${YELLOW}Test 8: Error Handling - Invalid Model${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "model": "non-existent-model",
  "messages": [
    {"role": "user", "content": "Hello"}
  ]
}'
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "error"; then
    echo -e "${GREEN}✓ Error handling works correctly${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Should have returned an error${NC}"
    echo "$response"
fi
echo ""

# Test 9: Error Handling - Missing Model
echo -e "${YELLOW}Test 9: Error Handling - Missing Model Field${NC}"
echo "POST /v1/chat/completions"
request_data='{
  "messages": [
    {"role": "user", "content": "Hello"}
  ]
}'
response=$(make_request "POST" "/v1/chat/completions" "$request_data")
if echo "$response" | grep -q "error"; then
    echo -e "${GREEN}✓ Error handling works correctly${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ Should have returned an error${NC}"
    echo "$response"
fi
echo ""

# Summary
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                                                              ║"
echo "║                    Test Suite Complete                       ║"
echo "║                                                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "✓ Tests completed"
echo ""
echo "To run specific Claude models:"
echo "  • Claude 3 Haiku:  model: \"claude-3-haiku\""
echo "  • Claude 3 Sonnet: model: \"claude-3-sonnet\""
echo "  • Claude 3 Opus:   model: \"claude-3-opus\""
echo "  • Claude 3.5 Sonnet: model: \"claude-3-5-sonnet\""
echo ""
