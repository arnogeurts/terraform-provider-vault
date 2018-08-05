---
layout: "vault"
page_title: "Vault: vault_secret_backend resource"
sidebar_current: "docs-vault-resource-secret-backend"
description: |-
  Creates an arbitrary secret backend for Vault.
---

# vault\_secret\_backend

## Example Usage

```hcl
resource "vault_secret_backend" "ssh" {
  type = "ssh"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The name of the secret backend.

* `path` - (Optional) The unique path this backend should be mounted at. Must
not begin or end with a `/`. Defaults to the type.

* `description` - (Optional) A human-friendly description for this backend.

* `default_lease_ttl_seconds` - (Optional) The default TTL for credentials
issued by this backend.

* `max_lease_ttl_seconds` - (Optional) The maximum TTL that can be requested
for credentials issued by this backend.

## Attributes Reference

No additional attributes are exported by this resource.
