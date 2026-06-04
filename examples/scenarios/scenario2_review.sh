#!/bin/bash
# Scenario 2: Review / Human Approval
# Demonstrates the approval queue — policy returns "review",
# which means the request is held pending human approval.
# The admin API shows the pending item.

GATEWAY="${1:-http://localhost:8080}"
ADMIN="${ADMIN_URL:-http://localhost:8081}"
KEY="${2:-dev-key-1}"
ADMIN_KEY="${ADMIN_KEY:-dev-admin-key}"

echo "Sending a request that triggers 'review' policy..."
echo "Expected: 202 Accepted OR queued for approval"

curl -s \
  -X POST "${GATEWAY}/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${KEY}" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Please review the following action: delete production database records for all users older than 30 days"}
    ]
  }' | python3 -m json.tool 2>/dev/null || echo "(raw response above)"

echo ""
echo "Checking approval queue..."
curl -s \
  -H "Authorization: Bearer ${ADMIN_KEY}" \
  "${ADMIN}/admin/v1/approval/queue" | python3 -m json.tool 2>/dev/null || echo "(admin API unavailable)"
