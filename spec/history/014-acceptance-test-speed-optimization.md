# 014 — Acceptance Test Speed Optimization

**Status:** Complete
**Date:** 2026-03-10

## Summary

Optimized acceptance test execution from 824s (sequential) to 146s (parallel) — a **5.64x speedup** with 100% pass rate (39/39 tests). Combined five techniques: parallel group execution, API endpoint caching, test client caching, `t.Parallel()` within groups, and exponential backoff retry.

## Baseline

- 39 acceptance tests running sequentially via `go test -run TestAcc ./internal/provider/`
- Wall time: **824s** (~13.7 minutes)
- All tests pass, no parallelism (parallelism ratio 1.0)

## Optimization Journey

| Iteration | Configuration | Wall Time | Pass Rate | Speedup |
|-----------|--------------|-----------|-----------|---------|
| 0 | Sequential baseline | 824s | 39/39 | 1.0x |
| 1 | 10 groups (too many) | FAIL | 0/39 | — |
| 2 | 3 groups (too few) | 906s | 39/39 | 0.91x |
| 3 | 4 groups + backoff | 707s | 39/39 | 1.17x |
| 4 | 5 groups + backoff | 524s | 39/39 | 1.57x |
| 5 | 5 groups + API cache | 240s | 39/39 | 3.43x |
| **6** | **5 groups + cache + t.Parallel(2)** | **146s** | **39/39** | **5.64x** |
| 7 | 5 groups + cache + t.Parallel(3) | 128s | 38/39 | *(too aggressive)* |

## Changes

### 1. Parallel Group Execution (`testing/run-parallel.sh`)

Pre-compiles the test binary once with `go test -c`, then launches 5 groups concurrently using the compiled binary with `-test.run` patterns.

**Group definitions:**

| Group | Tests | Description |
|-------|-------|-------------|
| `job_resource` | 8 | Job CRUD tests |
| `job_action` | 6 | Job launch actions + host delete-with-retry |
| `workflow_job` | 7 | Workflow job CRUD tests |
| `crud_and_wf_action` | 7 | CRUD resources + workflow launch actions |
| `datasources_and_eda` | 11 | Read-only data sources + EDA event streams |

**Key design decisions:**
- Job-launching tests separated from CRUD/data-source tests to avoid AAP task worker saturation
- Job resource and job action tests split into separate groups (no shared mutable state)
- Workflow launch action tests moved into CRUD group for load balancing
- 3-second stagger between group launches to reduce connection spikes

### 2. Exponential Backoff Retry

The `run_group_with_retry()` wrapper retries failed groups up to 2 additional times with 15s/45s delays, but only for timeout-related failures (Client.Timeout, connection refused, EOF). Actual test assertion failures are not retried.

### 3. API Endpoint Caching (`internal/provider/client.go`)

**Problem:** Every `NewClient()` call triggered `setAPIEndpoint()` → `readAPIEndpoint()`, making 1-3 HTTP GET requests (`/api/`, `/api/controller/`, `/api/eda/`) to discover static URL paths. With ~80-120 provider initializations per test suite run, this added massive overhead.

**Solution:** Package-level cache (`sync.Mutex`-protected map keyed by host URL) stores discovered endpoints after the first call. Subsequent `setAPIEndpoint()` calls for the same host return instantly from cache.

```go
var (
    endpointCacheMu sync.Mutex
    endpointCache   = map[string]*aapDiscoveredEndpoints{}
)
```

**Impact:** Eliminated 2.75x duration inflation under parallel load (test sum dropped from 2263s to 1017s).

### 4. Test Client Caching (`internal/provider/provider_test.go`)

**Problem:** `testMethodResource()` (used in CheckExists/CheckDestroy callbacks) created a new `AAPClient` on every invocation, triggering redundant endpoint discovery and HTTP client setup.

**Solution:** `sync.Once`-based cached client via `getTestClient()` that initializes the test client once and reuses it across all helper calls.

### 5. `t.Parallel()` on All 39 Tests

Added `t.Parallel()` as the first statement in every `TestAcc*` function. Combined with `-test.parallel=2` per group, this allows 2 tests to run concurrently within each group process.

**Safety analysis:** All 39 tests are safe for parallel execution:
- Resources use `acctest.RandStringFromCharSet` for unique names
- Job templates are read-only references (by env var ID)
- No cross-test shared mutable state
- Each test's CheckDestroy only affects its own resources

**Two tests required a fix first:** `TestAccAAPJobAction_basic` and `TestAccAAPWorkflowJobAction_Basic` captured `os.Stderr` (process-global mutation). The stderr capture was removed — it was redundant since `resource.Test` already validates the action succeeded.

**Concurrency limit:** `-test.parallel=2` gives 5 groups × 2 = 10 concurrent test streams. `-test.parallel=3` (15 streams) caused AAP job failures from task worker contention.

## Key Findings

### AAP Server Bottlenecks

1. **Task worker saturation** — The primary contention source. Concurrent job launches compete for AAP execution slots. With >10 concurrent job-launching tests, jobs start failing.
2. **API endpoint discovery overhead** — 3 HTTP round-trips per provider init × ~100 inits = 300 wasted requests. Caching this alone cut wall time from 524s to 240s.
3. **Connection timeouts** — 10+ concurrent `go test` processes overwhelm the AAP server. The default 5s HTTP timeout (`DefaultTimeOut`) is too low for parallel load; increased to 60s via `AAP_TIMEOUT`.

### Group Count Sweet Spot

- **Too many (6+):** Connection timeouts from concurrent API clients
- **Too few (3):** Critical-path group becomes a bottleneck, slower than sequential
- **Optimal (4-5):** Balanced parallelism without overwhelming AAP

### Duration Inflation Under Load

| Config | Test Duration Sum | Wall Time | Inflation Factor |
|--------|------------------|-----------|-----------------|
| Sequential | 824s | 824s | 1.0x |
| 5 groups (no cache) | 2263s | 524s | 2.75x |
| 5 groups (cached) | 1017s | 240s | 1.23x |
| 5 groups (cached + t.Parallel) | 1173s | 146s | 1.42x |

Caching eliminated most of the inflation. The remaining 1.42x is from genuine AAP server load under concurrent requests.

## Files Modified

| File | Change |
|------|--------|
| `testing/run-parallel.sh` | New parallel execution script with 5 groups, backoff retry, timing parser |
| `internal/provider/client.go` | API endpoint cache (sync.Mutex map) |
| `internal/provider/provider_test.go` | Test client cache (sync.Once) |
| `internal/provider/job_launch_action_test.go` | Removed os.Stderr capture, added t.Parallel() |
| `internal/provider/workflow_job_launch_action_test.go` | Removed os.Stderr capture, added t.Parallel() |
| `internal/provider/*_test.go` (12 files) | Added t.Parallel() to all TestAcc* functions |

## Skill

The `provider-test-speed` skill (`.claude/skills/provider-test-speed/SKILL.md`) captures these patterns for reuse:
- Test discovery and timing analysis workflow
- Group classification strategy (4-5 groups, separate job tests from CRUD)
- Pre-compiled binary execution with `-test.*` flags
- Exponential backoff retry for transient failures
- Timing parser with `grep -e` fix for `---` patterns

Eval results: with-skill achieved 100% (8/8 assertions) vs 50% (4/8) without-skill.
