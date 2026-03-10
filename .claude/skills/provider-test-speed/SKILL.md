---
name: provider-test-speed
description: Analyze and optimize Terraform provider acceptance test execution for maximum speed with 100% pass rate. Use when the user wants to run tests faster, analyze test timing, find slow tests, parallelize test execution, or optimize their test suite's throughput. Also use when the user mentions "test speed", "slow tests", "parallel tests", "test performance", "speed up tests", or asks why tests take so long.
---

# Provider Test Speed

Produce a parallel execution script that runs Terraform provider acceptance
tests at maximum speed while maintaining a 100% pass rate.

## Workflow

### 1. Discover Tests

Find all acceptance tests (functions prefixed `TestAcc`):

```bash
grep -rn 'func TestAcc.*\(t \*testing.T\)' ./internal/provider/ --include='*_test.go'
```

Build a test inventory: test name, file, resource type (from file name).

### 2. Analyze Timing (if output available)

Parse `--- PASS` / `--- FAIL` lines from prior test output to extract
per-test durations. Calculate:

- **Total wall time** — from `ok ... Ns` line
- **Sum of test durations**
- **Parallelism ratio** — sum / wall_time (1.0 = sequential, higher = parallel)

### 3. Classify Into Parallel Groups

Group tests by the resource type they manage. The grouping rule:

- Tests in the **same resource group run sequentially** (shared mutable state)
- Tests in **different groups run in parallel** (independent API endpoints)

Determine groups from file names and test prefixes. Read test configs to
verify shared resource dependencies. Data source tests (read-only) are safe
to parallel with everything.

**Target 4-5 consolidated groups, not one group per resource type.** More
groups means more concurrent API clients hitting the target infrastructure.
Too many parallel processes (8+) will overwhelm most test environments with
connection timeouts (`Client.Timeout exceeded while awaiting headers`).
Too few groups (e.g., 3) can be SLOWER than sequential if the critical-path
group becomes a bottleneck — all other groups finish early and wait for the
largest one. The sweet spot is 4-5 groups.

Consolidation strategy — merge related resource types into larger groups:
- Combine resource tests + their action tests (e.g., job resource + job action)
- Combine all read-only data source tests into one group
- Combine small independent resource groups (inventory, host, group) together
- Split job tests from job action tests — they don't share mutable state and use different code paths
- Move workflow launch action tests out of the workflow_job group into a CRUD group for better load balancing

### 4. Generate the Parallel Execution Script

Produce a self-contained bash script that:

1. **Pre-compiles the test binary once** using `go test -c -o <binary>`.
   Each parallel group then runs the compiled binary directly instead of
   invoking `go test`. This avoids N parallel compilations which waste
   time and can cause I/O contention.
2. Runs each resource group as a separate process in parallel
3. Uses `-test.run` patterns to select tests per group
4. Logs each group to a separate file
5. Waits for all groups, collects exit codes
6. Parses per-test durations from logs and reports a summary table
7. Exits non-zero if any group failed

The script must:
- Pre-compile with `go test -c` and run the binary with `-test.*` flags
- Set `TF_ACC=1` on every test process
- Use `-test.timeout 30m` per group
- Use `-test.count=1` to prevent cached results
- Accept the same env vars as the provider's existing test setup
- Print a comparison table at the end showing per-group timing and pass/fail
- Be immediately runnable without modification
- Stagger group launches by 2-3 seconds to avoid initial connection storms

**Exponential Backoff for Transient Failures**

When running multiple groups in parallel, connection timeouts and task worker
contention can cause transient failures. The script should include a retry
wrapper that:
- Retries a failed group up to 2 additional times (3 total attempts)
- Uses exponential backoff delays: 15s, 45s
- Only retries on timeout-related failures (Client.Timeout, connection refused, EOF)
- Does NOT retry actual test failures (assertion failures, panic, etc.)
- Logs retry attempts to the group's log file

This allows using more parallel groups (4-5) for better throughput while
gracefully recovering from transient AAP server overload.

**Timing Parser Note**

`grep` interprets patterns starting with `---` as flags. Always use
`grep -oE -e 'pattern'` (with explicit `-e`) when the pattern starts with
dashes, e.g., `grep -oE -e '--- PASS.*'`.

Save the script to `testing/run-parallel.sh` and make it executable.

Example group launch using the pre-compiled binary:

```bash
# Pre-compile once
TEST_BINARY="${LOG_DIR}/provider.test"
go test -c -o "$TEST_BINARY" ./internal/provider/

# Launch each group using the binary
TF_ACC=1 "$TEST_BINARY" \
  -test.count=1 \
  -test.timeout 30m \
  -test.run "^TestAccAAPJob|^TestAccAAPJobAction_" \
  -test.v \
  > "$LOG_DIR/jobs.log" 2>&1 &
```

### 5. Run and Validate

Execute the parallel script and verify:

- **100% pass rate** — every test must pass
- **Timing improvement** — compare wall time against sequential baseline

If any group fails, check for connection timeouts first. Timeouts usually
mean too many parallel groups, not test conflicts. Reduce group count by
merging the smallest groups together and retry.

If failures persist after reducing groups, re-run the failing group alone
to confirm whether the failure is a parallelism conflict or a pre-existing
issue.

Report results as:

```
## Test Execution Results

| Metric              | Sequential | Parallel | Speedup |
|---------------------|------------|----------|---------|
| Wall time           | 824s       | 524s     | 1.57x  |
| Tests passed        | 39/39      | 39/39    | -       |
| Groups              | 1          | 5        | -       |
```

### 6. Optimize (Only When Asked)

Only suggest code-level optimizations when the user explicitly asks:

- Add `t.Parallel()` to tests that don't share mutable state
- Split long multi-step tests into focused tests
- Use `-short` flag to skip long-running tests during development

## Key Constraints

- Never sacrifice correctness for speed. 100% pass rate is non-negotiable.
- Always pre-compile the test binary once, then run the binary per group.
- Always set `TF_ACC=1` for acceptance tests.
- Default to non-verbose output unless diagnosing failures.
- Keep parallel group count between 4-5 to avoid overwhelming the target
  infrastructure. More groups does not mean faster — connection timeouts
  from too many concurrent clients will cause all groups to fail.
- Too few groups (3 or fewer) can be slower than sequential if the
  critical-path group contains most tests. The optimal range is 4-5 groups.
- Separate job-launching tests (that exercise AAP task workers) from
  CRUD-only tests. Job tests are the primary source of contention.
- If a parallel strategy causes flaky failures, fall back to sequential for
  that group and report which tests conflicted.
