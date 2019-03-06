package handlers

import (
	"Concurs/model"
	"encoding/json"
	"strconv"

	"github.com/valyala/fasthttp"
)

/*Info - сервисная информация*/
func Info(ctx *fasthttp.RequestCtx) {
	sid := string(ctx.QueryArgs().Peek("id"))
	id, err := strconv.Atoi(sid)
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}
	user := model.GetUser(uint32(id))
	data, err := json.Marshal(user)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(data)
}
