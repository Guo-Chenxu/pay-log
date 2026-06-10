package path

import "strings"

func Contains(paths map[string]struct{}, path string) bool {
	res, path := false, strings.TrimSuffix(path, "/")
	for p := range paths {
		if strings.HasSuffix(p, "*any") {
			prefix := strings.TrimSuffix(p, "/*any")
			if strings.HasPrefix(path, prefix) {
				res = true
				break
			}
		} else if path == p {
			res = true
			break
		}
	}
	return res
}
