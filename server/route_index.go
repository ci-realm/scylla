package server

import (
	"io/ioutil"
	"net/http"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	indexHtml, ok := ioutil.ReadFile("static/index.html")
	if ok != nil {
		logger.Fatal("Couldn't read static/index.html")
	}

	w.Write(indexHtml)

	// proxyTarget, err := url.Parse("http://localhost:8080")
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// proxy := httputil.NewSingleHostReverseProxy(proxyTarget)
	// proxy.ServeHTTP(w, r)
}
