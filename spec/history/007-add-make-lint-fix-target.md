# 007 - Add `make lint-fix` Target

**Date**: 2026-03-09
**Type**: Developer Tooling

## Summary

Added a `lint-fix` Makefile target to `makefiles/golangci.mk` that runs `golangci-lint run --fix ./...` to auto-fix supported lint issues. The `fix` option is CLI-only in golangci-lint v2 and cannot be set in `.golangci.yml`.

## Change

**File**: `makefiles/golangci.mk`

Added target:
```makefile
.PHONY: lint-fix
lint-fix: lint-tools ## Run golangci-lint and auto-fix supported issues
	@echo "==> Fixing lint issues..."
	$(GOLANGCI_LINT) run --fix ./...
```

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `make lint-fix` | Runs successfully, fixes applied |
| `make help` | Shows `lint-fix` target with description |
| `make lint` | 9 remaining issues (manual fix required) |

## Remaining Issues After Autofix

| Linter | Count | Files | Required action |
|--------|------:|-------|-----------------|
| forcetypeassert | 6 | `host_resource.go`, `base_datasource_test.go`, `eda_eventstream_post_action_test.go`, `organization_data_source_test.go` | Add checked type assertions (`val, ok := x.(Type)`) |
| copyloopvar | 3 | `aapcustomstring_type_test.go`, `customstring_value_test.go` | Remove redundant loop variable copies (Go 1.22+) |

## Issue Reduction Timeline

| Stage | Issues |
|------:|-------:|
| 005 — Initial scan | 104 |
| 006 — First `--fix` run | 11 |
| 007 — `make lint-fix` | 9 |
