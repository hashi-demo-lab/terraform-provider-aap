# Eval Report: provider-test-patterns Skill

## Summary

| Eval | Prompt | With Skill | Without Skill | Delta |
|------|--------|-----------|--------------|-------|
| 1 | Full test file (dns_record) | **14/14 (100%)** | 6/14 (43%) | **+57pp** |
| 2 | Disappears + error (storage_bucket) | **10/10 (100%)** | 7/10 (70%) | **+30pp** |
| 3 | CompareValue cross-step ID | **7/7 (100%)** | 1/7 (14%) | **+86pp** |
| **Total** | | **31/31 (100%)** | **14/31 (45%)** | **+55pp** |

## Eval 1: Full Test File Generation (dns_record)

### Assertions (14 total)

| # | Assertion | With Skill | Without Skill |
|---|-----------|-----------|--------------|
| 1 | Uses `resource.ParallelTest` | PASS | FAIL — uses `resource.Test` |
| 2 | Uses `ProtoV6ProviderFactories` | PASS | PASS |
| 3 | Uses `ConfigStateChecks` with `statecheck.StateCheck` | PASS | FAIL — uses `Check` with `ComposeAggregateTestCheckFunc` |
| 4 | Custom `statecheck.StateCheck` for exists | PASS — `dnsRecordExistsCheck` struct with `CheckState` | FAIL — `TestCheckFunc` closure |
| 5 | Uses `statecheck.ExpectKnownValue` | PASS | FAIL — uses `resource.TestCheckResourceAttr` |
| 6 | Uses `knownvalue` types | PASS — `StringExact`, `Int64Exact`, `NotNull` | FAIL |
| 7 | Uses `tfjsonpath.New` | PASS | FAIL |
| 8 | Has `CheckDestroy` | PASS | PASS |
| 9 | Has `PreCheck` | PASS | PASS |
| 10 | Config helpers use numbered format verbs | PASS | PASS |
| 11 | Update test has second TestStep | PASS | PASS |
| 12 | Import uses `ImportState`, `ImportStateVerify`, `ImportStateKind` | PASS | PARTIAL — missing `ImportStateKind` |
| 13 | Uses `acctest.RandStringFromCharSet` | PASS | PASS |
| 14 | Does NOT mix `Check` and `ConfigStateChecks` | PASS | N/A — uses only `Check` |

### Key Differentiators
- **Modern vs Legacy checks**: Skill output uses exclusively `ConfigStateChecks` with `statecheck.StateCheck`; baseline uses exclusively legacy `Check` with `TestCheckFunc`
- **Custom StateCheck implementation**: Skill produces a proper struct implementing the `statecheck.StateCheck` interface with `CheckState(ctx, req, resp)`; baseline uses `func(s *terraform.State) error` closure
- **API client pattern**: Skill uses `testAccAPIClient()` (framework-agnostic); baseline comments out `testAccProvider.Meta()` (SDK pattern)
- **ParallelTest**: Skill correctly uses `ParallelTest`; baseline uses `Test`

---

## Eval 2: Targeted Scenario Patterns (storage_bucket)

### Assertions (10 total)

| # | Assertion | With Skill | Without Skill |
|---|-----------|-----------|--------------|
| 1 | Disappears uses `ConfigStateChecks` | PASS | FAIL — uses `Check` with `ComposeTestCheckFunc` |
| 2 | Custom `statecheck.StateCheck` for disappears | PASS — `storageBucketDisappearsCheck` | FAIL — `TestCheckFunc` closure |
| 3 | `ExpectNonEmptyPlan: true` | PASS | PASS |
| 4 | Exists check before disappears check | PASS | PASS |
| 5 | `ExpectError` with `regexp.MustCompile` | PASS | PASS |
| 6 | Empty string for bucket name | PASS | PASS |
| 7 | Uses `ProtoV6ProviderFactories` | PASS | PASS |
| 8 | Uses `resource.ParallelTest` | PASS | PASS |
| 9 | Config helpers use numbered format verbs | PASS | PASS |
| 10 | No legacy `TestCheckFunc` for step assertions | PASS | FAIL — uses `TestCheckFunc` |

### Key Differentiators
- **Disappears pattern architecture**: Skill produces a complete `statecheck.StateCheck` struct for disappears with `CheckState` method that reads from `tfjson.State` and calls API; baseline produces a `TestCheckFunc` closure with commented-out API call
- **Shared utilities**: Skill includes `stateResourceAtAddress` helper and `testAccAPIClient`; baseline has no shared utilities
- **API implementation**: Skill actually calls `conn.DeleteStorageBucket(id)`; baseline comments out the API call

---

## Eval 3: CompareValue Cross-Step ID Check (example_server)

### Assertions (7 total)

| # | Assertion | With Skill | Without Skill |
|---|-----------|-----------|--------------|
| 1 | Uses `statecheck.CompareValue` with `compare.ValuesSame()` | PASS | FAIL — custom `extractResourceID`/`verifyResourceIDUnchanged` |
| 2 | `CompareValue` declared outside Steps | PASS | N/A |
| 3 | `AddStateValue` in each step's `ConfigStateChecks` | PASS | FAIL — uses `Check` with custom `TestCheckFunc` |
| 4 | At least 2 test steps with different configs | PASS | PASS |
| 5 | References `tfjsonpath.New("id")` | PASS | FAIL — uses `rs.Primary.ID` |
| 6 | Uses `ConfigStateChecks` | PASS | FAIL — uses `Check` |
| 7 | Explanation mentions cross-step tracking | PASS | FAIL — doesn't mention `CompareValue` |

### Key Differentiators
- **Built-in vs manual**: Skill uses `statecheck.CompareValue(compare.ValuesSame())` (4 lines); baseline reinvents with two custom 20-line `TestCheckFunc` functions (~40 lines total)
- **Plan checks bonus**: Skill also adds `plancheck.ExpectResourceAction(ResourceActionUpdate)` for double verification; baseline has no plan-level checks
- **Knowledge gap**: Baseline mentions `TestCheckResourceAttrPtr` as an alternative but is unaware of `CompareValue` — the purpose-built mechanism

---

## Overall Analysis

### Consistent Skill Wins
1. **ConfigStateChecks over Check** — 3/3 evals, skill always produces modern pattern
2. **Custom statecheck.StateCheck implementations** — 3/3 evals, proper interface structs
3. **Built-in statecheck/plancheck usage** — ExpectKnownValue, CompareValue, ExpectResourceAction
4. **ParallelTest** — 3/3 evals (baseline missed in eval 1)
5. **ImportStateKind** — included by skill, missed by baseline

### Baseline Strengths (preserved in skill)
- Numbered format verbs in config helpers ✅
- CheckDestroy / PreCheck boilerplate ✅
- Test structure (basic/update/import separation) ✅
- acctest.RandStringFromCharSet for random names ✅

### Conclusion
The skill produces **100% assertion pass rate** across all 3 evals. The baseline passes **45%** of assertions. The primary gap is the modern `statecheck.StateCheck` interface vs legacy `TestCheckFunc` — exactly what the Phase 1 revision targeted.
