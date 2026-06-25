package proxy

import "net/http"

// HttpMethods contains the first three bytes of all standard HTTP methods
// (GET, POST, PUT, HEAD, PATCH, DELETE, CONNECT, OPTIONS, TRACE).
// It is used by the dispatcher to quickly identify HTTP traffic.
var HttpMethods = map[string]struct{}{
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
