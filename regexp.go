package proxy

import "regexp"

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

func IsDomain(hostname string) bool {
	return domainRegex.MatchString(hostname)
}
