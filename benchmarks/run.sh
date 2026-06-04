#!/bin/bash
# AgentOS Benchmark Runner
# Runs HTTP/1.1 vs HTTP/2 benchmarks and produces a formatted report.
#
# Usage:
#   ./benchmarks/run.sh [benchtime] [output_file]
#
# Examples:
#   ./benchmarks/run.sh              # 10s per benchmark, report.md
#   ./benchmarks/run.sh 30s          # 30s per benchmark
#   ./benchmarks/run.sh 10s results.md

set -e

BENCH_TIME="${1:-10s}"
OUTPUT="${2:-benchmarks/report.md}"
TIMESTAMP=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "AgentOS Benchmark Suite"
echo "========================"
echo "Benchmark time: ${BENCH_TIME}"
echo "Output: ${OUTPUT}"
echo ""

# Run the benchmarks and capture output
echo "Running HTTP/1.1 vs HTTP/2 benchmarks..."
RAW_OUTPUT=$(go test ./benchmarks/... \
  -bench=. \
  -benchtime="${BENCH_TIME}" \
  -benchmem \
  -cpu=1,4,8 \
  -count=3 \
  2>&1)

echo "${RAW_OUTPUT}"

# Write the markdown report
cat > "${OUTPUT}" << MARKDOWN
# AgentOS Benchmark Results

**Generated:** ${TIMESTAMP}  
**Git SHA:** ${GIT_SHA}  
**Benchmark time:** ${BENCH_TIME} per test  
**Go version:** $(go version)  
**OS:** $(uname -s) $(uname -m)

## Summary

| Protocol | Concurrency | RPS (approx) | Allocs/req |
|----------|-------------|--------------|------------|
| HTTP/1.1 | Sequential  | See raw output below | — |
| HTTP/1.1 | 100 goroutines | — | — |
| HTTP/1.1 | 500 goroutines | — | — |
| HTTP/1.1 | 1000 goroutines | — | — |
| HTTP/2   | Sequential  | — | — |
| HTTP/2   | 100 streams | — | — |
| HTTP/2   | 500 streams | — | — |
| HTTP/2   | 1000 streams | — | — |

> Fill in the table from the raw results below.

## Key Observations

- **HTTP/2 multiplexing advantage**: Under high concurrency (500+), HTTP/2 should
  outperform HTTP/1.1 because all requests share a single TCP connection.
- **Head-of-line blocking**: HTTP/1.1 suffers from HOL blocking per connection.
  At 1000 concurrent goroutines this becomes the dominant bottleneck.
- **Memory efficiency**: HTTP/2 should show fewer allocations per request due to
  connection reuse and header compression (HPACK).

## Raw Benchmark Output

\`\`\`
${RAW_OUTPUT}
\`\`\`

## Methodology

- Mock handler simulates 500µs gateway overhead (excluding LLM provider latency)
- \`-benchmem\` captures allocations per operation
- \`-count=3\` runs each benchmark 3× for stability
- \`-cpu=1,4,8\` tests across different GOMAXPROCS settings
MARKDOWN

echo ""
echo "Report written to: ${OUTPUT}"
