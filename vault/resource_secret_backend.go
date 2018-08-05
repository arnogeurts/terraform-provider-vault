package vault

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/vault/api"
)

func secretBackendResource() *schema.Resource {
	return &schema.Resource{
		Create: secretBackendCreate,
		Read:   secretBackendRead,
		Update: secretBackendUpdate,
		Delete: secretBackendDelete,
		Exists: secretBackendExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the secret backend",
			},
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Path to mount the backend at, defaults to the type",
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(string)
					if strings.HasSuffix(value, "/") {
						errs = append(errs, fmt.Errorf("path cannot end in '/'"))
					}
					return
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old+"/" == new || new+"/" == old
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Human-friendly description of the mount for the backend.",
			},
			"default_lease_ttl_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Default lease duration for secrets in seconds.",
			},
			"max_lease_ttl_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum possible lease duration for secrets in seconds.",
			},
		},
	}
}

func secretBackendCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	name := d.Get("type").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	log.Printf("[DEBUG] Mounting %q backend at %q", name, path)
	err := client.Sys().Mount(path, &api.MountInput{
		Type:        name,
		Description: description,
		Config:      secretBackendMountConfigInput(d),
	})
	if err != nil {
		return fmt.Errorf("Error mounting %q secret backend to %q: %s", name, path, err)
	}
	log.Printf("[DEBUG] Mounted %q secret backend at %q", name, path)
	d.SetId(path)

	return secretBackendRead(d, meta)
}

func secretBackendRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()

	log.Printf("[DEBUG] Reading backend mount %q from Vault", path)
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("Error reading mount %q: %s", path, err)
	}
	log.Printf("[DEBUG] Read backend mount %q from Vault", path)

	// the API always returns the path with a trailing slash, so let's make
	// sure we always specify it as a trailing slash.
	mount, ok := mounts[strings.Trim(path, "/")+"/"]
	if !ok {
		log.Printf("[WARN] Mount %q not found, removing backend from state.", path)
		d.SetId("")
		return nil
	}

	d.Set("path", path)
	d.Set("type", mount.Type)
	d.Set("description", mount.Description)
	d.Set("default_lease_ttl_seconds", mount.Config.DefaultLeaseTTL)
	d.Set("max_lease_ttl_seconds", mount.Config.MaxLeaseTTL)

	return nil
}

func secretBackendUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()
	log.Printf("[DEBUG] Updating lease TTLs for %q", path)
	err := client.Sys().TuneMount(path, secretBackendMountConfigInput(d))
	if err != nil {
		return fmt.Errorf("Error updating mount TTLs for %q: %s", path, err)
	}
	log.Printf("[DEBUG] Updated lease TTLs for %q", path)

	return secretBackendRead(d, meta)
}

func secretBackendDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()

	log.Printf("[DEBUG] Unmounting secret backend %q", path)
	err := client.Sys().Unmount(path)
	if err != nil {
		return fmt.Errorf("Error unmounting secret backend from %q: %s", path, err)
	}
	log.Printf("[DEBUG] Unmounted secret backend %q", path)
	return nil
}

func secretBackendExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*api.Client)
	path := d.Id()
	log.Printf("[DEBUG] Checking if secret backend exists at %q", path)
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return true, fmt.Errorf("Error retrieving list of mounts: %s", err)
	}
	log.Printf("[DEBUG] Checked if secret backend exists at %q", path)
	_, ok := mounts[strings.Trim(path, "/")+"/"]

	return ok, nil
}

func secretBackendMountConfigInput(d *schema.ResourceData) api.MountConfigInput {
	return api.MountConfigInput{
		DefaultLeaseTTL: fmt.Sprintf("%ds", d.Get("default_lease_ttl_seconds")),
		MaxLeaseTTL:     fmt.Sprintf("%ds", d.Get("max_lease_ttl_seconds")),
	}
}
