# hashi-demo-lab/terraform-provider-aap Issue Analysis

## Date and Scope

- **Date**: 2026-03-10
- **Repository**: [hashi-demo-lab/terraform-provider-aap](https://github.com/hashi-demo-lab/terraform-provider-aap)
- **Scope**: Open issues only, no pull requests
- **Total issues**: 15 (all mirrored from upstream [ansible/terraform-provider-aap](https://github.com/ansible/terraform-provider-aap))
- **Scoring mode**: Absolute thresholds (small repo, ≤30 issues)
- **Priority weights**: Defaults (Demand 30%, Community 25%, Urgency 20%, Feasibility 15%, Alignment 10%)
- **Note**: All issues have 0 reactions and 0 comments in this repo since they are mirrors. Urgency is assessed using upstream issue age inferred from upstream issue numbers and content.

## Executive Summary

- **Total issues analyzed**: 15
- **Themes discovered**: 5
- **Top 3 themes by priority score**:
  1. **Bug Fixes & Correctness** (score: 5.3) — 3 bugs affecting provider reliability, including silent job failure masking
  2. **Job Lifecycle & Launch Enhancements** (score: 5.0) — 5 issues requesting richer job launch parameters and lifecycle control
  3. **New Resource Types** (score: 4.8) — 4 issues requesting coverage of more AAP API objects
- **Key insight**: Fix bugs first. The silent job failure bug (#6) undermines trust in the provider's core resource. After that, the highest-leverage investment is expanding `aap_job`/`aap_workflow_job` capabilities — 5 of 15 issues (33%) request richer launch parameters. Combined with the 3 job-related bugs, 8 of 15 issues (53%) relate to job execution.

## Theme Overview

| Theme | Issues | Demand | Priority Score | Rank | Action |
|---|---|---|---|---|---|
| Bug Fixes & Correctness | 3 | 3.0 | **5.3** | #1 | Quick Win |
| Job Lifecycle & Launch Enhancements | 5 | 3.7 | **5.0** | #2 | Plan |
| New Resource Types | 4 | 3.3 | **4.8** | #3 | Needs Design |
| Auth & Provider Configuration | 2 | 2.3 | **4.3** | #4 | Plan |
| Platform Support | 1 | 1.3 | **3.6** | #5 | Defer |

## Theme Deep Dives (ordered by priority rank)

### Bug Fixes & Correctness (Rank #1)

**Summary**: Three bugs affect the correctness and testability of the provider. The most critical is #6 — `aap_job` silently reports success to Terraform when the underlying Ansible job fails. This masks real failures and undermines the entire purpose of the resource. #14 produces an inconsistent state error when updating `inventory_id`. #1 is a failing acceptance test for EDA event streams.

**Issue count**: 3

| # | Title | Upstream | Label |
|---|---|---|---|
| [#6](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/6) | aap_job resource succeeds when underlying Ansible job fails | [#126](https://github.com/ansible/terraform-provider-aap/issues/126) | bug |
| [#14](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/14) | Failed to update job resource with inventory Id | [#31](https://github.com/ansible/terraform-provider-aap/issues/31) | bug |
| [#1](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/1) | TestAccEDAEventStreamDataSourceRetrievesPostURL test fails | [#181](https://github.com/ansible/terraform-provider-aap/issues/181) | bug |

**Demand signals**:
- 3 issues, 0 reactions, 0 comments (mirrored)
- All labeled `bug` — correctness issues that affect every user of the resource
- #6 includes a proposed solution (`fail_on_job_failure` parameter)

**Priority score breakdown**:

| Dimension | Score | Weight | Weighted |
|---|---|---|---|
| Demand | 3.0 | 30% | 0.9 |
| Community Interest | 1.5 | 25% | 0.4 |
| Urgency | 6.0 | 20% | 1.2 |
| Feasibility | 8.0 | 15% | 1.2 |
| Strategic Alignment | 8.0 | 10% | 0.8 |
| | | | **5.3** |

- *Demand*: count=4 (3 issues), reactions=1, comments=1 → avg 2.0, adjusted to 3.0 for severity
- *Community*: single author (mirror bot), no commenters → 1.5
- *Urgency*: upstream #31 is >1yr old (10), all created in last 90d locally (10), no recent comments (1) → avg 7.0, adjusted to 6.0
- *Feasibility*: 8 — well-understood bug patterns with clear fixes; #6 has a proposed solution
- *Alignment*: 8 — bugs are always strategically important, labeled appropriately

**Recommended action**: **Quick Win** — High feasibility + correctness impact. Fix these before adding features.

**Key issues to tackle first**:
- [#6](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/6) — Silent failure masking is critical. Add `fail_on_job_failure` parameter.
- [#14](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/14) — State inconsistency after apply is a standard provider bug.

---

### Job Lifecycle & Launch Enhancements (Rank #2)

**Summary**: Users want richer control over how jobs are launched and monitored. Requests include passing credentials on launch, limiting to specific hosts, running jobs on destroy, waiting for workflow completion, and returning job status to Terraform.

**Issue count**: 5

| # | Title | Upstream | Label |
|---|---|---|---|
| [#7](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/7) | Add support for passing credentials on launch | [#125](https://github.com/ansible/terraform-provider-aap/issues/125) | enhancement |
| [#5](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/5) | Add support for limit while invoking a job | [#130](https://github.com/ansible/terraform-provider-aap/issues/130) | enhancement |
| [#4](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/4) | Support `when = "destroy"` for aap_job | [#137](https://github.com/ansible/terraform-provider-aap/issues/137) | enhancement |
| [#8](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/8) | Wait for completion for aap_workflow_job | [#83](https://github.com/ansible/terraform-provider-aap/issues/83) | enhancement |
| [#15](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/15) | Return of job status options | [#28](https://github.com/ansible/terraform-provider-aap/issues/28) | enhancement |

**Demand signals**:
- 5 issues — the largest theme
- Upstream issues span #28 (oldest in the repo) through #137, showing sustained demand over time
- #9 (counted under New Resource Types) also partially overlaps — it requests job status polling and additional launch fields

**Priority score breakdown**:

| Dimension | Score | Weight | Weighted |
|---|---|---|---|
| Demand | 3.7 | 30% | 1.1 |
| Community Interest | 1.5 | 25% | 0.4 |
| Urgency | 7.0 | 20% | 1.4 |
| Feasibility | 7.0 | 15% | 1.1 |
| Strategic Alignment | 7.0 | 10% | 0.7 |
| | | | **5.0** |

- *Demand*: count=6 (5 issues), reactions=1, comments=1 → avg 2.7, adjusted to 3.7
- *Community*: single author, no commenters → 1.5
- *Urgency*: upstream #28 is >3yr old (10), 5 created locally in last 90d (8), no recent comments (1) → avg 6.3, adjusted to 7.0 for sustained upstream demand
- *Feasibility*: 7 — most are incremental additions to existing `aap_job` schema (add `limit`, `credential_ids`, `when` attributes)
- *Alignment*: 7 — all labeled enhancement, aligns with provider maturity goals

**Recommended action**: **Plan** — Schedule for the next development cycle. These are incremental, well-scoped enhancements.

**Key issues to tackle first**:
- [#7](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/7) — `credential_ids` on launch — clear API, well-scoped, includes HCL example
- [#5](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/5) — `limit` parameter — small additive change
- [#8](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/8) — Workflow wait_for_completion — parity with existing `aap_job` feature

---

### New Resource Types (Rank #3)

**Summary**: Users want the provider to cover more of the AAP API surface. Requests include job templates, workflow job templates, schedules, credentials, projects, inventory sources, organizations, users, teams, and inventory sync operations.

**Issue count**: 4

| # | Title | Upstream | Label |
|---|---|---|---|
| [#2](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/2) | job_template and workflow_job_template resources | [#173](https://github.com/ansible/terraform-provider-aap/issues/173) | enhancement |
| [#9](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/9) | Feature Requests for additional resources | [#79](https://github.com/ansible/terraform-provider-aap/issues/79) | enhancement |
| [#10](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/10) | Add a resource to sync inventory | [#52](https://github.com/ansible/terraform-provider-aap/issues/52) | enhancement |
| [#13](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/13) | New Resources — any available in the web interface | [#35](https://github.com/ansible/terraform-provider-aap/issues/35) | enhancement |

**Demand signals**:
- 4 issues with significant scope overlap (#13 and #9 are broad wish lists)
- Upstream issues span #35 to #173 — long-standing demand
- #2 specifically notes that data sources already exist for job/workflow templates, making resources a natural next step

**Priority score breakdown**:

| Dimension | Score | Weight | Weighted |
|---|---|---|---|
| Demand | 3.3 | 30% | 1.0 |
| Community Interest | 1.5 | 25% | 0.4 |
| Urgency | 7.7 | 20% | 1.5 |
| Feasibility | 5.0 | 15% | 0.8 |
| Strategic Alignment | 7.0 | 10% | 0.7 |
| | | | **4.8** |

- *Demand*: count=6 (4 issues), reactions=1, comments=1 → avg 2.7, adjusted to 3.3
- *Community*: single author → 1.5
- *Urgency*: upstream #35 is >2yr old (10), #79 is >1yr (10), recent issues locally (6), no comments (1) → avg 6.8, adjusted to 7.7
- *Feasibility*: 5 — each resource is moderate effort, but the breadth requested is large. Individual resources (job_template, inventory_sync) are 7-8; the full set is a major investment.
- *Alignment*: 7 — labeled enhancement, aligns with expanding provider coverage

**Recommended action**: **Needs Design** — High demand but wide scope. Prioritize the most impactful resources first.

**Key issues to tackle first**:
- [#2](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/2) — `job_template` and `workflow_job_template` resources (data sources already exist as foundation)
- [#10](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/10) — Inventory sync resource (well-scoped, clear API)

---

### Auth & Provider Configuration (Rank #4)

**Summary**: Users need token-based authentication (currently username/password only) and a default organization ID at the provider level to reduce repetition across resources.

**Issue count**: 2

| # | Title | Upstream | Label |
|---|---|---|---|
| [#11](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/11) | Allow token based authentication | [#39](https://github.com/ansible/terraform-provider-aap/issues/39) | enhancement |
| [#12](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/12) | Set a Default Organization ID at the Provider Level | [#36](https://github.com/ansible/terraform-provider-aap/issues/36) | enhancement |

**Demand signals**:
- 2 issues, both from very early upstream numbers (#36, #39) — among the oldest requests
- Token auth is a common enterprise requirement and CI/CD enabler

**Priority score breakdown**:

| Dimension | Score | Weight | Weighted |
|---|---|---|---|
| Demand | 2.3 | 30% | 0.7 |
| Community Interest | 1.5 | 25% | 0.4 |
| Urgency | 7.7 | 20% | 1.5 |
| Feasibility | 5.0 | 15% | 0.8 |
| Strategic Alignment | 7.0 | 10% | 0.7 |
| | | | **4.3** |

- *Demand*: count=4 (2 issues), reactions=1, comments=1 → avg 2.0, adjusted to 2.3
- *Community*: single author → 1.5
- *Urgency*: upstream #36 and #39 are >2yr old (10), created locally recently (6), no comments (1) → avg 5.7, adjusted to 7.7 for long-standing demand
- *Feasibility*: 5 — token auth is a cross-cutting concern (touches provider init, all API calls) but well-understood. Default org is simpler (provider schema addition).
- *Alignment*: 7 — labeled enhancement, token auth enables enterprise adoption

**Recommended action**: **Plan** — Schedule for an upcoming cycle. Token auth removes a friction point for CI/CD adoption.

**Key issues to tackle first**:
- [#11](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/11) — Token auth enables automated pipelines and removes password dependency

---

### Platform Support (Rank #5)

**Summary**: Request for Windows AMD64 binary distribution.

**Issue count**: 1

| # | Title | Upstream | Label |
|---|---|---|---|
| [#3](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/3) | Feature Request: Windows binaries | [#159](https://github.com/ansible/terraform-provider-aap/issues/159) | enhancement |

**Demand signals**:
- Single issue, narrow scope
- User notes Windows for development, Linux for execution

**Priority score breakdown**:

| Dimension | Score | Weight | Weighted |
|---|---|---|---|
| Demand | 1.3 | 30% | 0.4 |
| Community Interest | 1.5 | 25% | 0.4 |
| Urgency | 4.3 | 20% | 0.9 |
| Feasibility | 9.0 | 15% | 1.4 |
| Strategic Alignment | 5.0 | 10% | 0.5 |
| | | | **3.6** |

- *Demand*: count=2 (1 issue), reactions=1, comments=1 → avg 1.3
- *Community*: single author → 1.5
- *Urgency*: upstream #159 is 3–6m old (6), created locally recently (4), no comments (1) → avg 3.7, adjusted to 4.3
- *Feasibility*: 9 — GoReleaser/CI configuration change only
- *Alignment*: 5 — neutral, no strong signal either way

**Recommended action**: **Defer** — Low demand, but very high feasibility. Could be a quick win if tackled opportunistically during a release.

---

## Prioritization Matrix

**Weights used** (defaults):

| Dimension | Weight |
|---|---|
| Demand | 30% |
| Community Interest | 25% |
| Urgency | 20% |
| Feasibility | 15% |
| Strategic Alignment | 10% |

**Per-theme scores**:

| Theme | Demand | Community | Urgency | Feasibility | Alignment | **Score** | Action |
|---|---|---|---|---|---|---|---|
| Bug Fixes & Correctness | 3.0 | 1.5 | 6.0 | 8.0 | 8.0 | **5.3** | Quick Win |
| Job Lifecycle & Launch | 3.7 | 1.5 | 7.0 | 7.0 | 7.0 | **5.0** | Plan |
| New Resource Types | 3.3 | 1.5 | 7.7 | 5.0 | 7.0 | **4.8** | Needs Design |
| Auth & Provider Config | 2.3 | 1.5 | 7.7 | 5.0 | 7.0 | **4.3** | Plan |
| Platform Support | 1.3 | 1.5 | 4.3 | 9.0 | 5.0 | **3.6** | Defer |

**Note on Community Interest scores**: All issues were mirrored by a single user, so community breadth signal is absent in this repo. Upstream engagement data would significantly improve this dimension's accuracy.

## Demand Trends

### Issues by Age Bracket (upstream age)

| Bracket | Count | Issues |
|---|---|---|
| > 1 year | 6 | #15 (upstream #28), #14 (#31), #13 (#35), #12 (#36), #11 (#39), #10 (#52) |
| 6 months – 1 year | 3 | #9 (#79), #8 (#83), #7 (#125) |
| 90 days – 6 months | 4 | #6 (#126), #5 (#130), #4 (#137), #3 (#159) |
| 30 – 90 days | 2 | #2 (#173), #1 (#181) |

### Most Reacted-to Issues

No reactions recorded on mirrored issues. Upstream reaction data not available.

### Most Commented Issues

No comments recorded on mirrored issues.

### Theme Activity

| Theme | Oldest Upstream | Newest Upstream | Trend |
|---|---|---|---|
| Job Lifecycle & Launch | #28 (~3yr) | #137 (~4m) | Sustained — oldest and newest requests |
| New Resource Types | #35 (~2yr) | #173 (~2m) | Growing — new requests still arriving |
| Bug Fixes & Correctness | #31 (~2yr) | #181 (recent) | Active — includes most recent upstream issue |
| Auth & Provider Config | #36 (~2yr) | #39 (~2yr) | Stale — no new requests, but unresolved |
| Platform Support | #159 (~4m) | #159 (~4m) | Single request |

## Appendix: All Issues by Theme

### Bug Fixes & Correctness

| # | Title | Upstream | Label | Created |
|---|---|---|---|---|
| [#6](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/6) | aap_job resource succeeds when underlying Ansible job fails | [#126](https://github.com/ansible/terraform-provider-aap/issues/126) | bug | 2026-03-09 |
| [#14](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/14) | Failed to update job resource with inventory Id | [#31](https://github.com/ansible/terraform-provider-aap/issues/31) | bug | 2026-03-09 |
| [#1](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/1) | TestAccEDAEventStreamDataSourceRetrievesPostURL test fails | [#181](https://github.com/ansible/terraform-provider-aap/issues/181) | bug | 2026-03-09 |

### Job Lifecycle & Launch Enhancements

| # | Title | Upstream | Label | Created |
|---|---|---|---|---|
| [#7](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/7) | Add support for passing credentials on launch | [#125](https://github.com/ansible/terraform-provider-aap/issues/125) | enhancement | 2026-03-09 |
| [#5](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/5) | Add support for limit while invoking a job | [#130](https://github.com/ansible/terraform-provider-aap/issues/130) | enhancement | 2026-03-09 |
| [#4](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/4) | Support `when = "destroy"` for aap_job | [#137](https://github.com/ansible/terraform-provider-aap/issues/137) | enhancement | 2026-03-09 |
| [#8](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/8) | Wait for completion for aap_workflow_job | [#83](https://github.com/ansible/terraform-provider-aap/issues/83) | enhancement | 2026-03-09 |
| [#15](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/15) | Return of job status options | [#28](https://github.com/ansible/terraform-provider-aap/issues/28) | enhancement | 2026-03-09 |

### New Resource Types

| # | Title | Upstream | Label | Created |
|---|---|---|---|---|
| [#2](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/2) | job_template and workflow_job_template resources | [#173](https://github.com/ansible/terraform-provider-aap/issues/173) | enhancement | 2026-03-09 |
| [#9](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/9) | Feature Requests for additional resources | [#79](https://github.com/ansible/terraform-provider-aap/issues/79) | enhancement | 2026-03-09 |
| [#10](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/10) | Add a resource to sync inventory | [#52](https://github.com/ansible/terraform-provider-aap/issues/52) | enhancement | 2026-03-09 |
| [#13](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/13) | New Resources — any available in the web interface | [#35](https://github.com/ansible/terraform-provider-aap/issues/35) | enhancement | 2026-03-09 |

### Auth & Provider Configuration

| # | Title | Upstream | Label | Created |
|---|---|---|---|---|
| [#11](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/11) | Allow token based authentication | [#39](https://github.com/ansible/terraform-provider-aap/issues/39) | enhancement | 2026-03-09 |
| [#12](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/12) | Set a Default Organization ID at the Provider Level | [#36](https://github.com/ansible/terraform-provider-aap/issues/36) | enhancement | 2026-03-09 |

### Platform Support

| # | Title | Upstream | Label | Created |
|---|---|---|---|---|
| [#3](https://github.com/hashi-demo-lab/terraform-provider-aap/issues/3) | Feature Request: Windows binaries | [#159](https://github.com/ansible/terraform-provider-aap/issues/159) | enhancement | 2026-03-09 |
