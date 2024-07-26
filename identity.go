package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	aadTokenExpireTime time.Time
	registryToken      string //a JWT
	tokenMu            sync.RWMutex
)

func AcquireRegistryToken(ctx context.Context, azIdentity azcore.TokenCredential, acrDomain string, forceRenew bool) (err error) {
	// https://stackoverflow.com/a/72665080

	tokenMu.RLock()
	if (!forceRenew) && (aadTokenExpireTime.Sub(time.Now()) > tokenUpdateThreshold) {
		// no need to renew
		tokenMu.RUnlock()
		return nil
	}
	tokenMu.RUnlock()

	// renew token
	aadToken, err := azIdentity.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return fmt.Errorf("unable to get azure identity token: %w", err)
	}

	// get registry login token
	registryTokenResp, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", acrDomain), url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrDomain},
		"access_token": {aadToken.Token},
	})
	if err != nil {
		return fmt.Errorf("unable to exchange registry token: %w", err)
	}

	var response map[string]interface{}
	err = json.NewDecoder(registryTokenResp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("unable to decode registry token: %w", err)
	}

	tokenMu.Lock()
	aadTokenExpireTime = aadToken.ExpiresOn
	registryToken = response["refresh_token"].(string)
	tokenMu.Unlock()
	return nil
}
