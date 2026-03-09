package cloudstore_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"example.com/terraform-provider-cloudstore/internal/provider"
	"example.com/terraform-provider-cloudstore/internal/client"
)

// Provider factory for Protocol 6
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cloudstore": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// PreCheck verifies required environment variables are set.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("CLOUDSTORE_API_KEY") == "" {
		t.Fatal("CLOUDSTORE_API_KEY must be set for acceptance tests")
	}
}

// testAccAPIClient returns a configured API client for test helpers.
func testAccAPIClient() *client.Client {
	return client.NewClient(os.Getenv("CLOUDSTORE_API_KEY"))
}

// ---------------------------------------------------------------------------
// State helpers
// ---------------------------------------------------------------------------

// stateResourceAtAddress finds a resource in state by its address.
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

// ---------------------------------------------------------------------------
// Custom StateCheck: Exists
// ---------------------------------------------------------------------------

type storageBucketExistsCheck struct {
	resourceAddress string
	bucket          *client.StorageBucket
}

func (e storageBucketExistsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
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
	bucket, err := conn.GetStorageBucket(id)
	if err != nil {
		resp.Error = fmt.Errorf("%s not found via API: %w", e.resourceAddress, err)
		return
	}

	if e.bucket != nil {
		*e.bucket = *bucket
	}
}

func stateCheckStorageBucketExists(name string, bucket *client.StorageBucket) statecheck.StateCheck {
	return storageBucketExistsCheck{resourceAddress: name, bucket: bucket}
}

// ---------------------------------------------------------------------------
// Custom StateCheck: Disappears
// ---------------------------------------------------------------------------

type storageBucketDisappearsCheck struct {
	resourceAddress string
}

func (e storageBucketDisappearsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	r, err := stateResourceAtAddress(req.State, e.resourceAddress)
	if err != nil {
		resp.Error = err
		return
	}

	id := r.AttributeValues["id"].(string)
	conn := testAccAPIClient()
	resp.Error = conn.DeleteStorageBucket(id)
}

func stateCheckStorageBucketDisappears(name string) statecheck.StateCheck {
	return storageBucketDisappearsCheck{resourceAddress: name}
}

// ---------------------------------------------------------------------------
// CheckDestroy
// ---------------------------------------------------------------------------

func testAccCheckStorageBucketDestroy(s *terraform.State) error {
	conn := testAccAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstore_storage_bucket" {
			continue
		}
		_, err := conn.GetStorageBucket(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("storage bucket %s still exists", rs.Primary.ID)
		}
		if !client.IsNotFoundError(err) {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestAccStorageBucket_disappears verifies that the provider handles external
// deletion of a storage bucket gracefully and plans recreation.
func TestAccStorageBucket_disappears(t *testing.T) {
	var bucket client.StorageBucket
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cloudstore_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					stateCheckStorageBucketExists(resourceName, &bucket),
					stateCheckStorageBucketDisappears(resourceName),
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccStorageBucket_invalidName validates that an empty bucket name
// produces an appropriate error during planning.
func TestAccStorageBucket_invalidName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStorageBucketConfig_invalidName(""),
				ExpectError: regexp.MustCompile(`name must not be empty`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccStorageBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "cloudstore_storage_bucket" "test" {
  name       = %[1]q
  versioning = true

  lifecycle_rules {
    action = "Delete"
    days   = 30
  }
}
`, rName)
}

func testAccStorageBucketConfig_invalidName(name string) string {
	return fmt.Sprintf(`
resource "cloudstore_storage_bucket" "test" {
  name       = %[1]q
  versioning = false
}
`, name)
}
