package registry

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAuthMetadata(acrDomain string, ua string) (realm string, service string, err error) {
	req, err := http.NewRequest("HEAD", fmt.Sprintf("https://%s/v2/", acrDomain), nil)
	if err != nil {
		return "", "", fmt.Errorf("unable to build request: %w", err)
	}
	req.Header.Add("User-Agent", ua)
	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("unable to get auth metadata: %w", err)
	}

	// Example: `www-authenticate: Bearer realm="https://auth.docker.io/token",service="registry.docker.io"`
	m := rep.Header.Get("WWW-Authenticate")
	m = strings.SplitN(m, " ", 2)[1]
	kvs := strings.Split(m, ",")
	for _, kv := range kvs {
		s := strings.SplitN(kv, "=", 2)
		v := strings.Trim(s[1], "\"")
		switch s[0] {
		case "realm":
			realm = v
		case "service":
			service = v
		}
	}
	return
}
