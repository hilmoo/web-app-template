//go:build dev

package app

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// registerFrontend proxies everything not matched by another route to the Vite dev server.
func registerFrontend(mux *http.ServeMux) {
	targetURL, err := url.Parse("http://localhost:3000")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", httputil.NewSingleHostReverseProxy(targetURL))
}
