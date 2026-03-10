---
name: github-issue-analysis
description: >-
  Analyze GitHub issues for any repository to discover themes, assess community
  demand, and produce prioritized roadmap recommendations. Use when the user
  wants to understand issue trends, group issues by theme, prioritize features,
  plan a roadmap, triage a backlog, or assess community demand for a GitHub
  project. Also use when the user mentions "issue analysis", "backlog
  prioritization", "feature demand", "issue themes", "roadmap planning", or
  asks what users are asking for most in a repo.
---

# GitHub Issue Analysis

Analyze a GitHub repository's issues to discover themes, measure demand, and
recommend roadmap priorities. Works against any public or accessible GitHub
repo using the `gh` CLI.

## Workflow

### 1. Identify the target

Confirm the repository with the user. Accept either:
- A `owner/repo` string (e.g. `hashicorp/terraform-provider-aws`)
- A GitHub URL (extract `owner/repo` from it)

Ask whether to analyze open issues only (default), closed issues, or both.
Ask whether to include pull requests or issues only (default: issues only).

### 2. Fetch issues

Use `gh` to pull issues with key metadata. Fetch in batches — repos can have
thousands of issues.

```bash
gh issue list --repo <owner/repo> --state open --limit 500 \
  --json number,title,body,labels,reactions,comments,createdAt,updatedAt,author,assignees
```

If the repo has more than 500 open issues, paginate or ask the user if they
want to cap the analysis at a manageable number.

For repos with mirrored upstream issues (like forks that track an upstream),
ask whether to include both repos and deduplicate.

### 3. Discover themes

Group issues into themes automatically based on content analysis:

1. Read every issue title and body (scan labels too — they carry signal).
2. Identify recurring topics, feature requests, bug patterns, and user needs.
3. Cluster into 5–12 themes. Fewer is better — merge small clusters.
4. Name each theme with a short, descriptive label (e.g. "New resource types",
   "Authentication improvements", "Error handling gaps").
5. Tag each issue with its theme. An issue can belong to multiple themes if
   it genuinely spans them, but prefer a single primary theme.

Themes should emerge from the data, not from a predefined taxonomy. However,
common patterns in provider/infrastructure repos include:
- New resource or data source requests
- Existing resource enhancement
- Bug fixes and regressions
- Authentication and authorization
- Documentation gaps
- Testing and reliability
- Performance and scalability
- Developer experience / UX

Use these as inspiration, not as a fixed list.

### 4. Score and prioritize

Apply the prioritization matrix from `references/prioritization-matrix.md`.

The matrix uses **adaptive scoring** — for repos with more than 30 issues,
scores are computed relative to peer themes (percentile-based) so the full
1–10 range is used. For small repos (≤30 issues), fixed thresholds apply.
This prevents score compression where every theme clusters near 9/10 on
large repos.

For each theme, compute scores across five dimensions, then produce a weighted
total. Present the scores transparently so the user can adjust weights.

Before scoring, ask the user:
- "Do you want to adjust the default priority weights, or shall I use the
  defaults?" — then show the default weights briefly.
- If the user has strategic goals (e.g. "we care most about enterprise
  adoption"), factor that into the Strategic Alignment dimension.

### 5. Generate the report

Produce a markdown report and save it to a file (default:
`<repo-name>-issue-analysis.md` in the current directory). The report
structure:

```
# <repo> Issue Analysis
## Date and scope
## Executive Summary
  - Total issues analyzed
  - Number of themes discovered
  - Top 3 themes by priority score
  - Key insight or recommendation

## Theme Overview
  Table: Theme | Issues | Demand Score | Priority Rank

## Theme Deep Dives (ordered by priority rank)
  For each theme:
  ### <Theme Name> (Rank #N)
  - Summary of what users are asking for
  - Issue count and list (number + title, linked)
  - Demand signals (reactions, comments, participants, age)
  - Priority score breakdown (each dimension)
  - Recommended action (build, defer, needs-design, quick-win, etc.)
  - Key issues to tackle first

## Prioritization Matrix
  Show the weights used and per-theme scores so the user can verify or adjust

## Demand Trends
  - Issues by age bracket (last 30d, 90d, 6m, 1y, older)
  - Most reacted-to issues (top 10)
  - Most commented issues (top 10)
  - Recently active themes vs stale themes

## Appendix: All Issues by Theme
  Full list grouped by theme with number, title, reactions, comments, created date
```

### 6. Discuss and refine

After presenting the report, offer to:
- Adjust priority weights and regenerate scores
- Drill into a specific theme
- Compare two themes head-to-head
- Export a simplified roadmap (theme + recommended action + target quarter)
- Analyze a different repo or the upstream counterpart

## Load References On Demand

- Read `references/prioritization-matrix.md` for the scoring dimensions,
  weights, and formulas. Load it before step 4.
