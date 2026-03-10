# 013 — Playbook Async Optimization

**Status:** Complete
**Date:** 2026-03-10

## Summary

Optimized `testing/playbook.yml` by converting sequential API calls to async parallel execution using Ansible's `async`/`poll` pattern. The original playbook ran 20+ AAP API calls sequentially; the optimized version fires independent tasks concurrently and only serializes where dependencies exist.

## Changes

### 1. Async Task Execution (4-Phase Pipeline)

Restructured the playbook from a flat sequential list into a 4-phase dependency-aware pipeline:

| Phase | Description | Tasks |
|---|---|---|
| **Phase 1** | Fire all independent tasks asynchronously | Project sync (git clone), orgs, credential lookup, instance group lookup, label, workflow templates, inventory, EDA credential/event stream |
| **Phase 2** | Collect Phase 1 results, fire dependent tasks | Job templates (depend on Demo Project/Credential), workflow with inventory (depends on inventory creation) |
| **Phase 3** | Wait for project sync + Phase 2, fire remaining | Sleep and fail job templates (depend on Test Playbooks project sync) |
| **Phase 4** | Wait for Phase 3, sequential final tasks | Workflow node wiring, token creation, env file generation |

### 2. Key Optimizations

- **`gather_facts: false`** — Skips fact gathering since no host facts are needed (localhost-only playbook)
- **Project sync started first** — The git clone of `ansible/test-playbooks` is the slowest operation; it now runs in the background from the start while other tasks execute
- **EDA tasks run during Phase 1 waits** — The sequential `block/rescue` EDA tasks (credential + event stream) execute while async tasks complete in the background
- **`async: 300` / `poll: 0`** on all independent tasks — Fire-and-forget pattern with `async_status` polling to collect results
- **`changed_when: false`** on lookup tasks — Credential and instance group lookups are idempotent reads, suppressing false "changed" reports

### 3. Dependency Graph

```
Phase 1 (parallel):
  project_async ─────────────────────────────────────────┐
  default_org_async ──────┐                              │
  organization_non_default_async ──┐                     │
  demo_credential_async ──┤        │                     │
  default_instance_group_async ──┤ │                     │
  test_label_async ───────┤        │                     │
  workflow_job_template_async ──┤  │                     │
  workflow_job_template_failure_async ──┤                 │
  inventory_for_workflow_async ──┤                        │
  EDA credential (sequential) ──┤                        │
  EDA event stream (sequential) ──┘                      │
                                                         │
Phase 2 (parallel, after Phase 1):                       │
  job_template_async ──────────────┐                     │
  job_template_inventory_prompt_async ──┐                │
  job_template_all_fields_async ──┤    │                 │
  workflow_with_inventory_async ──┘    │                 │
                                       │                 │
Phase 3 (after Phase 2 + project sync):│                 │
  job_template_sleep_async ──┐    ◄────┘            ◄────┘
  job_template_fail_async ───┘

Phase 4 (sequential, after Phase 3):
  workflow node wiring → token creation → env file → summary
```

## Files Changed

| File | Change |
|---|---|
| `testing/playbook.yml` | Restructured from sequential to async 4-phase pipeline |
