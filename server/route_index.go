package server

import (
	"net/http/httputil"
	"net/url"

	macaron "gopkg.in/macaron.v1"
)

func getIndex(ctx *macaron.Context) {
	proxyTarget, err := url.Parse("http://localhost:8080")
	if err != nil {
		logger.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyTarget)
	proxy.ServeHTTP(ctx.Resp, ctx.Req.Request)
}
