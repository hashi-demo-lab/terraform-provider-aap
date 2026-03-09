// Verifying That a Resource Update Does Not Recreate It
// =====================================================
//
// When Terraform updates a resource in-place, the resource ID must remain
// stable. If the ID changes between steps, it means Terraform destroyed and
// recreated the resource rather than updating it — which is often undesirable.
//
// The standard approach in the Terraform Plugin Testing Framework is:
//
// 1. Use a multi-step acceptance test where Step 1 creates the resource and
//    Step 2 updates an attribute.
// 2. In Step 1, use ImportStateVerifyIdentifierAttribute or a simple
//    custom Check function to capture the resource ID into a variable.
// 3. In Step 2, apply a modified config and use a Check function that
//    compares the current ID against the saved ID from Step 1.
//
// If the IDs differ, the test fails — proving the resource was recreated
// instead of updated in place.
//
// Below is a complete, self-contained Go test file demonstrating this pattern
// for a hypothetical "example_server" resource with "name" and "size"
// attributes, built on the Terraform Plugin Framework.

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestExampleServer_updateInPlace verifies that changing the "size" attribute
// on an example_server resource performs an in-place update (the ID stays the
// same) rather than a destroy-and-recreate.
func TestExampleServer_updateInPlace(t *testing.T) {
	// resourceID stores the ID captured after the initial creation step.
	var resourceID string

	resource.Test(t, resource.TestCase{
		// ProtoV6ProviderFactories wires up Plugin Framework providers.
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the resource and capture its ID.
			{
				Config: testAccExampleServerConfig("my-server", "small"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the attributes are set correctly.
					resource.TestCheckResourceAttr(
						"example_server.test", "name", "my-server",
					),
					resource.TestCheckResourceAttr(
						"example_server.test", "size", "small",
					),
					// Capture the ID so we can compare it in the next step.
					extractResourceID("example_server.test", &resourceID),
				),
			},
			// Step 2: Update the "size" attribute and verify the ID has NOT changed.
			{
				Config: testAccExampleServerConfig("my-server", "large"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the size attribute was updated.
					resource.TestCheckResourceAttr(
						"example_server.test", "size", "large",
					),
					// The name should remain unchanged.
					resource.TestCheckResourceAttr(
						"example_server.test", "name", "my-server",
					),
					// Confirm the resource was updated in place by comparing IDs.
					verifyResourceIDUnchanged("example_server.test", &resourceID),
				),
			},
		},
	})
}

// extractResourceID is a TestCheckFunc that reads the "id" attribute from the
// Terraform state for the given resource address and stores it in the provided
// pointer. This is called after Step 1 to capture the initial ID.
func extractResourceID(resourceAddr string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceAddr)
		}

		if rs.Primary == nil {
			return fmt.Errorf("resource %s has no primary instance", resourceAddr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has an empty ID", resourceAddr)
		}

		*id = rs.Primary.ID
		return nil
	}
}

// verifyResourceIDUnchanged is a TestCheckFunc that compares the current "id"
// attribute of the resource against a previously captured ID. If they differ,
// the resource was destroyed and recreated rather than updated in place.
func verifyResourceIDUnchanged(resourceAddr string, previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceAddr)
		}

		if rs.Primary == nil {
			return fmt.Errorf("resource %s has no primary instance", resourceAddr)
		}

		currentID := rs.Primary.ID
		if currentID == "" {
			return fmt.Errorf("resource %s has an empty ID after update", resourceAddr)
		}

		if currentID != *previousID {
			return fmt.Errorf(
				"resource %s was recreated: ID changed from %q to %q — expected an in-place update",
				resourceAddr, *previousID, currentID,
			)
		}

		return nil
	}
}

// testAccExampleServerConfig returns an HCL configuration string for the
// example_server resource with the given name and size.
func testAccExampleServerConfig(name, size string) string {
	return fmt.Sprintf(`
resource "example_server" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

// ---------------------------------------------------------------------------
// HOW IT WORKS
// ---------------------------------------------------------------------------
//
// The test has two steps that Terraform applies sequentially:
//
//   Step 1 — Create:
//     - Applies the config with size = "small".
//     - The Check function extractResourceID reads the ID from the Terraform
//       state and saves it into the local variable `resourceID`.
//
//   Step 2 — Update:
//     - Applies the config with size = "large" (name unchanged).
//     - Terraform diffs the state and sees only "size" changed.
//     - If the provider's schema marks "size" with RequiresReplace, Terraform
//       will destroy and recreate the resource, causing the ID to change.
//     - If "size" is updatable in place, the ID stays the same.
//     - The Check function verifyResourceIDUnchanged compares the current ID
//       to the saved one. If they differ, the test fails with a clear message.
//
// This pattern works for any resource and any attribute. Simply adjust the
// config helper and resource address. The key insight is: a stable ID across
// steps proves the update was in-place; a changed ID proves a recreate.
//
// ADDITIONAL NOTES
// ----------------
//
// - In the Plugin Framework, marking an attribute with
//   `planmodifiers.RequiresReplace()` causes Terraform to destroy and recreate
//   the resource when that attribute changes. If your test fails (ID changed),
//   check whether the attribute's schema includes RequiresReplace.
//
// - For Plugin Framework providers, use ProtoV6ProviderFactories (or
//   ProtoV5ProviderFactories for protocol v5). The older ProviderFactories
//   field is for SDKv2 providers.
//
// - The extractResourceID / verifyResourceIDUnchanged pattern is lightweight
//   and explicit. An alternative is to use terraform-plugin-testing's
//   built-in TestCheckResourceAttrPtr, but writing custom check functions
//   gives you clearer error messages when the assertion fails.
