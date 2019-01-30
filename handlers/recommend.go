package handlers

import (
	"Concurs/model"
	"bytes"
	"fmt"
	"strconv"

	"github.com/valyala/fasthttp"
)

var enParam = []string{"country", "city"}

/*Recommend - рекомендуемые id*/
func Recommend(ctx *fasthttp.RequestCtx, id int) {
	city := ""
	country := ""
	var limit = -1
	// получение параметров и верификация
	errFlag := false
	noneFlag := false
	ctx.QueryArgs().VisitAll(func(kp, v []byte) {
		k := string(kp)
		val := string(v)
		switch k {
		case "city":
			city = val
			if city == "" {
				errFlag = true
			}
			_, ok := model.DataCity.Get(city)
			if !ok {
				noneFlag = true
			}
		case "country":
			country = val
			if country == "" {
				errFlag = true
			}
			_, ok := model.DataCountry.Get(country)
			if !ok {
				noneFlag = true
			}
		case "limit":
			num, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				errFlag = true
			}
			limit = int(num)
			if limit <= 0 {
				errFlag = true
			}
		case "query_id":
		default: // неизвестный параметр
			errFlag = true
		}
	})
	if errFlag {
		ctx.SetStatusCode(400)
		return
	}
	if noneFlag {
		retZero(ctx)
		return
	}
	//var account model.User
	// находим аккаункт
	_, err := model.GetAccount(uint32(id))
	// Если нет такого аккаунта
	if err != nil {
		ctx.SetStatusCode(404)
		return
	}
	kcountry, _ := model.DataCountry.Get(country)
	kcity, _ := model.DataCity.Get(city)
	buff := ubuff.Get().([]*model.User)
	filtered := model.GetFPointers(uint32(id), kcountry, kcity, limit, buff)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	bbuff := bbuf.Get().(*bytes.Buffer)
	ctx.Write(recommendOutput(filtered, bbuff))
	ubuff.Put(buff)
	bbuf.Put(bbuff)
}

func recommendOutput(accounts []*model.User, buff *bytes.Buffer) []byte {
	bg := "{\"accounts\":["
	end := "]}"
	buff.Reset()
	buff.WriteString(bg)

	for i, account := range accounts {
		buff.WriteString("{")
		buff.WriteString(fmt.Sprintf("\"email\":\"%s\",\"id\":%d,", account.Email, account.ID))
		//dat := make(map[string]interface{})
		if account.Start > 0 {
			buff.WriteString(fmt.Sprintf("\"premium\":{\"start\":%d,\"finish\":%d},", account.Start, account.Finish))
		}
		buff.WriteString(fmt.Sprintf("\"status\":\"%s\",", model.GetSPVal("status", uint16(account.Status))))

		if account.SName != 0 {
			buff.WriteString(fmt.Sprintf("\"sname\":\"%s\",", account.GetSname()))
		}
		if account.FName != 0 {
			buff.WriteString(fmt.Sprintf("\"fname\":\"%s\",", account.GetFname()))
		}
		//dat["interests"] = account.Interests
		buff.WriteString(fmt.Sprintf("\"birth\":%d", account.Birth))
		buff.WriteString("}")
		if i != (len(accounts) - 1) {
			buff.WriteString(",")
		}
	}
	buff.WriteString(end)
	return buff.Bytes()
}

type tempR struct {
	p  *model.User
	t  uint16
	dt float64
}
