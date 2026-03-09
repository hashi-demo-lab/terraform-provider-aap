# 003 - Uplift golangci-lint Config to Scaffolding Framework Standard

## Date
2026-03-09

## Summary
Aligned `.golangci.yml` with the HashiCorp `terraform-provider-scaffolding-framework` reference config while preserving local extras that exceed the baseline.

## Changes

### `.golangci.yml`

**8 linters added** (present in scaffolding, missing locally):

| Linter | Purpose |
|---|---|
| `copyloopvar` | Detects unnecessary copies of loop variables (post Go 1.22) |
| `depguard` | Blocks deprecated `terraform-plugin-sdk/v2` imports |
| `durationcheck` | Catches bad `time.Duration` multiplications |
| `forcetypeassert` | Flags unchecked type assertions |
| `godot` | Enforces periods at end of comments |
| `makezero` | Detects slices initialized with non-zero length then appended |
| `nilerr` | Catches `err != nil` checks that return `nil` |
| `predeclared` | Flags shadowed predeclared Go identifiers |

**`depguard` settings added** — two deny rules:
- `prevent_unmaintained_packages` — blocks `io/ioutil` (deprecated since Go 1.16)
- `prevent_sdk_v2` — blocks `terraform-plugin-sdk/v2` imports in non-internal/non-test files, directing to `terraform-plugin-framework` or `terraform-plugin-testing`

**`issues` section added**:
- `max-issues-per-linter: 0` — report all issues, no cap
- `max-same-issues: 0` — report all duplicate issues, no cap

**Kept as-is** (local extras beyond scaffolding baseline):
- 16 additional linters: `dogsled`, `gocheckcompilerdirectives`, `gochecknoinits`, `goconst`, `gocritic`, `gocyclo`, `goprintffuncname`, `gosec`, `govet`, `lll`, `mnd`, `nakedret`, `noctx`, `nolintlint`, `revive`, `whitespace`
- Formatters: `gofmt` + `goimports` (scaffolding only has `gofmt`)

## Verification Results
- `go build ./...` — passes clean
- `golangci-lint run ./...` — config parses, 104 issues from newly enabled linters:
  - `copyloopvar`: 3 (loop variable copies removable since Go 1.22)
  - `forcetypeassert`: 6 (unchecked type assertions)
  - `godot`: 91 (comments missing trailing period)
  - `gofmt`: 2 (pre-existing formatting in test files)
  - `gosec` G704: 2 (pre-existing SSRF taint analysis findings)
