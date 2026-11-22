//go:build dev

package app

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func registerFrontend(mux *http.ServeMux) {
	targetURL, err := url.Parse("http://localhost:3000")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", httputil.NewSingleHostReverseProxy(targetURL))
}
