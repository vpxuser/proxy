package proxy

import (
	"net/http"
)

var httpMethods = map[string]struct{}{
	http.MethodGet[:3]:     {},
	http.MethodHead[:3]:    {},
	http.MethodPost[:3]:    {},
	http.MethodPut[:3]:     {},
	http.MethodPatch[:3]:   {},
	http.MethodConnect[:3]: {},
	http.MethodDelete[:3]:  {},
	http.MethodOptions[:3]: {},
	http.MethodTrace[:3]:   {},
}
