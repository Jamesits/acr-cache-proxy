package registry

import "strings"

func PathPrepend(origin string, prefix string) string {
	if prefix == "" {
		return origin
	}

	return "/v2/" + prefix + strings.TrimPrefix(origin, "/v2")
}
