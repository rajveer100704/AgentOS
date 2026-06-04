#!/bin/bash
# AgentOS Demo Scenarios
# ======================
# Run all 4 demo scenarios to show the full AegisFlow pipeline.
#
# Prerequisites:
#   - AegisFlow running: go run ./cmd/aegisflow/ --config configs/aegisflow.yaml
#   - A valid tenant API key in $AEGISFLOW_KEY (or set below)
#
# Usage:
#   cd examples/scenarios && ./run_all.sh

GATEWAY="${AEGISFLOW_URL:-http://localhost:8080}"
KEY="${AEGISFLOW_KEY:-dev-key-1}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   AgentOS Control Plane Demo Suite        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
echo ""
echo "Gateway: ${GATEWAY}"
echo ""

run_scenario() {
  local name="$1"
  local script="$2"
  echo -e "${YELLOW}▶ ${name}${NC}"
  bash "$script" "$GATEWAY" "$KEY"
  echo ""
}

run_scenario "Scenario 1: Allowed Request (normal flow)"        "$(dirname $0)/scenario1_allow.sh"
run_scenario "Scenario 2: Review / Human Approval Flow"         "$(dirname $0)/scenario2_review.sh"
run_scenario "Scenario 3: Blocked Request (policy violation)"   "$(dirname $0)/scenario3_block.sh"
run_scenario "Scenario 4: Rate Limit Exceeded"                  "$(dirname $0)/scenario4_ratelimit.sh"

echo -e "${GREEN}All scenarios complete.${NC}"
echo "View traces: http://localhost:8081"
echo "View spans:  http://localhost:16686 (Jaeger)"
