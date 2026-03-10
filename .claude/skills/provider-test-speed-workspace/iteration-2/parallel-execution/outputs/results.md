# Parallel Test Execution Results

## Test Execution Results

| Metric              | Sequential | Parallel | Speedup |
|---------------------|------------|----------|---------|
| Wall time           | 824s       | 533s     | 1.55x   |
| Tests passed        | 39/39      | 39/39    | -       |
| Tests failed        | 0          | 0        | -       |
| Groups              | 1          | 5        | -       |

## Parallel Group Breakdown

| Group               | Tests | Seq Duration (sum) | Parallel Duration (sum) | Wall Time   | Status |
|---------------------|-------|--------------------|-------------------------|-------------|--------|
| job_resource        | 8     | 208.47s            | 524.10s                 | ~525s       | PASS   |
| job_action          | 6     | 117.64s            | 267.51s                 | ~268s       | PASS   |
| workflow_job        | 7     | 153.06s            | 468.53s                 | ~469s       | PASS   |
| crud_and_wf_action  | 7     | 158.37s            | 481.33s                 | ~482s       | PASS   |
| datasources_and_eda | 11    | 199.80s            | 521.26s                 | ~523s       | PASS   |
| **TOTAL**           | **39**| **837.34s**        | **2262.73s**            | **533s**    | **PASS** |

## Configuration

- **AAP_TIMEOUT**: Increased from default 5s to 60s to prevent HTTP client timeouts under concurrent load
- **Pre-compiled binary**: Test binary compiled once before launching groups to avoid 5 parallel `go test -c` compilations
- **Staggered launches**: 3-second delay between group launches to reduce initial connection spike
- **Retry mechanism**: Groups retry up to 2 times with exponential backoff (15s, 45s) on timeout-related failures

## Analysis

### Why test durations inflate under parallelism

The AAP server (single instance) becomes a bottleneck when 5 concurrent test groups are making API calls simultaneously. Individual test durations increased by approximately 2-3x compared to sequential execution:

- `TestAccHostResourceDeleteWithRetry`: 41s (seq) -> 188s (parallel, worst case)
- `TestAccWorkflowJobTemplateDataSource`: 27s (seq) -> 88s (parallel)
- `TestAccInventoryResource`: 42s (seq) -> 111s (parallel)

Despite the per-test slowdown, the parallel wall time (533s) is still 35% faster than sequential (824s) because 5 groups run concurrently.

### Grouping rationale

Tests are grouped by resource type to avoid shared mutable state conflicts:

1. **job_resource** - All `TestAccAAPJob_*` tests (CRUD operations on job resources)
2. **job_action** - All `TestAccAAPJobAction_*` tests + `TestAccHostResourceDeleteWithRetry` (launches jobs)
3. **workflow_job** - All `TestAccAAPWorkflowJob*` tests except actions (workflow CRUD)
4. **crud_and_wf_action** - CRUD resources (host, group, inventory, org+inventory) + workflow job actions
5. **datasources_and_eda** - Read-only data source tests + EDA subsystem tests

Each test creates resources with random names (`acctest.RandStringFromCharSet`), so no naming collisions occur between groups.

### Key constraints discovered

1. **AAP HTTP timeout**: The default 5s provider timeout is insufficient under concurrent load. Set via `AAP_TIMEOUT` env var.
2. **Server capacity**: The AAP instance handles ~3-4 concurrent test groups well, but 6+ causes severe slowdown.
3. **Bash `GROUPS` variable**: Cannot use `GROUPS` as a variable name in bash scripts - it's a built-in that contains the current user's group IDs.

## Script Location

`testing/run-parallel.sh` - self-contained, immediately runnable without modification.

## How to Run

```bash
set -a
source testing/.env
source testing/acceptance_test_vars.env
set +a
./testing/run-parallel.sh
```
