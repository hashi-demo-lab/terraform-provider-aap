---
name: golangci-lint-uplift
description: Uplift golangci-lint config to HashiCorp terraform-provider-scaffolding-framework standard. Additive only — adds missing linters/settings/sections, never removes existing config.
user-invocable: true
argument-hint: "[path-to-golangci-yml] - Defaults to .golangci.yml in the repo root"
---

# Skill: Uplift golangci-lint Config to Scaffolding Framework Standard

## Overview

Additive uplift of a project's `.golangci.yml` to the HashiCorp `terraform-provider-scaffolding-framework` baseline. Adds missing linters, settings, and sections from the scaffolding standard. **Never removes or overwrites existing config** — local extras that exceed the baseline are preserved as-is.

## Canonical Reference

**Source of truth**: https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.golangci.yml

Always fetch the latest scaffolding config before performing the uplift:

```bash
gh api repos/hashicorp/terraform-provider-scaffolding-framework/contents/.golangci.yml --jq '.content' | base64 -d
```

## Scaffolding Baseline (snapshot)

The scaffolding `.golangci.yml` as of the last update:

```yaml
version: "2"
linters:
  default: none
  enable:
    - copyloopvar
    - depguard
    - durationcheck
    - errcheck
    - forcetypeassert
    - godot
    - ineffassign
    - makezero
    - misspell
    - nilerr
    - predeclared
    - staticcheck
    - unconvert
    - unparam
    - unused
    - usetesting
  exclusions:
    generated: lax
    presets:
      - comments
      - common-false-positives
      - legacy
      - std-error-handling
    paths:
      - third_party$
      - builtin$
      - examples$
  settings:
    depguard:
      rules:
        main:
          list-mode: lax
          deny:
            - pkg: "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
              desc: "Use github.com/hashicorp/terraform-plugin-testing/helper/acctest"
            - pkg: "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
              desc: "Use github.com/hashicorp/terraform-plugin-testing/helper/resource"
            - pkg: "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
              desc: "Use github.com/hashicorp/terraform-plugin-testing/terraform"
            - pkg: "github.com/hashicorp/terraform-plugin-sdk/v2"
              desc: "Use Terraform Plugin Framework"
issues:
  max-issues-per-linter: 0
  max-same-issues: 0
formatters:
  enable:
    - gofmt
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
```

## Core Principle: Additive Only

This skill is **strictly additive**. The local config is always treated as the authority:

| Scenario | Action |
|---|---|
| Linter in baseline, **missing** locally | **Add** to `linters.enable` |
| Linter in local, **not** in baseline | **Keep** — local is stricter |
| Linter in **both** | **Keep local** — do not change |
| Setting in baseline, **missing** locally | **Add** the setting block |
| Setting in **both** | **Keep local** — do not overwrite |
| Section in baseline, **missing** locally | **Add** the section |
| Formatter in baseline, **missing** locally | **Add** the formatter |
| Formatter in local, **not** in baseline | **Keep** — local is a superset |
| Exclusion preset in baseline, **missing** locally | **Add** the preset |
| Exclusion preset in local, **not** in baseline | **Keep** |

**Never remove, replace, or overwrite** any existing linter, setting, formatter, exclusion, or section.

## Execution Steps

### 1. Fetch the Latest Baseline

Fetch the current scaffolding config from GitHub:

```bash
gh api repos/hashicorp/terraform-provider-scaffolding-framework/contents/.golangci.yml --jq '.content' | base64 -d
```

If the fetch fails, fall back to the snapshot embedded in this skill.

### 2. Read the Target Config

Read `.golangci.yml` (or the path provided as argument). If the file doesn't exist, inform the user and stop.

### 3. Delta Analysis

Compare the local config against the fetched baseline. Produce a delta report with three categories:

**To add** (in baseline, missing locally):
- List each missing linter with its purpose
- List each missing setting block
- List each missing section

**Already present** (in both):
- List shared linters/settings (no action needed)

**Local extras** (in local, not in baseline):
- List extra linters/settings/formatters that will be preserved

Present this delta to the user before making changes.

### 4. Apply Changes

1. **Add missing linters** to `linters.enable` — insert alphabetically into the existing list.
2. **Add missing settings** under `linters.settings` — add new blocks only; do not touch existing settings.
3. **Add missing sections** (e.g., `issues`) at the top-level YAML structure.
4. **Add missing exclusion presets** to `linters.exclusions.presets`.
5. **Add missing formatters** to `formatters.enable`.
6. **Merge `depguard` deny rules** — if local already has `depguard` settings, merge any missing deny rules from the baseline into the existing rule set. Do not duplicate rules that already exist. Do not remove local-only rules.

### 5. Validate

Run both commands and report results:

```bash
go build ./...
```

For linting, use whichever is available:
```bash
golangci-lint run ./...
# or if not on PATH:
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
```

- Config must parse without errors.
- New linter findings are expected — report as a summary (count per linter).
- Build must pass with no regressions.

### 6. Document

Create a spec history file at `spec/history/NNN-uplift-golangci-lint-config.md` documenting:
- Date of uplift
- The delta analysis (what was added, what was kept)
- Verification results (build status, lint issue counts by linter)

Use the next available number prefix (`NNN`).

## Key Rules

- **Additive only** — never remove existing linters, settings, formatters, or exclusions.
- **Alphabetical order** — maintain alphabetical ordering in `linters.enable`.
- **YAML v2 format** — the config uses `version: "2"` (golangci-lint v2 schema).
- **Fetch before applying** — always get the latest scaffolding config; don't rely solely on the snapshot.
- **Merge, don't replace** — for compound settings like `depguard` rules, merge missing entries into existing blocks.
- **Local settings win** — if both baseline and local define the same setting with different values, keep the local value.

## Success Criteria

- [ ] Latest scaffolding config fetched from GitHub
- [ ] Delta analysis presented to user
- [ ] All baseline linters present in `linters.enable`
- [ ] All baseline `depguard` deny rules present (merged, not replaced)
- [ ] `issues` section present with uncapped reporting
- [ ] All baseline exclusion presets present
- [ ] `go build ./...` passes
- [ ] `golangci-lint run ./...` parses config without errors
- [ ] Local extras preserved (no linters/settings/formatters/exclusions removed)
- [ ] Spec history file created

## Related Skills

- `provider-resources` — Provider resource implementation (benefits from stricter linting)
- `provider-test-patterns` — Test patterns (benefits from `copyloopvar`, `forcetypeassert`)
