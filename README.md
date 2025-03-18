# acr-cache-proxy

Azure Container Registry as an auto-authorized, pull-through Docker Hub proxy.

![Works - On My Machine](https://img.shields.io/badge/Works-On_My_Machine-2ea44f)
![Project Status - Feature Complete](https://img.shields.io/badge/Project_Status-Feature_Complete-2ea44f)
[![Docker Image Version](https://img.shields.io/docker/v/jamesits/acr-cache-proxy?label=Docker%20Hub)](http://hub.docker.com/r/jamesits/acr-cache-proxy)

Features:

- Supports a cache rule with a prefix (works around an [upstream bug](https://github.com/Azure/acr/issues/599#issuecomment-2182544764
))
- Supports EntraID authentication (automatic through machine identity, or manually configured), so you can use the much cheaper ACR basic service tier which cannot have private endpoints

## Usage

Config ACR as a pull-through proxy:
```hcl
resource "azurerm_key_vault" "kv" {
  name                            = ""
  location                        = ""
  resource_group_name             = ""
  tenant_id                       = ""
  sku_name                        = "standard"
  public_network_access_enabled   = true
  enable_rbac_authorization       = true
}

resource "azurerm_key_vault_secret" "dockerhub-username" {
  key_vault_id = azurerm_key_vault.kv.id
  name         = "dockerhub-username"
  value        = "your-username"
}

resource "azurerm_key_vault_secret" "dockerhub-password" {
  key_vault_id = azurerm_key_vault.kv.id
  name         = "dockerhub-password"
  value        = "your-access-token"
}

resource "azurerm_container_registry" "acr" {
  name                          = ""
  resource_group_name           = ""
  location                      = ""
  sku                           = "Premium"
  public_network_access_enabled = true
}

resource "azurerm_container_registry_credential_set" "dockerhub" {
  name                  = "dockerhub"
  container_registry_id = azurerm_container_registry.acr.id
  login_server          = "docker.io"
  identity {
    type = "SystemAssigned"
  }
  authentication_credentials {
    username_secret_id = "${azurerm_key_vault.kv.vault_uri}secrets/${azurerm_key_vault_secret.dockerhub-username.name}"
    password_secret_id = "${azurerm_key_vault.kv.vault_uri}secrets/${azurerm_key_vault_secret.dockerhub-password.name}"
  }
}

resource "azurerm_role_assignment" "kv-acr" {
  principal_id         = azurerm_container_registry_credential_set.dockerhub.identity[0].principal_id
  role_definition_name = "Key Vault Secrets User"
  scope                = azurerm_key_vault.kv.id
}

resource "azurerm_container_registry_cache_rule" "dockerhub" {
  name                  = "dockerhub"
  container_registry_id = azurerm_container_registry.acr.id
  source_repo           = "docker.io/*"
  target_repo           = "cache/dockerhub/*"
  credential_set_id     = azurerm_container_registry_credential_set.dockerhub.id
}
```
(If you use this ACR entirely for caching, use `target_repo = "*"`.)

Make sure you:
- [have Azure credentials on your device](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#readme-defaultazurecredential)
- assigned your identity at least [AcrPull](https://www.azadvertizer.net/azrolesadvertizer/7f951dda-4ed3-4680-a7ca-43fe172d538d.html) RBAC role on the ACR instance

Start the server locally:
```shell
# if using a specific user-managed identity
# swap the GUID with your identity's client ID
export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
# otherwise, log in first with `az login`

acr-cache-proxy --upstream-domain example.azurecr.io --upstream-prefix hub --listen-address 127.0.0.1:8080
```
(If you have `target_repo = "*"` setup, do not set `--upstream-prefix` here).

Or use a service manager:
- [Docker Compose](contrib/docker-compose)
- [Nomad](contrib/nomad)

Config your Docker daemon to use the mirror in `/etc/docker/daemon.json`:
```json
{
    "registry-mirrors": ["http://localhost:8080"]
}
```

Restart your Docker daemon and profit.

### Command Line Arguments

- `--upstream-domain` (required): your Azure Container Registry domain
- `--upstream-prefix` (optional): the cache rule prefix (without `/`)
- `--listen-address` (optional): HTTP proxy listen address (Golang format)

## Building

With GoReleaser:

```shell
goreleaser build --single-target --snapshot --clean
```

Alternatively, using native Golang toolchain:

```shell
go build .
```

## Notes

### Feature Parity

- This program only proxies metadata. Azure CR serves container image layers through a redirection to a Azure Blob URL; this URL will not be proxied by us.
- Only [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec/blob/v1.0.1/spec.md) APIs are supported.

### Security

Docker [does not support any form of authentication on registry mirrors](https://github.com/moby/moby/issues/30880), so no authentication can be implemented. Please protect the HTTP endpoint from untrusted networks. It's recommended to run one instance per host, and only listen on the loopback address.

### Availability

Docker daemon's `registry-mirrors` ~~is failsafe~~ fails silently unless you enable debug logging. If one mirror does not work, other mirrors and then the original endpoints will be tried. Just make sure you don't hit the annoying rate limit.

The program is stateless. HA can be achieved by simply running multiple instances of it or load-balancing them with a TCP/HTTP LB.
