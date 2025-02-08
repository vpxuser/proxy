package proxy

import (
	"regexp"
	"strings"
)

func WildcardFQDN(fqdn string) string {
	fqdnSplit := strings.Split(fqdn, ".")
	if len(fqdnSplit) < 2 {
		return fqdn
	}
	return "*." + strings.Join(fqdnSplit[len(fqdnSplit)-2:], ".")
}

func IsDomain(host string) bool {
	matched, err := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, host)
	if err != nil {
		return false
	}
	return matched
}
