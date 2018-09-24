package main

import macaron "gopkg.in/macaron.v1"

func getAlive(ctx *macaron.Context) {
	ctx.RawData(200, []byte("ALIVE"))
}
