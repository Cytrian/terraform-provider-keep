package keep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccResourceApiKey_basic tests the basic CRUD operations
func TestAccResourceApiKey_basic(t *testing.T) {
	var apiKeyID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				// Test Create
				Config: testAccResourceApiKeyConfig_basic("test-key", "admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApiKeyExists("keep_apikey.test", &apiKeyID),
					resource.TestCheckResourceAttr("keep_apikey.test", "name", "test-key"),
					resource.TestCheckResourceAttr("keep_apikey.test", "role", "admin"),
					resource.TestCheckResourceAttrSet("keep_apikey.test", "reference_id"),
					resource.TestCheckResourceAttrSet("keep_apikey.test", "secret"),
				),
			},
			{
				// Test Update (should only refresh state)
				Config: testAccResourceApiKeyConfig_basic("test-key-updated", "viewer"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApiKeyExists("keep_apikey.test", &apiKeyID),
					// Name and role shouldn't change as updates aren't supported
					resource.TestCheckResourceAttr("keep_apikey.test", "name", "test-key"),
					resource.TestCheckResourceAttr("keep_apikey.test", "role", "admin"),
				),
			},
			/*			{
							// Test Import
							ResourceName:      "keep_apikey.test",
							ImportState:       true,
							ImportStateVerify: true,
							// Don't verify secret as it's only available on creation
							ImportStateVerifyIgnore: []string{"secret"},
						},
			*/
		},
	})
}

// TestAccResourceApiKey_disappears tests the case where the API key is deleted outside of Terraform
func TestAccResourceApiKey_disappears(t *testing.T) {
	var apiKeyID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApiKeyConfig_basic("test-key", "admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApiKeyExists("keep_apikey.test", &apiKeyID),
					testAccCheckApiKeyDisappears(&apiKeyID),
					// Should trigger recreation of the API key
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Helper function to generate test configuration
func testAccResourceApiKeyConfig_basic(name, role string) string {
	return fmt.Sprintf(`
resource "keep_apikey" "test" {
  name = "%s"
  role = "%s"
}
`, name, role)
}

// Helper function to check if API key exists
func testAccCheckApiKeyExists(resourceName string, apiKeyID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no API key ID is set")
		}

		client := testAccProvider.Meta().(*Client)
		apikeys, _, err := client.GetApiKeys()
		if err != nil {
			return fmt.Errorf("error getting API keys: %v", err)
		}

		for _, apikeyMap := range apikeys {
			apikey := apikeyMap.(map[string]interface{})
			if apikey["reference_id"].(string) == rs.Primary.ID {
				*apiKeyID = rs.Primary.ID
				return nil
			}
		}

		return fmt.Errorf("API key not found: %s", rs.Primary.ID)
	}
}

// Helper function to check if API key is destroyed
func testAccCheckApiKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "keep_apikey" {
			continue
		}

		apikeys, _, err := client.GetApiKeys()
		if err != nil {
			return fmt.Errorf("error getting API keys: %v", err)
		}

		for _, apikeyMap := range apikeys {
			apikey := apikeyMap.(map[string]interface{})
			if apikey["reference_id"].(string) == rs.Primary.ID && !apikey["is_deleted"].(bool) {
				return fmt.Errorf("API key still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

// Helper function to simulate API key deletion outside of Terraform
func testAccCheckApiKeyDisappears(apiKeyID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		_, err := client.DeleteApiKey(*apiKeyID)
		if err != nil {
			return fmt.Errorf("error deleting API key: %v", err)
		}
		return nil
	}
}
