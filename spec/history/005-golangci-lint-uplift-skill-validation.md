# 005 - golangci-lint Uplift Skill Validation

**Date**: 2026-03-09
**Type**: Validation
**Skill**: `/golangci-lint-uplift`

## Summary

Ran the `/golangci-lint-uplift` skill against the existing `.golangci.yml` to validate that the skill works correctly after its creation in 004. The config had already been uplifted manually in 003.

## Delta Analysis

### Linters (16/16 baseline linters already present)

All baseline linters were already present: `copyloopvar`, `depguard`, `durationcheck`, `errcheck`, `forcetypeassert`, `godot`, `ineffassign`, `makezero`, `misspell`, `nilerr`, `predeclared`, `staticcheck`, `unconvert`, `unparam`, `unused`, `usetesting`.

### Local Extras Preserved (16 linters)

`dogsled`, `gocheckcompilerdirectives`, `gochecknoinits`, `goconst`, `gocritic`, `gocyclo`, `goprintffuncname`, `gosec`, `govet`, `lll`, `mnd`, `nakedret`, `noctx`, `nolintlint`, `revive`, `whitespace`.

### Sections

- Exclusion presets: All 4 present (`comments`, `common-false-positives`, `legacy`, `std-error-handling`)
- Exclusion paths: All 3 present (`third_party$`, `builtin$`, `examples$`)
- Issues section: Present with uncapped reporting
- Formatters: `gofmt` present, local extra `goimports` preserved

### Depguard Deny Rules

| Baseline `pkg` | Status |
|----------------|--------|
| `github.com/hashicorp/terraform-plugin-sdk/v2` | Already present |
| `github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource` | Already present |
| `github.com/hashicorp/terraform-plugin-sdk/v2/terraform` | Already present |
| `github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest` | **Added** |

Local extras preserved: `io/ioutil` deny, `helper/schema` deny, file-scoped restrictions.

## Changes Applied

1. Added missing `acctest` deny rule to `prevent_sdk_v2` depguard rule group

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `golangci-lint run ./...` | Config valid, 104 pre-existing findings |

### Lint Findings by Linter (all pre-existing)

| Linter | Count |
|--------|-------|
| godot | 91 |
| forcetypeassert | 6 |
| copyloopvar | 3 |
| gofmt | 2 |
| gosec | 2 |
| **Total** | **104** |
