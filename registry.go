package main

import (
	"fmt"
	"github.com/jamesits/acr-cache-proxy/pkg/registry"
	"github.com/jamesits/acr-cache-proxy/pkg/utils"
	"gopkg.in/elazarl/goproxy.v1"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

var proxyServer *goproxy.ProxyHttpServer

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func fakeAuthMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"http://%s%s\",service=\"%s\"", r.Host, tokenEndpointPath, masqueradeDomain))
	w.WriteHeader(http.StatusUnauthorized)
}

func fwdAuth(w http.ResponseWriter, r *http.Request) {
	tokenReq, err := http.NewRequest(http.MethodGet, upstreamRealm, nil)
	if err != nil {
		log.Printf("unable to create HTTP request: %v\n", err)
		return
	}
	q := tokenReq.URL.Query()
	q.Set("service", upstreamService)
	q.Set("scope", registry.ScopePrepend(r.URL.Query().Get("scope"), upstreamPrefix))
	q.Set("account", username)
	tokenReq.URL.RawQuery = q.Encode()
	tokenMu.RLock()
	tokenReq.Header.Set("Authorization", utils.BasicAuthHeaderValue(username, registryToken))
	tokenMu.RUnlock()

	tokenRep, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		log.Printf("auth failed: %v\n", err)
		return
	}

	utils.CopyHttpResponse(tokenRep, w)
}

func fwdDefault(w http.ResponseWriter, r *http.Request) {
	director := func(target *http.Request) {
		target.URL.Scheme = "https"
		target.URL.Path = registry.PathPrepend(r.URL.Path, upstreamPrefix)
		target.URL.Host = upstreamDomain
		target.Host = upstreamDomain
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("access: %s\n", r.URL.Path)

	if r.URL.Path == tokenEndpointPath {
		fwdAuth(w, r)
		return
	}

	if r.URL.Path == "/v2/" {
		fakeAuthMetadata(w, r)
		return
	}

	// reject unsupported requests
	if !strings.HasPrefix(r.URL.Path, "/v2/") {
		notImplemented(w, r)
		return
	}

	// default to proxy requests
	fwdDefault(w, r)
}

func startRegistrySync(addr string) error {
	http.HandleFunc("/", proxyHandler)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		return fmt.Errorf("unable to start proxy server: %w", err)
	}
	return nil
}
