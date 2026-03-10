#!/usr/bin/env bash
#
# run-parallel.sh - Run Terraform provider acceptance tests in parallel groups.
#
# Each resource group runs as a separate `go test` process. Groups that manage
# different resource types run concurrently; tests within a group run sequentially
# to avoid shared-state conflicts.
#
# Usage:
#   set -a
#   source testing/.env
#   source testing/acceptance_test_vars.env
#   set +a
#   ./testing/run-parallel.sh
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="${REPO_ROOT}/testing/parallel-logs"
mkdir -p "$LOG_DIR"

# Clean previous logs
rm -f "$LOG_DIR"/*.log

# Increase provider HTTP timeout so concurrent groups don't hit the default 5s
# ceiling when the AAP server is under parallel load.
export AAP_TIMEOUT="${AAP_TIMEOUT:-60}"

# Pre-compile the test binary once so parallel groups don't each compile.
echo "==> Pre-compiling test binary..."
TEST_BINARY="${LOG_DIR}/provider.test"
go test -c -o "$TEST_BINARY" ./internal/provider/ 2>&1
echo "    Done."
echo ""

###############################################################################
# Group definitions - parallel arrays for group names and test run patterns
#
# Grouping rationale (5 groups):
#   Split job/workflow tests into finer groups for better parallelism.
#   No shared mutable state between groups — each test creates its own
#   resources with random names. Exponential backoff retry handles
#   transient AAP task worker contention from concurrent job launches.
#
#   - job_resource (8 tests) -- job CRUD tests
#   - job_action (6 tests) -- job launch actions + host delete-with-retry
#   - workflow_job (7 tests) -- workflow job CRUD tests
#   - crud_and_wf_action (7 tests) -- CRUD resources + workflow launch actions
#   - datasources_and_eda (11 tests) -- read-only data sources + EDA
###############################################################################
TEST_GROUP_NAMES=(
  "job_resource"
  "job_action"
  "workflow_job"
  "crud_and_wf_action"
  "datasources_and_eda"
)

TEST_GROUP_PATTERNS=(
  "^TestAccAAPJob_"
  "^TestAccAAPJobAction_|^TestAccHostResourceDeleteWithRetry$"
  "^TestAccAAPWorkflowJob[^A]"
  "^TestAccHostResource$|^TestAccGroupResource$|^TestAccInventoryResource$|^TestAccInventoryResourceWithOrganizationDataSource$|^TestAccAAPWorkflowJobAction_"
  "^TestAccOrganizationDataSource|^TestAccJobTemplateDataSource$|^TestAccWorkflowJobTemplateDataSource$|^TestAccInventoryDataSource$|^TestAccEDA"
)

NUM_GROUPS=${#TEST_GROUP_NAMES[@]}

###############################################################################
# Retry wrapper - retries a group on timeout-related failures with backoff
###############################################################################
run_group_with_retry() {
  local group_name="$1"
  local run_pattern="$2"
  local log_file="$3"
  local max_retries=2
  local attempt=0
  local delay=15

  while true; do
    attempt=$((attempt + 1))
    if [ "$attempt" -gt 1 ]; then
      echo "  [retry] $group_name  attempt $attempt (delay: ${delay}s)" >> "$log_file"
      sleep "$delay"
      delay=$((delay * 3))  # 15 -> 45
    fi

    TF_ACC=1 "$TEST_BINARY" \
      -test.count=1 \
      -test.timeout 30m \
      -test.parallel="${TEST_PARALLEL:-2}" \
      -test.run "$run_pattern" \
      -test.v \
      >> "$log_file" 2>&1
    local exit_code=$?

    if [ "$exit_code" -eq 0 ]; then
      return 0
    fi

    # Check if failure is timeout-related (retryable)
    if grep -qE 'Client\.Timeout exceeded|connection refused|EOF|dial tcp.*timeout' "$log_file" 2>/dev/null; then
      if [ "$attempt" -le "$max_retries" ]; then
        echo "  [retry] $group_name  timeout detected, will retry..." >> "$log_file"
        continue
      fi
    fi

    # Non-retryable failure or max retries exhausted
    return "$exit_code"
  done
}

###############################################################################
# Launch all groups in parallel
###############################################################################
echo "==> Starting parallel acceptance test execution (${NUM_GROUPS} groups)"
echo "    Log directory: $LOG_DIR"
echo "    AAP_TIMEOUT: ${AAP_TIMEOUT}s"
echo ""

PIDS=()
WALL_START=$(date +%s)

for i in $(seq 0 $((NUM_GROUPS - 1))); do
  group_name="${TEST_GROUP_NAMES[$i]}"
  run_pattern="${TEST_GROUP_PATTERNS[$i]}"
  log_file="$LOG_DIR/${group_name}.log"

  echo "  [start] $group_name  (pattern: $run_pattern)"

  run_group_with_retry "$group_name" "$run_pattern" "$log_file" &

  PIDS+=($!)

  # Stagger launches by 3s to reduce initial connection spike
  sleep 3
done

echo ""
echo "==> All ${NUM_GROUPS} groups launched. Waiting for completion..."
echo ""

###############################################################################
# Wait for all groups and collect exit codes
###############################################################################
EXIT_CODES=()
for i in $(seq 0 $((NUM_GROUPS - 1))); do
  pid="${PIDS[$i]}"
  name="${TEST_GROUP_NAMES[$i]}"
  if wait "$pid"; then
    EXIT_CODES+=("0")
    echo "  [done]  $name  => PASS"
  else
    EXIT_CODES+=("$?")
    echo "  [done]  $name  => FAIL"
  fi
done

WALL_END=$(date +%s)
WALL_TIME=$((WALL_END - WALL_START))

###############################################################################
# Parse results from logs
###############################################################################
echo ""
echo "============================================================"
echo "  PARALLEL TEST EXECUTION SUMMARY"
echo "============================================================"
echo ""

TOTAL_PASS=0
TOTAL_FAIL=0
TOTAL_SKIP=0
OVERALL_EXIT=0

printf "%-40s %8s %6s %6s %6s %s\n" "GROUP" "TIME(s)" "PASS" "FAIL" "SKIP" "STATUS"
printf "%-40s %8s %6s %6s %6s %s\n" "----------------------------------------" "--------" "------" "------" "------" "------"

for i in $(seq 0 $((NUM_GROUPS - 1))); do
  name="${TEST_GROUP_NAMES[$i]}"
  exit_code="${EXIT_CODES[$i]}"
  log_file="$LOG_DIR/${name}.log"

  # Count PASS/FAIL/SKIP from log (top-level TestAcc tests only)
  pass_count=$(grep -cE '^--- PASS: TestAcc' "$log_file" 2>/dev/null) || pass_count=0
  fail_count=$(grep -cE '^--- FAIL: TestAcc' "$log_file" 2>/dev/null) || fail_count=0
  skip_count=$(grep -cE '^--- SKIP: TestAcc' "$log_file" 2>/dev/null) || skip_count=0

  # Sum individual test durations for this group
  group_time=$(grep -oE -e '--- (PASS|FAIL): TestAcc\S+ \([0-9]+\.[0-9]+s\)' "$log_file" 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | awk '{sum += $1} END {printf "%.2f", sum}') || group_time="0.00"

  if [ "$exit_code" = "0" ]; then
    status="PASS"
  else
    status="FAIL"
    OVERALL_EXIT=1
  fi

  TOTAL_PASS=$((TOTAL_PASS + pass_count))
  TOTAL_FAIL=$((TOTAL_FAIL + fail_count))
  TOTAL_SKIP=$((TOTAL_SKIP + skip_count))

  printf "%-40s %8s %6s %6s %6s %s\n" "$name" "$group_time" "$pass_count" "$fail_count" "$skip_count" "$status"
done

echo ""
echo "------------------------------------------------------------"
printf "%-40s %8s %6s %6s %6s\n" "TOTALS" "${WALL_TIME}s" "$TOTAL_PASS" "$TOTAL_FAIL" "$TOTAL_SKIP"
echo "------------------------------------------------------------"
echo ""
echo "Wall time: ${WALL_TIME}s"
echo "Total tests passed: $TOTAL_PASS"
echo "Total tests failed: $TOTAL_FAIL"
echo "Total tests skipped: $TOTAL_SKIP"
echo ""

if [ "$OVERALL_EXIT" -ne 0 ]; then
  echo "RESULT: FAIL - one or more groups had failures"
  echo ""
  echo "Failed group logs:"
  for i in $(seq 0 $((NUM_GROUPS - 1))); do
    if [ "${EXIT_CODES[$i]}" != "0" ]; then
      echo "  $LOG_DIR/${TEST_GROUP_NAMES[$i]}.log"
    fi
  done
else
  echo "RESULT: PASS - all tests passed"
fi

echo ""
exit "$OVERALL_EXIT"
