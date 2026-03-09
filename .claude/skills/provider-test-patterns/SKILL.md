---
name: provider-test-patterns
description: >-
  Terraform provider acceptance test patterns using terraform-plugin-testing
  with the Plugin Framework. Covers test structure, TestCase/TestStep fields,
  state checks, plan checks, config helpers, import testing, sweepers, and
  common scenario patterns. Use this skill when writing, reviewing, or
  debugging acceptance tests for a Terraform provider, including when the user
  asks about TestCheckFunc, statecheck, plancheck, import state verification,
  test sweepers, or how to structure provider test files.
---

# Provider Acceptance Test Patterns

Patterns for writing acceptance tests using
[terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing)
with the [Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

Source: [HashiCorp Testing Patterns](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns)

**References** (load when needed):
- `references/checks.md` — statecheck, plancheck, knownvalue types, tfjsonpath, comparers
- `references/sweepers.md` — sweeper setup, TestMain, dependencies

---

## Test Lifecycle

The framework runs each TestStep through: **plan → apply → refresh → final
plan**. If the final plan shows a diff, the test fails (unless
`ExpectNonEmptyPlan` is set). After all steps, destroy runs followed by
`CheckDestroy`. This means every test automatically verifies that
configurations apply cleanly and produce no drift — no assertions needed for
that.

---

## Test Function Structure

```go
func TestAccExample_basic(t *testing.T) {
    rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
    resourceName := "example_widget.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckExampleDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleConfig_basic(rName),
                ConfigStateChecks: []statecheck.StateCheck{
                    testAccCheckExampleExists(resourceName),
                    statecheck.ExpectKnownValue(resourceName,
                        tfjsonpath.New("name"), knownvalue.StringExact(rName)),
                    statecheck.ExpectKnownValue(resourceName,
                        tfjsonpath.New("id"), knownvalue.NotNull()),
                },
            },
        },
    })
}
```

Use `resource.ParallelTest` by default. Use `resource.Test` only when tests
share state or cannot run concurrently.

---

## Provider Factory

```go
// provider_test.go — Plugin Framework with Protocol 6 (use Protocol5 variant if needed)
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "example": providerserver.NewProtocol6WithError(New("test")()),
}
```

---

## TestCase Fields

| Field | Purpose |
|-------|---------|
| `PreCheck` | `func()` — verify prerequisites (env vars, API access) |
| `ProtoV6ProviderFactories` | Plugin Framework provider factories |
| `CheckDestroy` | `TestCheckFunc` — verify resources destroyed after all steps |
| `Steps` | `[]TestStep` — sequential test operations |
| `TerraformVersionChecks` | `[]tfversion.TerraformVersionCheck` — gate by CLI version |

---

## TestStep Fields

### Config Mode

| Field | Purpose |
|-------|---------|
| `Config` | Inline HCL string to apply |
| `ConfigStateChecks` | `[]statecheck.StateCheck` — modern assertions (preferred) |
| `Check` | Legacy `ComposeAggregateTestCheckFunc(...)` assertions |
| `ConfigPlanChecks` | `resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{...}}` |
| `ExpectError` | `*regexp.Regexp` — expect failure matching pattern |
| `ExpectNonEmptyPlan` | `bool` — expect non-empty plan after apply |
| `PlanOnly` | `bool` — plan without applying |
| `Destroy` | `bool` — run destroy step |
| `PreConfig` | `func()` — setup before step |

### Import Mode

| Field | Purpose |
|-------|---------|
| `ImportState` | `true` to enable import mode |
| `ImportStateVerify` | Verify imported state matches prior state |
| `ImportStateVerifyIgnore` | `[]string` — attributes to skip during verify |
| `ImportStateKind` | `resource.ImportBlockWithID` — import block generation |
| `ResourceName` | Resource address to import |
| `ImportStateId` | Override the ID used for import |

---

## Check Functions

### Modern: ConfigStateChecks (preferred)

Type-safe with aggregated error reporting. See `references/checks.md` for full
knownvalue types, tfjsonpath navigation, and comparers.

```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(resourceName,
        tfjsonpath.New("name"), knownvalue.StringExact("my-widget")),
    statecheck.ExpectKnownValue(resourceName,
        tfjsonpath.New("enabled"), knownvalue.Bool(true)),
    statecheck.ExpectKnownValue(resourceName,
        tfjsonpath.New("id"), knownvalue.NotNull()),
    statecheck.ExpectSensitiveValue(resourceName,
        tfjsonpath.New("api_key")),
},
```

### Legacy: TestCheckFunc (still common in existing code)

```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr(name, "key", "expected"),
    resource.TestCheckResourceAttrSet(name, "id"),
    resource.TestCheckNoResourceAttr(name, "removed"),
    resource.TestMatchResourceAttr(name, "url", regexp.MustCompile(`^https://`)),
    resource.TestCheckResourceAttrPair(res1, "ref_id", res2, "id"),
),
```

`ComposeAggregateTestCheckFunc` reports all errors; `ComposeTestCheckFunc`
fails fast on the first.

---

## Config Helpers

Use numbered format verbs — `%[1]q` for quoted strings, `%[1]s` for raw:

```go
func testAccExampleConfig_basic(rName string) string {
    return fmt.Sprintf(`
resource "example_widget" "test" {
  name = %[1]q
}
`, rName)
}

