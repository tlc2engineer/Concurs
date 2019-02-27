package handlers

import (
	"Concurs/model"
	"Concurs/rgbtree"
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
	var qid int
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
			var err error
			qid, err = strconv.Atoi(val)
			if err != nil {
				errFlag = true
			}
		default: // неизвестный параметр
			errFlag = true
		}
	})
	if errFlag {
		ctx.SetStatusCode(400)
		return
	}
	//------Если есть в кэш----------------
	memData, ok := getCache(qid) //accCashe[qid]
	if ok {                      // есть кэш
		out := make([]*model.User, 0, len(memData))
		for _, id := range memData {
			user := model.GetUser(id)
			out = append(out, user)
		}
		ctx.SetContentType("application/json")
		ctx.Response.Header.Set("charset", "UTF-8")
		ctx.SetStatusCode(200)
		bbuff := bbuf.Get().(*bytes.Buffer)
		ctx.Write(suggestOutput(out, bbuff))
		bbuf.Put(bbuff)
		return
	}
	//---------------------------------------------
	var account model.User
	// находим аккаунт
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
	//-----В кэш--------------
	toCh := make([]uint32, 0, len(sugg))
	for _, user := range sugg {
		toCh = append(toCh, user.ID)
	}
	setCache(qid, toCh) //accCashe[qid] = toCh
	//-------------------------
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
		//buff.WriteString(fmt.Sprintf("\"status\":\"%s\",", model.GetSPVal("status", uint16(account.Status))))
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
	wh := groupMap.Get().(rgbtree.UTree)
	wh.Clear()
	//wh := make(map[uint32]bool) // карта других кто еще лайкал
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
			if o != account.ID {
				_, ok := wh.Get(uint64(o))
				if !ok {
					wh.Put(uint64(o), 1)
				}
			}
		}
	}
	tmp := buffTmps.Get().([]tmpS)
	tmp = tmp[:0]
	sex := account.Sex
	rec := sex
	uintBuff := uintB.Get().([]uint64)
	tkeys := wh.Keys(uintBuff)
	for _, ival := range tkeys {
		i := uint32(ival)
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
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i].s == tmp[j].s {
			return tmp[i].user.ID < tmp[j].user.ID
		}
		return tmp[i].s > tmp[j].s
	})
	filtered = filtered[:0]
	for i := range tmp {
		filtered = append(filtered, tmp[i].user)
	}
	buffTmps.Put(tmp)
	uintB.Put(uintBuff)
	groupMap.Put(wh)
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
