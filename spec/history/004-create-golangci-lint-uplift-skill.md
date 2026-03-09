# 004 - Create golangci-lint-uplift Skill

## Date
2026-03-09

## Summary
Created a new Claude Code skill (`/golangci-lint-uplift`) that automates uplifting a project's `.golangci.yml` to the HashiCorp `terraform-provider-scaffolding-framework` standard. The skill is strictly additive — it adds missing linters, settings, and sections without removing or overwriting existing config.

## What Was Created

### File: `.claude/skills/golangci-lint-uplift/SKILL.md`

A user-invocable skill with the following capabilities:

| Capability | Detail |
|---|---|
| **Fetch baseline** | Pulls the latest scaffolding `.golangci.yml` from GitHub via `gh api` |
| **Delta analysis** | Compares local config against baseline, categorises as add/keep/local-extra |
| **Additive apply** | Adds missing linters, settings, sections; merges `depguard` deny rules |
| **Validate** | Runs `go build` and `golangci-lint run` to verify config and report findings |
| **Document** | Creates a spec history file with delta and verification results |

### Key Design Decisions

- **Additive only** — core principle; never removes or overwrites existing linters, settings, formatters, or exclusions
- **Live fetch** — always fetches the latest scaffolding config before applying; embedded snapshot serves as fallback
- **Merge semantics** — compound settings like `depguard` rules are merged (missing entries added to existing blocks), not replaced
- **User confirmation** — delta report is presented before changes are applied
- **Alphabetical ordering** — linters inserted in alphabetical order in `linters.enable`

### Invocation
```
/golangci-lint-uplift
/golangci-lint-uplift path/to/.golangci.yml
```

## Motivation
The golangci-lint uplift performed in [003](./003-uplift-golangci-lint-config.md) was a manual process. This skill codifies that process so it can be repeated across projects or re-run when the scaffolding baseline evolves.
