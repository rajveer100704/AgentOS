#!/bin/bash
# Scenario 1: Normal allowed request
# Demonstrates the happy path through AegisFlow:
#   Agent → Edge → Gateway → Policy (allow) → OpenAI → Evidence → Response

GATEWAY="${1:-http://localhost:8080}"
KEY="${2:-dev-key-1}"

echo "Sending a normal chat completion request..."
echo "Expected: 200 OK — policy allows, provider responds"

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X POST "${GATEWAY}/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${KEY}" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What is 2 + 2?"}
    ],
    "max_tokens": 50
  }')

HTTP_CODE=$(echo "${RESPONSE}" | tail -1)
BODY=$(echo "${RESPONSE}" | head -1)

if [ "${HTTP_CODE}" = "200" ]; then
  echo "✅ Status: ${HTTP_CODE} (expected 200)"
  echo "Response excerpt: $(echo "${BODY}" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['choices'][0]['message']['content'])" 2>/dev/null || echo "${BODY}" | head -c 200)"
else
  echo "❌ Status: ${HTTP_CODE} (expected 200)"
  echo "Body: ${BODY}"
fi
