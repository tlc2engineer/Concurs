package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

const it = 0 // целый тип
const st = 1 // строковой
const dt = 2 // дата

var keysLegal = []string{"city", "country", "sex", "interests", "status"}
var params = map[string]param{"order": param{tp: it}, "limit": param{tp: it}, "likes": param{tp: it}, "birth": param{tp: it}, "sex": param{tp: st}, "joined": param{tp: it}, "status": param{tp: st},
	"interests": param{tp: st}, "city": param{tp: st}, "country": param{tp: st}}
var loc, _ = time.LoadLocation("Europe/London")

//var resMap = make(map[uint64]int, 10000) // карта группировки

type param struct {
	tp   int
	sval string
	ival int64
	dval time.Time
}

type groupRes struct {
	params   map[string]uint16
	count    int
	accounts []model.User
}

var mapBuff = sync.Pool{
	New: func() interface{} {
		return make(map[uint64]int, 10000)
	},
}

/*Group - группировка*/
func Group(ctx *fasthttp.RequestCtx) {
	//now := time.Now()
	//vars := r.URL.Query()
	//Ключи группировки и их верификация
	vkey := string(ctx.QueryArgs().Peek("keys"))
	if vkey == "" {
		//fmt.Println("no keys")
		ctx.SetStatusCode(400)
		return
	}
	keys := strings.Split(vkey, ",")
	if len(keys) == 0 {
		//fmt.Println("no keys")
		ctx.SetStatusCode(400)
		return
	}
	// Проверка ключей поиска
	for _, key := range keys {
		found := false
		for _, legal := range keysLegal {
			if key == legal { // ключа нет в списке
				found = true
				break
			}
		}
		if !found {
			//fmt.Println("illegal key " + key)
			ctx.SetStatusCode(400)
			return
		}
	}
	// Получение параметров и верификация
	actParams := map[string]param{}
	for k, v := range params {
		p := string(ctx.QueryArgs().Peek(k))
		if p != "" {
			switch v.tp {
			case it:
				ival, err := strconv.ParseInt(p, 10, 0)
				if err != nil {
					ctx.SetStatusCode(400)
					return
				}
				if k == "order" && ival != 1 && ival != -1 {
					ctx.SetStatusCode(400)
					return
				}
				v.ival = ival
				actParams[k] = v
			case st:
				if k == "sex" {
					if p != "m" && p != "f" {
						ctx.SetStatusCode(400)
						return
					}
				}
				v.sval = p
				actParams[k] = v
			case dt:
				ival, err := strconv.ParseInt(p, 10, 0)
				if err != nil {
					ctx.SetStatusCode(400)
					return
				}
				tm := time.Unix(ival, 0)
				if tm.Year() < 1950 {
					ctx.SetStatusCode(400)
					return
				}
			}

		}
	}
	// order и limit
	limit := -1
	limP, ok := actParams["limit"]
	if ok {
		limit = int(limP.ival)
	}
	order := 1
	orderP, ok := actParams["order"]
	if ok {
		order = int(orderP.ival)
	}
	// группировка
	// удаление значений

	fkey := createFKey(keys) // преобразование в ключ поиска
	resMap := mapBuff.Get().(map[uint64]int)
	for k := range resMap {
		delete(resMap, k)
	}
	// Фильтрация
	// Список функций фильтрации
	ff := make([]func(model.User) bool, 0)
	if sexP, ok := actParams["sex"]; ok {
		if sexP.sval == "m" {
			f := func(acc model.User) bool {
				return acc.Sex
			}
			ff = append(ff, f)
		}
		if sexP.sval == "f" {
			f := func(acc model.User) bool {
				return !acc.Sex
			}
			ff = append(ff, f)
		}
	}
	if statusP, ok := actParams["status"]; ok {
		status := model.DataStatus[statusP.sval]
		f := func(acc model.User) bool {
			return acc.Status == status
		}
		ff = append(ff, f)
	}
	//-----------использование индексов---------------
	fInd := model.GroupAgg(toMessG(actParams), resMap, ff, fkey)
	//------------------------------------------------
	if !fInd {
		ckeys := map[string]bool{"city": false, "sex": false, "status": false}
		find := true
		for _, key := range keys {
			_, ok := ckeys[key]
			if !ok {
				find = false
				break
			}
		}
		if find {
			var isex, istatus int
			psex, ok := actParams["sex"]
			if !ok {
				isex = -1
			} else {
				ssex := psex.sval
				if ssex == "m" {
					isex = 1
				}
			}
			pstatus, ok := actParams["status"]
			if !ok {
				istatus = -1
			} else {
				istatus = int(model.DataStatus[pstatus.sval])
			}
			model.GroupCity(keys, isex, istatus, resMap)
		}
		//-------------------------------------------------------------------------
		if !find {
			cokeys := map[string]bool{"country": false, "sex": false, "status": false}
			find = true
			for _, key := range keys {
				_, ok := cokeys[key]
				if !ok {
					find = false
					break
				}
			}
			if find {
				var isex, istatus int
				psex, ok := actParams["sex"]
				if !ok {
					isex = -1
				} else {
					ssex := psex.sval
					if ssex == "m" {
						isex = 1
					}
				}
				pstatus, ok := actParams["status"]
				if !ok {
					istatus = -1
				} else {
					istatus = int(model.DataStatus[pstatus.sval])
				}
				model.GroupCountry(keys, isex, istatus, resMap)
			}
		}
		//-------------------------------------------
		if !find {
			ikeys := map[string]bool{"interests": false, "sex": false, "status": false}
			find = true
			for _, key := range keys {
				_, ok := ikeys[key]
				if !ok {
					find = false
					break
				}
			}
			if find {
				var isex, istatus int
				psex, ok := actParams["sex"]
				if !ok {
					isex = -1
				} else {
					ssex := psex.sval
					if ssex == "m" {
						isex = 1
					}
				}
				pstatus, ok := actParams["status"]
				if !ok {
					istatus = -1
				} else {
					istatus = int(model.DataStatus[pstatus.sval])
				}
				model.GroupInt(keys, isex, istatus, resMap)
			}

		}
		//--------------------------------------------
		if !find {
			// основной цикл перебор
			accounts := model.GetAccounts()
		m:
			for _, account := range accounts {
				// все фильтры
				for _, f := range ff {
					if !f(account) {
						continue m
					}
				}
				// группировка
				newSres := fkey(account)
				for _, r := range newSres {
					count, ok := resMap[r]
					if ok {
						count++
						resMap[r] = count
					} else {
						resMap[r] = 1
					}
				}
			}
		}
	}
	// if true {
	// 	ctx.SetStatusCode(400)
	// 	return
	// }
	//преобразование карты в срез результатов
	results := make([]res, 0, len(resMap)) //результаты группировки
	for k, v := range resMap {
		result := res{unpackKey(k, keys), v}
		results = append(results, result)
	}
	// Сортировка
	if order == 1 {
		sort.Slice(results, func(i, j int) bool {
			f := results[i]
			s := results[j]
			if s.count != f.count {
				return f.count < s.count
			}
			for _, key := range keys {
				if f.par[key] != s.par[key] {
					return strings.Compare(model.GetSPVal(key, f.par[key]), model.GetSPVal(key, s.par[key])) < 0
				}
			}
			return false
		})
	}
	if order == -1 {
		sort.Slice(results, func(i, j int) bool {
			f := results[i]
			s := results[j]
			if s.count != f.count {
				return f.count > s.count
			}
			for _, key := range keys {
				if f.par[key] != s.par[key] {
					return strings.Compare(model.GetSPVal(key, f.par[key]), model.GetSPVal(key, s.par[key])) > 0
				}
			}
			return false
		})
	}
	// ограничение по длине
	if len(results) > limit {
		results = results[:limit]
	}
	// Вывод
	bts := createGroupOutput(results, keys)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(bts)
	mapBuff.Put(resMap)
	//fmt.Println(time.Since(now), string(ctx.URI().QueryString()))
}