func testAccExampleConfig_full(rName, description string) string {
    return fmt.Sprintf(`
resource "example_widget" "test" {
  name        = %[1]q
  description = %[2]q
  enabled     = true
}
`, rName, description)
}
```

---

## Scenario Patterns

### Basic + Update (combine in one test — updates are supersets of basic)

```go
Steps: []resource.TestStep{
    {
        Config: testAccExampleConfig_basic(rName),
        ConfigStateChecks: []statecheck.StateCheck{
            testAccCheckExampleExists(resourceName),
            statecheck.ExpectKnownValue(resourceName,
                tfjsonpath.New("name"), knownvalue.StringExact(rName)),
        },
    },
    {
        Config: testAccExampleConfig_full(rName, "updated"),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue(resourceName,
                tfjsonpath.New("description"), knownvalue.StringExact("updated")),
        },
    },
},
```

### Import

After a config step, verify import produces identical state. Use
`ImportStateKind` for import block generation:

```go
{
    ResourceName:      resourceName,
    ImportState:       true,
    ImportStateVerify: true,
    ImportStateKind:   resource.ImportBlockWithID,
},
```

### Disappears (resource deleted externally)

```go
{
    Config: testAccExampleConfig_basic(rName),
    Check: resource.ComposeAggregateTestCheckFunc(
        testAccCheckExampleExists(resourceName),
        testAccCheckExampleDisappears(resourceName),
    ),
    ExpectNonEmptyPlan: true,
},
```

### Validation (expect error)

```go
{
    Config:      testAccExampleConfig_invalidName(""),
    ExpectError: regexp.MustCompile(`name must not be empty`),
},
```

### Regression (link to bug report in test name/comment)

```go
// TestAccExample_regressionGH1234 verifies fix for https://github.com/org/repo/issues/1234
func TestAccExample_regressionGH1234(t *testing.T) { ... }
```

---

## Helper Functions

### Exists Check

Separate API existence verification into a dedicated function for reuse across
steps — the source recommends this as a design principle:

```go
func testAccCheckExampleExists(name string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[name]
        if !ok {
            return fmt.Errorf("not found: %s", name)
        }
        conn := testAccProvider.Meta().(*client.Client)
        _, err := conn.GetWidget(rs.Primary.ID)
        return err
    }
}
```

### Destroy Check

```go
func testAccCheckExampleDestroy(s *terraform.State) error {
    conn := testAccProvider.Meta().(*client.Client)
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "example_widget" {
            continue
        }
        _, err := conn.GetWidget(rs.Primary.ID)
        if err == nil {
            return fmt.Errorf("widget %s still exists", rs.Primary.ID)
        }
        if !isNotFoundError(err) {
            return err
        }
    }
    return nil
}
```

### PreCheck

```go
func testAccPreCheck(t *testing.T) {
    t.Helper()
    if os.Getenv("EXAMPLE_API_KEY") == "" {
        t.Fatal("EXAMPLE_API_KEY must be set for acceptance tests")
    }
}
```

---

## Running Tests

```bash
# Single test
TF_ACC=1 go test ./internal/service/example -run TestAccExample_basic -v -timeout 60m

# Compile only (fast syntax check)
go test -c -o /dev/null ./internal/service/example

# Debug logging
TF_ACC=1 TF_LOG=debug go test ./internal/service/example -run TestAccExample_basic -v
```

For sweeper setup and execution, read `references/sweepers.md`.
