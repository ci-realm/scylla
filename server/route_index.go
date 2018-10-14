package server

import macaron "gopkg.in/macaron.v1"

func getIndex(ctx *macaron.Context) {
	ctx.HTML(200, "index")
}
