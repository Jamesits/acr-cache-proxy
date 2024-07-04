package utils

import (
	"encoding/base64"
	"io"
	"net/http"
)

func BasicAuthHeaderValue(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func CopyHttpResponse(src *http.Response, w http.ResponseWriter) {
	// copy headers
	for k, vs := range src.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(src.StatusCode)

	// copy body
	defer src.Body.Close()
	_, _ = io.Copy(w, src.Body)
}
