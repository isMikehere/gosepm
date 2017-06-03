package handler

import (
	"gopkg.in/macaron.v1"
)

func TrxHandler(ctx *macaron.Context) {
	ctx.HTML(200, "trx")
}
