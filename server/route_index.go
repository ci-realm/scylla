package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	proxyTarget, err := url.Parse("http://localhost:8080")
	if err != nil {
		logger.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyTarget)
	proxy.ServeHTTP(w, r)
}
