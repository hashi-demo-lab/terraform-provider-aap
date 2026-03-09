package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories configures the provider for acceptance testing
// using Protocol version 6 with the Plugin Framework.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"example": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables or preconditions
// are met before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	// Add any required environment variable checks here, e.g.:
	// if v := os.Getenv("EXAMPLE_API_KEY"); v == "" {
	// 	t.Fatal("EXAMPLE_API_KEY must be set for acceptance tests")
	// }
}

// TestAccDNSRecord_basic tests the creation of a dns_record resource with
// all required attributes and verifies the computed and configured values.
func TestAccDNSRecord_basic(t *testing.T) {
	rName := fmt.Sprintf("test-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSRecordConfig_basic(rName, "A", 300, "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists("example_dns_record.test"),
					resource.TestCheckResourceAttr("example_dns_record.test", "name", rName),
					resource.TestCheckResourceAttr("example_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("example_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("example_dns_record.test", "value", "192.168.1.1"),
					resource.TestCheckResourceAttrSet("example_dns_record.test", "id"),
				),
			},
		},
	})
}

// TestAccDNSRecord_update tests that a dns_record resource can be updated
// in-place, verifying that attribute changes are applied correctly.
func TestAccDNSRecord_update(t *testing.T) {
	rName := fmt.Sprintf("test-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccDNSRecordConfig_basic(rName, "A", 300, "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists("example_dns_record.test"),
					resource.TestCheckResourceAttr("example_dns_record.test", "name", rName),
					resource.TestCheckResourceAttr("example_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("example_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("example_dns_record.test", "value", "192.168.1.1"),
				),
			},
			// Update the TTL and value
			{
				Config: testAccDNSRecordConfig_basic(rName, "A", 600, "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists("example_dns_record.test"),
					resource.TestCheckResourceAttr("example_dns_record.test", "name", rName),
					resource.TestCheckResourceAttr("example_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("example_dns_record.test", "ttl", "600"),
					resource.TestCheckResourceAttr("example_dns_record.test", "value", "10.0.0.1"),
				),
			},
			// Update the record type and value (e.g., change to CNAME)
			{
				Config: testAccDNSRecordConfig_basic(rName, "CNAME", 3600, "alias.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists("example_dns_record.test"),
					resource.TestCheckResourceAttr("example_dns_record.test", "name", rName),
					resource.TestCheckResourceAttr("example_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("example_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("example_dns_record.test", "value", "alias.example.com"),
				),
			},
		},
	})
}

// TestAccDNSRecord_import tests that an existing dns_record resource can be
// imported into Terraform state using its resource ID.
func TestAccDNSRecord_import(t *testing.T) {
	rName := fmt.Sprintf("test-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	resourceName := "example_dns_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			// Create the resource first
			{
				Config: testAccDNSRecordConfig_basic(rName, "A", 300, "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(resourceName),
				),
			},
			// Import the resource and verify all attributes match
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckDNSRecordExists verifies that a dns_record resource exists
// in the Terraform state and can be retrieved from the API.
func testAccCheckDNSRecordExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for resource: %s", resourceName)
		}

		// In a real implementation, you would use the provider's API client
		// to verify the resource exists in the upstream service:
		//
		// client := testAccProvider.Meta().(*exampleClient)
		// _, err := client.GetDNSRecord(rs.Primary.ID)
		// if err != nil {
		//     return fmt.Errorf("error fetching dns_record (%s): %w", rs.Primary.ID, err)
		// }

		return nil
	}
}

// testAccCheckDNSRecordDestroy verifies that all dns_record resources have been
// destroyed after a test completes, ensuring proper cleanup.
func testAccCheckDNSRecordDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "example_dns_record" {
			continue
		}

		// In a real implementation, you would use the provider's API client
		// to verify the resource no longer exists:
		//
		// client := testAccProvider.Meta().(*exampleClient)
		// _, err := client.GetDNSRecord(rs.Primary.ID)
		// if err == nil {
		//     return fmt.Errorf("dns_record %s still exists", rs.Primary.ID)
		// }
		// Verify the error is a 404 / not found, not an unexpected error.
	}

	return nil
}

// testAccDNSRecordConfig_basic returns a Terraform configuration string for
// an example_dns_record resource with the given attributes.
func testAccDNSRecordConfig_basic(name, recordType string, ttl int, value string) string {
	return fmt.Sprintf(`
resource "example_dns_record" "test" {
  name  = %[1]q
  type  = %[2]q
  ttl   = %[3]d
  value = %[4]q
}
`, name, recordType, ttl, value)
}
