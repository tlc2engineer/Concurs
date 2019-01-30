package handlers

import (
	"Concurs/model"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

//var actParams = map[string]param{}

/*Group - группировка*/
func Group(ctx *fasthttp.RequestCtx) {
	var keys []string
	errFlag := false
	limit := -1
	order := 1
	actParams := map[string]param{}
	ctx.QueryArgs().VisitAll(func(kp, v []byte) {
		k := string(kp)
		val := string(v)
		switch k {
		case "keys":
			if val == "" {
				ctx.SetStatusCode(400)
				return
			}
			keys = strings.Split(val, ",")
			if len(keys) == 0 {
				errFlag = true
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
					errFlag = true
				}
			}
			actParams[k] = param{sval: val, tp: st}
		case "limit":
			num, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				errFlag = true
			}
			limit = int(num)
			if limit <= 0 {
				errFlag = true
			}
			actParams[k] = param{ival: int64(limit), tp: it}
		case "order":
			num, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				errFlag = true
			}
			order = int(num)
			if order != -1 && order != 1 {
				errFlag = true
			}
			actParams[k] = param{ival: int64(order), tp: it}
		case "likes", "birth", "joined", "query_id":
			num, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				errFlag = true
			}
			data := int(num)
			actParams[k] = param{ival: int64(data), tp: it}
		case "sex":
			if val != "m" && val != "f" {
				errFlag = true
			}
			actParams[k] = param{sval: val, tp: st}
		case "country", "city", "interests", "status":
			actParams[k] = param{sval: val, tp: st}
		default:
			errFlag = true

		}
	})
	if errFlag {
		//fmt.Println("err flag")
		ctx.SetStatusCode(400)
		return
	}
	//fmt.Println(actParams)
	// Получение параметров и верификация
	/*
		for k, v := range params {
			//ctx.QueryArgs().
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
	*/
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

	//----------------------------------
	ss := func() (int, int, uint16, uint16, []uint16, error) {
		var city, country uint16
		var dat []uint16
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
		_, ok = actParams["city"]
		if ok {
			city, ok = model.DataCity.Get(actParams["city"].sval)
			if !ok {
				return 0, 0, 0, 0, nil, fmt.Errorf("Err")
			}
		}
		_, ok = actParams["country"]
		if ok {
			country, ok = model.DataCountry.Get(actParams["country"].sval)
			if !ok {
				return 0, 0, 0, 0, nil, fmt.Errorf("Err")
			}
		}
		_, ok = actParams["interests"]
		if ok {
			pari := strings.Split(actParams["interests"].sval, ",")
			dat = make([]uint16, len(pari))
			for i := range pari {
				dat[i], ok = model.DataInter.Get(pari[i])
				if !ok {
					return 0, 0, 0, 0, nil, fmt.Errorf("Err")
				}
			}
		}
		return isex, istatus, country, city, dat, nil
	}
	//--------------------------------------------
	var f1 bool
	var f2 bool
	var f3 = false
	//--------------------------------------------
	_, okb := actParams["birth"]
	_, okl := actParams["likes"]
	_, okj := actParams["joined"]
	if okb && !okl && !okj {
		isex, istatus, country, city, dat, err := ss()
		year := actParams["birth"].ival
		if err == nil {
			f3 = model.GBirthY(keys, isex, istatus, resMap, country, city, dat, int(year))
		} else {
			f3 = true
		}
	}
	//--------------------------------------------
	if !okb && !okl && okj {
		isex, istatus, country, city, dat, err := ss()
		year := actParams["joined"].ival
		if err == nil {
			f3 = model.GJoinY(keys, isex, istatus, resMap, country, city, dat, int(year))
		} else {
			f3 = true
		}
	}
	//------Ключи для второго варианта------------
	secondInd := []string{"birth", "joined", "likes"}
	sf := false
ms:
	for par := range actParams {
		for _, skey := range secondInd {
			if par == skey {
				sf = true
				break ms
			}
		}
	}

	//----------------------------------------------------
	if !sf && !f3 {
		isex, istatus, country, city, dat, err := ss()
		//fmt.Println("first", err, isex, istatus, country, city, dat, keys, err)
		//-----------Первый выриант-----------------------
		if err == nil {
			f1 = model.GroupI(keys, isex, istatus, resMap, country, city, dat)
			//fmt.Println(resMap)
		} else {
			f1 = true
		}

	}
	//-----------Второй-------------------------------

	if !f1 && !f3 {

		msg := toMessG(actParams)

		f2 = model.GroupAgg(msg, resMap, ff, fkey)
	}
	//--------------FullScan---------------------------
	if !f1 && !f2 && !f3 {
		//fmt.Println("--mc--", keys, actParams)
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

	//преобразование карты в срез результатов
	var results []res
	var fromBuff bool
	if len(resMap) < 10000 { // берем из буффера
		fromBuff = true
		results := resBuff.Get().([]res)
		results = results[:0]
	} else { //если много то новый
		results = make([]res, 0, len(resMap))
	}
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
	buff := bbuf.Get().(*bytes.Buffer)
	bts := createGroupOutput(results, keys, buff)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(bts)
	bbuf.Put(buff)
	mapBuff.Put(resMap)
	if fromBuff {
		resBuff.Put(results)
	}

}

/*createGroupOutput -вывод данных*/
func createGroupOutput(res []res, keys []string, buff *bytes.Buffer) []byte {

	bg := "{\"groups\":["
	end := "]}"
	buff.Reset()
	buff.WriteString(bg)
	for i, r := range res {
		if len(r.par) == len(keys) {
			buff.WriteString("{")
			buff.WriteString(fmt.Sprintf("\"count\":%d", r.count))
			// if len(keys) > 0 {
			// 	buff.WriteString(",")
			// }
			for _, key := range keys {
				if !(r.par[key] == 0 && (key == "city" || key == "country")) {
					switch key {
					case "sex", "city", "country", "status", "interests":
						buff.WriteString(",")
						//	"interests":"\u0411\u043e\u0435\u0432\u044b\u0435 \u0438\u0441\u043a\u0443\u0441\u0441\u0442\u0432\u0430"
						buff.WriteString(fmt.Sprintf("\"%s\":\"%s\"", key, model.GetSPVal(key, r.par[key])))
						//dat[key] = model.GetSPVal(key, r.par[key])
					}
				}
				// if j != (len(keys) - 1) {
				// 	buff.WriteString(",")
				// }
			}
			buff.WriteString("}")
			if i != (len(res) - 1) {
				buff.WriteString(",")
			}
		}
	}
	buff.WriteString(end)
	return buff.Bytes()
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
