package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/jamesits/acr-cache-proxy/pkg/registry"
	"log"
	"time"
)

const (
	tokenUpdateThreshold = 300 * time.Second
	tokenEndpointPath    = "/token"
	username             = "00000000-0000-0000-0000-000000000000"
)

var (
	// config
	masqueradeDomain string
	upstreamDomain   string
	upstreamPrefix   string
	listenAddress    string

	// runtime
	upstreamRealm   string
	upstreamService string
)

func init() {
	flag.StringVar(&masqueradeDomain, "masquerade-domain", "registry.docker.io", "pretend I'm this registry")
	flag.StringVar(&upstreamDomain, "upstream-domain", "example.azurecr.io", "redirect requests to this registry")
	flag.StringVar(&upstreamPrefix, "upstream-prefix", "", "add a common prefix to the container names")
	flag.StringVar(&listenAddress, "listen-address", ":80", "HTTP listen address")
	flag.Parse()
}

func main() {
	var err error
	upstreamRealm, upstreamService, err = registry.GetAuthMetadata(upstreamDomain, "")
	if err != nil {
		log.Printf("unable to get upstream registry auth config: %v\n", err)
	}

	for err = updateToken(upstreamDomain); err != nil; err = updateToken(upstreamDomain) {
		log.Printf("initial token acquiring failed: %v\n", err)
	}
	log.Printf("token acquired")

	go func() {
		for {
			<-time.After(300 * time.Second)
			err := updateToken(upstreamDomain)
			if err != nil {
				log.Printf("unable to renew the token: %v\n", err)
			}
		}
	}()

	err = startRegistrySync(listenAddress)
	if err != nil {
		log.Printf("unable to start the proxy server: %v\n", err)
	}
}

func updateToken(acrDomain string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("unable to acquire Azure identity: %w\n", err)
	}

	return AcquireRegistryToken(context.Background(), cred, acrDomain, false)
}
