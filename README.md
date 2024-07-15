# acr-cache-proxy

Azure Container Registry as an auto-authorized, pull-through Docker Hub proxy.

![Works - On My Machine](https://img.shields.io/badge/Works-On_My_Machine-2ea44f)

Features:

- Supports a cache rule with a prefix (works around an [upstream bug](https://github.com/Azure/acr/issues/599#issuecomment-2182544764
))
- Supports EntraID authentication, so works with ACR configured for public network access only (ACR basic service tier can only have public endpoints)

## Usage

Config ACR as a pull-through proxy:
```hcl
resource "azurerm_container_registry_cache_rule" "docker-io" {
  name                  = "docker-io"
  container_registry_id = azurerm_container_registry.acr.id
  target_repo           = "hub/*"
  source_repo           = "docker.io/*"
}
```

Make sure you [have Azure credentials on your device](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#readme-defaultazurecredential).

Start the server locally:
```shell
# if using a specific user-managed identity
export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
# otherwise, log in first with `az login`

acr-cache-proxy --upstream-domain example.azurecr.io --upstream-prefix hub --listen-address :8080
```
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

### Security

Docker [does not support any form of authentication on registry mirrors](https://github.com/moby/moby/issues/30880), so no authentication can be implemented. Please protect the HTTP endpoint from untrusted networks. It's better to run one instance per host, and only listen on the loopback address.

### Availability

`registry-mirrors` option is failsafe. If one mirror does not work, other mirrors and then the original endpoints will be tried. Just make sure you don't hit the annoying rate limit.

The program is mostly stateless. HA can be achieved by simply running multiple instances of it or load-balancing them with a TCP/HTTP LB.
