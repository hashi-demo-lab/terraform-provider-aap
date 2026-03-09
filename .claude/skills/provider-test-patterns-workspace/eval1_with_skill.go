package example_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	tfjson "github.com/hashicorp/terraform-json"
)

// Provider factory for Protocol 6
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"example": providerserver.NewProtocol6WithError(New("test")()),
}

// -----------------------------------------------------------------------------
// PreCheck
// -----------------------------------------------------------------------------

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("EXAMPLE_API_KEY") == "" {
		t.Fatal("EXAMPLE_API_KEY must be set for acceptance tests")
	}
}

// -----------------------------------------------------------------------------
// State Resource Lookup (shared utility)
// -----------------------------------------------------------------------------

func stateResourceAtAddress(state *tfjson.State, address string) (*tfjson.StateResource, error) {
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		return nil, fmt.Errorf("no state available")
	}
	for _, r := range state.Values.RootModule.Resources {
		if r.Address == address {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found in state: %s", address)
}

// -----------------------------------------------------------------------------
// Custom StateCheck: Exists
// -----------------------------------------------------------------------------

type dnsRecordExistsCheck struct {
	resourceAddress string
}

func (e dnsRecordExistsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	r, err := stateResourceAtAddress(req.State, e.resourceAddress)
	if err != nil {
		resp.Error = err
		return
	}

	id, ok := r.AttributeValues["id"].(string)
	if !ok {
		resp.Error = fmt.Errorf("no id found for %s", e.resourceAddress)
		return
	}

	conn := testAccAPIClient()
	_, err = conn.GetDNSRecord(id)
	if err != nil {
		resp.Error = fmt.Errorf("%s not found via API: %w", e.resourceAddress, err)
		return
	}
}

func stateCheckDNSRecordExists(name string) statecheck.StateCheck {
	return dnsRecordExistsCheck{resourceAddress: name}
}

// -----------------------------------------------------------------------------
// CheckDestroy
// -----------------------------------------------------------------------------

func testAccCheckDNSRecordDestroy(s *terraform.State) error {
	conn := testAccAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "example_dns_record" {
			continue
		}
		_, err := conn.GetDNSRecord(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("dns_record %s still exists", rs.Primary.ID)
		}
		if !isNotFoundError(err) {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Test: Basic + Update
// -----------------------------------------------------------------------------

func TestAccDNSRecord_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "example_dns_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					stateCheckDNSRecordExists(resourceName),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("value"), knownvalue.StringExact("192.0.2.1")),
				},
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Test: Update
// -----------------------------------------------------------------------------

func TestAccDNSRecord_update(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "example_dns_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					stateCheckDNSRecordExists(resourceName),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("value"), knownvalue.StringExact("192.0.2.1")),
				},
			},
			{
				Config: testAccDNSRecordConfig_updated(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					stateCheckDNSRecordExists(resourceName),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("type"), knownvalue.StringExact("CNAME")),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("ttl"), knownvalue.Int64Exact(600)),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("value"), knownvalue.StringExact("example.com")),
				},
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Test: Import
// -----------------------------------------------------------------------------

func TestAccDNSRecord_import(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "example_dns_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					stateCheckDNSRecordExists(resourceName),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateKind:   resource.ImportBlockWithID,
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Config Helpers
// -----------------------------------------------------------------------------

func testAccDNSRecordConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "example_dns_record" "test" {
  name  = %[1]q
  type  = "A"
  ttl   = 300
  value = "192.0.2.1"
}
`, rName)
}

func testAccDNSRecordConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "example_dns_record" "test" {
  name  = %[1]q
  type  = "CNAME"
  ttl   = 600
  value = "example.com"
}
`, rName)
}
