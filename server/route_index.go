package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	indexHtml, ok := ioutil.ReadFile("static/index.html")
	if ok != nil {
		logger.Fatal("Couldn't read static/index.html")
	}

	w.Write(indexHtml)
}

func proxyIndex(w http.ResponseWriter, r *http.Request) {
	proxyTarget, err := url.Parse("http://localhost:8000")
	if err != nil {
		logger.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyTarget)
	proxy.ServeHTTP(w, r)
}
