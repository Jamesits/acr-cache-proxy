package registry

import "strings"

func ScopePrepend(origin string, prefix string) string {
	// origin: `repository:library/hello-world:pull`

	if prefix == "" {
		return origin
	}

	sc := strings.Split(origin, ":")
	sc[1] = prefix + "/" + sc[1]
	return strings.Join(sc, ":")
}
