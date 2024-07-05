# acr-cache-proxy

Azure Container Registry as an auto-authorized, pull-through Docker Hub proxy.

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

Make sure [you have Azure credentials on your device](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#readme-defaultazurecredential).

Start the server locally:
```shell
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