/*createGroupOutput -вывод данных*/
func createGroupOutput(res []res, keys []string) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(res))
	for _, r := range res {
		if len(r.par) == len(keys) {
			dat := make(map[string]interface{})
			dat["count"] = r.count
			for _, key := range keys {
				if !(r.par[key] == 0 && (key == "city" || key == "country")) {
					switch key {
					case "sex", "city", "country", "status":
						dat[key] = model.GetSPVal(key, r.par[key])
						// if key == "status" {
						// 	fmt.Println(r.par[key], dat[key])
						// }
					}
				}
			}
			out = append(out, dat)
		}
	}
	resp["groups"] = out
	bts, _ := json.Marshal(resp)
	return bts
}

/*groupResults - группировка по строковым параметрам*/
func groupResults(name string, results []groupRes) []groupRes {
	for _, gr := range results {
		fmap := make(map[uint16][]model.User)
		var val uint16
		for _, account := range gr.accounts {
			switch name {
			case "city":
				val = account.City
			case "country":
				val = account.Country
			case "status":
				val = uint16(account.Status)
			case "sex":
				val = 0
				if account.Sex {
					val = 1
				}
			}
			list, ok := fmap[val]
			if ok {
				list = append(list, account)
			} else {
				list = []model.User{account}
			}
			fmap[val] = list
		}
		params := gr.params
		for k, v := range fmap {
			newpar := make(map[string]uint16)
			for k1, v1 := range params {
				newpar[k1] = v1
			}
			newpar[name] = k
			one := groupRes{params: newpar, accounts: v}
			results = append(results, one)
		}

	}
	return results
}

