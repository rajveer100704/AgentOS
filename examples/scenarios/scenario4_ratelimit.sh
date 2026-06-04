#!/bin/bash
# Scenario 4: Rate Limit Exceeded
# Sends 60+ requests per minute to trigger the per-tenant rate limiter.
# Expected: First N requests succeed (200), then 429 Too Many Requests.

GATEWAY="${1:-http://localhost:8080}"
KEY="${2:-dev-key-1}"

echo "Sending 35 rapid requests to trigger rate limiting..."
echo "Expected: First ~30 return 200, then 429 Too Many Requests"
echo ""

SUCCESS=0
RATELIMITED=0
OTHER=0

for i in $(seq 1 35); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${GATEWAY}/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${KEY}" \
    -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"ping"}],"max_tokens":1}')

  if [ "${CODE}" = "200" ]; then
    SUCCESS=$((SUCCESS + 1))
    printf "."
  elif [ "${CODE}" = "429" ]; then
    RATELIMITED=$((RATELIMITED + 1))
    printf "R"
  else
    OTHER=$((OTHER + 1))
    printf "?"
  fi
done

echo ""
echo ""
echo "Results: ${SUCCESS} allowed | ${RATELIMITED} rate-limited (429) | ${OTHER} other"

if [ "${RATELIMITED}" -gt 0 ]; then
  echo "✅ Rate limiter triggered correctly (${RATELIMITED} requests blocked)"
else
  echo "⚠️  No requests were rate limited."
  echo "   Increase rate_limit in aegisflow.yaml or run more requests."
fi
