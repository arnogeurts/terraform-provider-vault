package vault

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/vault/api"
	"strings"
)

func TestAccSecretBackend_basic(t *testing.T) {
	path := "ssh-" + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckSecretBackendDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretBackendConfig_basic(path),
				Check:  testAccSecretBackendCheck_basic(path),
			},
		},
	})
}

func TestAccSecretBackend_import(t *testing.T) {
	path := "ssh-" + acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		Providers: testProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccSecretBackendConfig_basic(path),
				Check:  testAccSecretBackendCheck_basic(path),
			},
			{
				ResourceName:      "vault_secret_backend.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecretBackend_updated(t *testing.T) {
	path := "ssh-" + acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		Providers: testProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccSecretBackendConfig_basic(path),
				Check:  testAccSecretBackendCheck_basic(path),
			},
			{
				Config: testAccSecretBackendConfig_updated(path),
				Check:  testAccSecretBackendCheck_updated(path),
			},
		},
	})
}

func testAccCheckSecretBackendDestroy(s *terraform.State) error {
	client := testProvider.Meta().(*api.Client)

	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vault_secret_backend" {
			continue
		}
		for path, mount := range mounts {
			path = strings.Trim(path, "/")
			rsPath := strings.Trim(rs.Primary.Attributes["path"], "/")
			if mount.Type == "aws" && path == rsPath {
				return fmt.Errorf("Mount %q still exists", path)
			}
		}
	}
	return nil
}

func testAccSecretBackendConfig_basic(path string) string {
	return fmt.Sprintf(`
resource "vault_secret_backend" "test" {
  type = "ssh"
  path = "%s"
  description = "test description"
  default_lease_ttl_seconds = 3600
  max_lease_ttl_seconds = 86400
}`, path)
}

func testAccSecretBackendCheck_basic(path string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("vault_secret_backend.test", "path", path),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "description", "test description"),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "default_lease_ttl_seconds", "3600"),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "max_lease_ttl_seconds", "86400"),
	)
}

func testAccSecretBackendConfig_updated(path string) string {
	return fmt.Sprintf(`
resource "vault_secret_backend" "test" {
  type = "ssh"
  path = "%s"
  description = "test description"
  default_lease_ttl_seconds = 1800
  max_lease_ttl_seconds = 43200
}`, path)
}

func testAccSecretBackendCheck_updated(path string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("vault_secret_backend.test", "path", path),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "description", "test description"),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "default_lease_ttl_seconds", "1800"),
		resource.TestCheckResourceAttr("vault_secret_backend.test", "max_lease_ttl_seconds", "43200"),
	)
}
