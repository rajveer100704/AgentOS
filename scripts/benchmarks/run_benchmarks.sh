#!/bin/bash
set -euo pipefail

echo "AgentOS Governance Overhead Benchmarks"
echo "========================================="
echo ""
go test ./scripts/benchmarks/benchgovern/ -bench=Benchmark -benchmem -count=3 -run='^$' | tee benchmark_results.txt
echo ""
echo "Results saved to benchmark_results.txt"
