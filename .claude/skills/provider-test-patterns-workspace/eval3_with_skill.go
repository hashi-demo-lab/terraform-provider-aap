// Verifying That an Update Does Not Recreate a Resource
//
// When a Terraform resource is updated in-place, its ID must remain the same.
// If the ID changes, the provider destroyed and recreated the resource instead
// of performing a true update. This is a common source of bugs -- especially
// when ForceNew is inadvertently set on a mutable attribute.
//
// The terraform-plugin-testing library provides a purpose-built mechanism for
// this: statecheck.CompareValue with the compare.ValuesSame() comparer. You
// declare a CompareValue check once, then add state values to it in each step.
// After all steps run, the framework asserts that every collected value is
// identical. If the ID changed between step 1 (create) and step 2 (update),
// the test fails.
//
// Additionally, you can use plancheck.ExpectResourceAction to confirm the
// update step produces an "update" action rather than a "destroy + create"
// (replace) action. Combining both techniques gives you two layers of
// confidence:
//
//   1. The plan action is ResourceActionUpdate (not destroy/create).
//   2. The ID in state is identical before and after the update.
//
// Below is a complete test for an "example_server" resource with "name" and
// "size" attributes. The test creates the server, updates the size, and
// verifies the ID stays the same.

package example_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccExampleServer_updateDoesNotRecreate verifies that changing the "size"
// attribute on an example_server resource performs an in-place update rather
// than destroying and recreating the resource. The key assertion is that the
// resource ID remains identical across both test steps.
func TestAccExampleServer_updateDoesNotRecreate(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "example_server.test"

	// compareIDSame is a cross-step state check. We add the same attribute
	// path ("id") in each step. After all steps complete, the framework
	// asserts every collected value is identical via compare.ValuesSame().
	compareIDSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckExampleServerDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create the server with an initial size.
			{
				Config: testAccExampleServerConfig(rName, "small"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify the resource exists and has the expected attributes.
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("size"), knownvalue.StringExact("small")),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("id"), knownvalue.NotNull()),
					// Record the ID value for cross-step comparison.
					compareIDSame.AddStateValue(resourceName,
						tfjsonpath.New("id")),
				},
			},
			// Step 2: Update the size. The name stays the same.
			{
				Config: testAccExampleServerConfig(rName, "large"),
				// Plan checks run after the plan is computed but before apply.
				// Confirm the planned action is Update, not a replace.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName,
							plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify the size was actually updated.
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("size"), knownvalue.StringExact("large")),
					// The name should be unchanged.
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("name"), knownvalue.StringExact(rName)),
					// Record the ID again. After both steps, the framework
					// asserts this value equals the one from step 1.
					compareIDSame.AddStateValue(resourceName,
						tfjsonpath.New("id")),
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helper
// ---------------------------------------------------------------------------

// testAccExampleServerConfig returns an HCL configuration for an
// example_server resource with the given name and size. Numbered format verbs
// ensure safe quoting.
func testAccExampleServerConfig(name, size string) string {
	return fmt.Sprintf(`
resource "example_server" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

// ---------------------------------------------------------------------------
// Supporting test infrastructure
// ---------------------------------------------------------------------------

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("EXAMPLE_API_KEY") == "" {
		t.Fatal("EXAMPLE_API_KEY must be set for acceptance tests")
	}
}

// testAccCheckExampleServerDestroy verifies that all example_server resources
// have been destroyed after the test completes. This is a TestCheckFunc
// because the CheckDestroy field on TestCase requires that signature.
func testAccCheckExampleServerDestroy(s *terraform.State) error {
	conn := testAccAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "example_server" {
			continue
		}
		_, err := conn.GetServer(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("example_server %s still exists", rs.Primary.ID)
		}
		if !isNotFoundError(err) {
			return err
		}
	}
	return nil
}
