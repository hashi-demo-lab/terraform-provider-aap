# Future: API Endpoint Caching for Test Speed

**Status:** Proposed
**Date:** 2026-03-10
**Related:** spec/history/014-acceptance-test-speed-optimization.md

## Problem

Every `NewClient()` call triggers `setAPIEndpoint()` → `readAPIEndpoint()`, making 1-3 HTTP GET requests (`/api/`, `/api/controller/`, `/api/eda/`) to discover static API version paths. With ~80-120 provider initializations per test suite run, this adds 240-360 redundant HTTP round-trips.

Under parallel test load, this causes 2.75x duration inflation (test duration sum of 2263s vs 824s sequential).

## Proposed Changes

### 1. API Endpoint Cache (`internal/provider/client.go`)

Package-level `sync.Mutex`-protected map keyed by host URL. Caches discovered endpoints after the first call per host. Subsequent `setAPIEndpoint()` calls return instantly from cache.

```go
type aapDiscoveredEndpoints struct {
    controllerEndpoint string
    edaEndpoint        string
}

var (
    endpointCacheMu sync.Mutex
    endpointCache   = map[string]*aapDiscoveredEndpoints{}
)

func clearEndpointCache() {
    endpointCacheMu.Lock()
    endpointCache = map[string]*aapDiscoveredEndpoints{}
    endpointCacheMu.Unlock()
}
```

Modified `setAPIEndpoint()` checks cache before making HTTP calls, stores results after discovery.

### 2. Test Client Cache (`internal/provider/provider_test.go`)

`sync.Once`-based cached client via `getTestClient()` that initializes the test client once and reuses it across all CheckExists/CheckDestroy helper calls.

```go
var (
    testClientOnce     sync.Once
    testClientInstance  *AAPClient
    testClientInitErr   error
)

func getTestClient() (*AAPClient, error) {
    testClientOnce.Do(func() { /* init from env vars */ })
    return testClientInstance, testClientInitErr
}
```

### 3. `t.Parallel()` on All 39 Tests

Add `t.Parallel()` as the first statement in every `TestAcc*` function. Two tests require a prerequisite fix — `TestAccAAPJobAction_basic` and `TestAccAAPWorkflowJobAction_Basic` capture `os.Stderr` (process-global mutation) which must be removed first.

## Impact

| Optimization | Wall Time | Speedup |
|-------------|-----------|---------|
| Baseline (sequential) | 824s | 1.0x |
| Parallel groups only | 524s | 1.57x |
| + API endpoint cache | 240s | 3.43x |
| + t.Parallel(2) | 146s | 5.64x |

The API endpoint cache alone accounts for a 2.18x speedup on top of parallel groups.

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Stale cache during run | None | API version paths don't change during a terraform run |
| TOCTOU race (two clients discover simultaneously) | Benign | Both write identical values; last-write-wins is safe |
| Error caching | None | Errors return early before cache write |
| Multi-host safety | Safe | Cache keyed by `HostURL` — different hosts get separate entries |
| Unit test interference | Low | `clearEndpointCache()` helper provided for test isolation |
| Production impact | None | Cache is process-scoped; each `terraform` CLI invocation starts fresh |
| Test client cache (provider_test.go) | None | Test-only code; `_test.go` never compiled into production binary |

## Prerequisites

- Remove `os.Stderr` capture from `job_launch_action_test.go` and `workflow_job_launch_action_test.go` before adding `t.Parallel()`
- Add `clearEndpointCache()` calls in any unit tests using `httptest` servers with different URLs

## Files to Modify

| File | Change |
|------|--------|
| `internal/provider/client.go` | Add `aapDiscoveredEndpoints` struct, endpoint cache, `clearEndpointCache()`, modify `setAPIEndpoint()` |
| `internal/provider/provider_test.go` | Add `getTestClient()` with `sync.Once`, modify `testMethodResourceWithParams()` |
| `internal/provider/job_launch_action_test.go` | Remove os.Stderr capture, add `t.Parallel()` |
| `internal/provider/workflow_job_launch_action_test.go` | Remove os.Stderr capture, add `t.Parallel()` |
| `internal/provider/*_test.go` (12 files) | Add `t.Parallel()` to all `TestAcc*` functions |
