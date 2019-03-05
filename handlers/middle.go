package handlers

import (
	"github.com/valyala/fasthttp"
)

func retZero(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(zeroOut)
}
