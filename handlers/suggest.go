package handlers

import (
	"Concurs/model"
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/valyala/fasthttp"
)

/*Suggest - предпочитаемые id*/
func Suggest(ctx *fasthttp.RequestCtx, id int) {
	city := ""
	country := ""
	var limit = -1
	// получение параметров и верификация
	errFlag := false
	ctx.QueryArgs().VisitAll(func(kp, v []byte) {
		k := string(kp)
		val := string(v)
		switch k {
		case "city":
			city = val
			if city == "" {
				errFlag = true
			}
		case "country":
			country = val
			if country == "" {
				errFlag = true
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

	var account model.User
	// находим аккаункт
	account, err := model.GetAccount(uint32(id))
	// Если нет такого аккаунта
	if err != nil {
		ctx.SetStatusCode(404)
		return
	}
	// фильтрация по стране  полу городу
	cityVal, ok := model.DataCity.Get(city)
	if !ok {
		retZero(ctx)
		return
	}
	countryVal, ok := model.DataCountry.Get(country)
	if !ok {
		retZero(ctx)
		return
	}
	filtered := ubuff.Get().([]*model.User)
	filtered = filterSuggest(account, countryVal, cityVal, filtered)

	// сортировка по предпочтениям
	idMap := make(map[uint32]bool)
	lids := model.UnPackLSlice(model.GetLikes(account.ID))
	for _, lid := range lids {
		idMap[uint32(lid.ID)] = false
	}
	kcity, _ := model.DataCity.Get(city)
	kcountry, _ := model.DataCountry.Get(country)

	sugg := getSuggestAcc(filtered, idMap, limit, kcity, kcountry)

	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	bbuff := bbuf.Get().(*bytes.Buffer)
	ctx.Write(suggestOutput(sugg, bbuff))
	ubuff.Put(filtered)
	bbuf.Put(bbuff)
}

/*suggestOutput - вывод данных*/
func suggestOutput(accounts []*model.User, buff *bytes.Buffer) []byte {
	bg := "{\"accounts\":["
	end := "]}"
	buff.Reset()
	buff.WriteString(bg)
	for i, account := range accounts {
		buff.WriteString("{")
		buff.WriteString(fmt.Sprintf("\"email\":\"%s\",\"id\":%d,", account.Email, account.ID))
		buff.WriteString(fmt.Sprintf("\"status\":\"%s\",", model.GetSPVal("status", uint16(account.Status))))
		if account.SName != 0 {
			buff.WriteString(fmt.Sprintf("\"sname\":\"%s\",", account.GetSname()))
		}
		if account.FName != 0 {
			buff.WriteString(fmt.Sprintf("\"fname\":\"%s\",", account.GetFname()))
		}
		buff.WriteString(fmt.Sprintf("\"status\":\"%s\"", model.GetSPVal("status", uint16(account.Status))))
		buff.WriteString("}")
		if i != (len(accounts) - 1) {
			buff.WriteString(",")
		}
	}
	buff.WriteString(end)
	return buff.Bytes()
}

/*filterSuggest - фильтрация пользователей по полу,стране,городу*/
func filterSuggest(account model.User, country uint16, city uint16, filtered []*model.User) []*model.User {
	whos := model.GetLikes(account.ID) // лайки данного аккаунта
	wh := make(map[uint32]bool)        // карта других кто еще лайкал
	//t2 := time.Now()
	for i := 0; i < len(whos)/8; i++ {
		var id uint32 // кого лайкал
		id = uint32(whos[i*8]) | uint32(whos[i*8+1])<<8 | uint32(whos[i*8+2])<<16
		oth, err := model.GetWho(id) // другие кто еще лайкал тот же аккаунт
		if err != nil {
			continue
		}
		// добавляем других в карту
		for i := 0; i < oth.Len(); i++ {
			o := oth.GetId(i)
			_, ok := wh[o]
			if !ok {
				wh[o] = true
			}
		}
	}
	tmp := buffTmps.Get().([]tmpS)
	tmp = tmp[:0]
	sex := account.Sex
	rec := sex
	for i := range wh {
		if i == account.ID {
			continue
		}
		acc := model.MainMap[i]
		if acc.Sex != sex {
			continue
		}
		if acc.Sex != rec {
			continue
		}
		if country != 0 {
			if acc.Country != country {
				continue
			}
		}
		if city != 0 {
			if acc.City != city {
				continue
			}
		}
		s := account.Suggest(acc)

		if s == 0.0 {
			continue
		}
		tmp = append(tmp, tmpS{s: s, user: acc})
	}
	//fmt.Println("sudd", time.Since(t2))
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i].s == tmp[j].s {
			return tmp[i].user.ID < tmp[j].user.ID
		}
		return tmp[i].s > tmp[j].s
	})
	//fmt.Println(tmp[0].s, tmp[0].user.ID, tmp[1].s, tmp[1].user.ID)
	filtered = filtered[:0]
	for i := range tmp {
		filtered = append(filtered, tmp[i].user)
	}
	buffTmps.Put(tmp)
	return filtered
}

/*getSuggestAcc - аккаунты которые любят пользователи с близкими симпатиями*/
func getSuggestAcc(sugg []*model.User, exclID map[uint32]bool, limit int, city uint16, country uint16) []*model.User {
	filtMap := make(map[uint32]bool)
	ret := make([]*model.User, 0) // возвращаемое значение
	for i := range sugg {
		//fmt.Println(sugg[i].ID)
		data := model.GetLikes(sugg[i].ID)         // id предпочитает данный пользователь
		tmp := make([]*model.User, 0, len(data)/8) // временный срез для id пользователя которые не предпочитает целевой
		for i := 0; i < len(data)/8; i++ {         // id которые предпочитал пользователь
			id := uint32(data[i*8]) | uint32(data[i*8+1])<<8 | uint32(data[i*8+2])<<16
			_, ok := exclID[uint32(id)] // фильтрация id которые предпочитает целевой пользователь
			if ok {
				continue
			}
			_, ok = filtMap[id]
			if ok { // уже было
				continue
			}
			filtMap[id] = false
			acc, _ := model.MainMap[uint32(id)]
			tmp = append(tmp, acc)

		}
		// сортировка по ID для одного пользователя
		sort.Slice(tmp, func(i, j int) bool {
			return tmp[i].ID > tmp[j].ID
		})
		ret = append(ret, tmp...)
		if len(ret) > limit && limit != -1 {
			ret = ret[:limit]
			break
		}
	}
	return ret
}

type tmpS struct {
	s    float64
	user *model.User
}
