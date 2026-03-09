# 006 - golangci-lint Autofix

**Date**: 2026-03-09
**Type**: Code Quality

## Summary

Ran `golangci-lint run --fix ./...` to automatically fix all supported lint findings identified in 005.

## Note on Config-Based Fix

The `fix` option is **CLI-only** in golangci-lint v2 (`--fix` flag). It cannot be set in `.golangci.yml` — the `output` section in v2 only controls output format/path settings, not the fix behavior.

## Results

### Before

| Linter | Count |
|--------|------:|
| godot | 91 |
| forcetypeassert | 6 |
| copyloopvar | 3 |
| gofmt | 2 |
| gosec | 2 |
| **Total** | **104** |

### After `--fix`

| Linter | Count | Change |
|--------|------:|--------|
| forcetypeassert | 6 | No autofix support |
| copyloopvar | 3 | Reported as fixable but not auto-applied |
| gosec | 2 | No autofix support |
| **Total** | **11** | **-93 fixed** |

### Fixed Automatically (93 issues)

| Linter | Fixed | What changed |
|--------|------:|--------------|
| godot | 91 | Added missing periods to comment endings |
| gofmt | 2 | Reformatted import blocks |

### Remaining (11 issues, require manual fixes)

| Linter | Count | Required action |
|--------|------:|-----------------|
| forcetypeassert | 6 | Add checked type assertions (`val, ok := x.(Type)`) |
| copyloopvar | 3 | Remove unnecessary loop variable copies (Go 1.22+) |
| gosec (G704) | 2 | Review SSRF taint analysis in `client.go` |

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `golangci-lint run ./...` | 11 remaining issues (no regressions) |