/*groupInterests - группировка по интересам*/
func groupInterests(results []groupRes) []groupRes {
	for _, gr := range results {
		fmap := make(map[uint16][]model.User)
		for _, account := range gr.accounts {
			for _, inter := range account.Interests {
				list, ok := fmap[inter]
				if ok {
					list = append(list, account)
				} else {
					list = []model.User{account}
				}
				fmap[inter] = list
			}
		}
		params := gr.params
		for k, v := range fmap {
			newpar := make(map[string]uint16)
			for k1, v1 := range params {
				newpar[k1] = v1
			}
			newpar["interests"] = k
			one := groupRes{params: newpar, accounts: v}
			results = append(results, one)
		}
	}
	return results
}

/*createFKey - создается функция которая генерирует ключ*/
func createFKey(keys []string) func(user model.User) []uint64 {
	f := false // флаг интересов
	for _, key := range keys {
		if key == "interests" {
			f = true
		}
	}
	return func(user model.User) []uint64 {
		cnt := 1 // число интересов если они есть
		if f {
			cnt = len(user.Interests) // если есть интересы
		}
		out := make([]uint64, cnt)
		for i := 0; i < cnt; i++ {
			var buff = make([]byte, 0)
			for _, key := range keys {
				switch key {
				case "interests":
					inter := user.Interests[i]
					b0 := byte(inter)
					b1 := byte(inter >> 8)
					buff = append(buff, b0, b1)
				case "city":
					city := user.City
					b0 := byte(city)
					b1 := byte(city >> 8)
					buff = append(buff, b0, b1)
				case "country":
					country := user.Country
					b0 := byte(country)
					b1 := byte(country >> 8)
					buff = append(buff, b0, b1)
				case "status":
					buff = append(buff, user.Status)
				case "sex":
					if user.Sex {
						buff = append(buff, 1)
					} else {
						buff = append(buff, 0)
					}
				}

			}

			// запаковка
			for j := 0; j < len(buff); j++ {
				out[i] |= uint64(buff[j]) << (uint16(j) * 8)
			}
		}
		return out
	}
}

type res struct {
	par   map[string]uint16
	count int
}

func unpackKey(vkey uint64, keys []string) map[string]uint16 {
	m := make(map[string]uint16)
	mark := 0
	for _, k := range keys {
		switch k {
		case "interests", "city", "country": // 2 байта
			b0 := byte(vkey >> (uint16(mark) * 8))
			b1 := byte(vkey >> (uint16(mark+1) * 8))
			val := uint64(b0) | (uint64(b1) << 8)
			m[k] = uint16(val)
			mark += 2
		case "status", "sex": // 1 байт
			b0 := byte(vkey >> (uint16(mark) * 8))
			//fmt.Println("-------", b0, vkey, mark)
			m[k] = uint16(b0)
			mark++
		}
	}
	return m
}

func toMessG(in map[string]param) []model.Mess {
	out := make([]model.Mess, 0, len(in))
	for k, v := range in {
		var val string
		switch v.tp {
		case it:
			val = fmt.Sprintf("%d", v.ival)
		case st:
			val = v.sval
		case dt:
			val = fmt.Sprintf("%d", v.dval.Unix())
		}
		m := model.Mess{Par: k, Val: val, Act: ""}
		out = append(out, m)
	}
	return out
}
