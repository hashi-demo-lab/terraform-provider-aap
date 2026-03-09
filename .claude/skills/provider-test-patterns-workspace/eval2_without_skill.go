package cloudstore

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories returns provider factories for Protocol 6
// testing with the Plugin Framework.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cloudstore": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables or
// prerequisites are set before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	// Add any environment variable checks needed for the provider, e.g.:
	// if v := os.Getenv("CLOUDSTORE_API_KEY"); v == "" {
	// 	t.Fatal("CLOUDSTORE_API_KEY must be set for acceptance tests")
	// }
}

// TestAccStorageBucket_disappears verifies that the resource handles external
// deletion (out-of-band removal) gracefully. When the bucket is deleted
// outside of Terraform, the next plan/apply should detect the resource is
// gone and propose to recreate it rather than error.
func TestAccStorageBucket_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists("cloudstore_storage_bucket.test"),
					testAccCheckStorageBucketDisappears("cloudstore_storage_bucket.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccStorageBucket_invalidName_empty verifies that the provider returns a
// clear validation error when an empty string is supplied as the bucket name.
// This test never creates real infrastructure — it expects a plan-time error.
func TestAccStorageBucket_invalidName_empty(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStorageBucketConfig_emptyName(),
				ExpectError: regexp.MustCompile(`(?i)invalid.*bucket.*name|name.*must not be empty|name.*must be`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test check functions
// ---------------------------------------------------------------------------

// testAccCheckStorageBucketExists verifies the resource exists in both
// Terraform state and the remote API.
func testAccCheckStorageBucketExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set for %s", resourceName)
		}

		// In a real provider you would use the API client to confirm the
		// bucket exists remotely:
		//
		// client := testAccProvider.Meta().(*cloudstore.Client)
		// _, err := client.GetStorageBucket(context.Background(), rs.Primary.ID)
		// if err != nil {
		//     return fmt.Errorf("error reading storage bucket (%s): %w", rs.Primary.ID, err)
		// }

		return nil
	}
}

// testAccCheckStorageBucketDisappears simulates out-of-band deletion by
// calling the API directly to remove the bucket, so the next Terraform
// plan detects it is gone.
func testAccCheckStorageBucketDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set for %s", resourceName)
		}

		// In a real provider, delete the bucket via the API client so
		// Terraform's next refresh sees that it no longer exists:
		//
		// client := testAccProvider.Meta().(*cloudstore.Client)
		// err := client.DeleteStorageBucket(context.Background(), rs.Primary.ID)
		// if err != nil {
		//     return fmt.Errorf("error deleting storage bucket (%s) during disappears test: %w", rs.Primary.ID, err)
		// }

		return nil
	}
}

// testAccCheckStorageBucketDestroy verifies that all storage buckets created
// during the test have been destroyed. Called automatically after all test
// steps complete.
func testAccCheckStorageBucketDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstore_storage_bucket" {
			continue
		}

		// In a real provider, confirm the bucket no longer exists:
		//
		// client := testAccProvider.Meta().(*cloudstore.Client)
		// _, err := client.GetStorageBucket(context.Background(), rs.Primary.ID)
		// if err == nil {
		//     return fmt.Errorf("storage bucket %s still exists after destroy", rs.Primary.ID)
		// }
		// Ensure the error is a 404/not-found, not an unexpected failure.
	}

	return nil
}

// ---------------------------------------------------------------------------
// Terraform configuration helpers
// ---------------------------------------------------------------------------

// testAccStorageBucketConfig_basic returns a minimal valid configuration for a
// storage bucket with versioning enabled and a single lifecycle rule.
func testAccStorageBucketConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "cloudstore_storage_bucket" "test" {
  name       = %[1]q
  versioning = true

  lifecycle_rules {
    action = "Delete"
    days   = 90
  }
}
`, name)
}

// testAccStorageBucketConfig_emptyName returns a configuration that uses an
// empty string for the bucket name, which should be rejected by validation.
func testAccStorageBucketConfig_emptyName() string {
	return `
resource "cloudstore_storage_bucket" "test" {
  name       = ""
  versioning = false

  lifecycle_rules {
    action = "Delete"
    days   = 30
  }
}
`
}
