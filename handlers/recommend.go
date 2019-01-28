package handlers

import (
	"Concurs/model"
	"encoding/json"
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
	filtered := model.GetFPointers(uint32(id), kcountry, kcity, limit)
	/*
		temps := make([]*tempR, len(filtered))
		rev := func(v byte) byte {
			switch v {
			case 0:
				return 2
			case 2:
				return 0
			}
			return v
		}
		for i := range filtered {
			f := filtered[i]
			var t uint16
			var b byte
			if f.IsPremium() {
				b |= 128
			}
			b |= rev(f.Status)
			b2 := byte(account.GetCommInt(*f))
			agef := math.Abs(float64(int64(account.Birth) - int64(f.Birth)))
			//delta := math.Float64bits(agef)
			//t |= delta
			t |= (uint16(b2))
			t |= (uint16(b) << 8)
			tr := new(tempR)
			tr.p = f
			tr.dt = agef
			tr.t = t
			temps[i] = tr //tempR{filtered[i], t, agef}
		}
		// фильтрация
		sort.Slice(temps, func(i, j int) bool {
			if temps[i].t != temps[j].t {
				return temps[i].t > temps[j].t
				//return temps[i].p.ID < temps[j].p.ID
			}
			if temps[i].dt != temps[j].dt {
				return temps[i].dt < temps[j].dt
			}
			return temps[i].p.ID < temps[j].p.ID
			// f := filtered[i]
			// s := filtered[j]
			// // по премиум аккаунту
			// if f.IsPremium() != s.IsPremium() {
			// 	return f.IsPremium()
			// }
			// // по статусу
			// if f.Status != s.Status {
			// 	return f.Status < s.Status
			// }
			// // общие интересы
			// commf := f.GetCommInt(account)
			// comms := s.GetCommInt(account)
			// if commf != comms {
			// 	return commf > comms // у кого больше общих интересов
			// }
			// // по разнице в возрасте
			// agef := math.Abs(float64(int64(account.Birth) - int64(f.Birth)))
			// ages := math.Abs(float64(int64(account.Birth) - int64(s.Birth)))
			// if agef != ages {
			// 	return agef < ages
			// }
			// // по id
			// return f.ID < s.ID
		})
		// for _, t := range temps {
		// 	fmt.Println(t.t, t.dt, (*t.p).IsPremium(), account.GetCommInt(*t.p), math.Abs(float64(int64(account.Birth)-int64((*t.p).Birth))))
		// }
		if len(temps) > limit {
			temps = temps[:limit]
		}
		filtered = filtered[:0]
		for i := 0; i < len(temps); i++ {
			p := temps[i].p
			filtered = append(filtered, p)
		}*/
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(recommendOutput(filtered))
}

func recommendOutput(accounts []*model.User) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		if account.Start > 0 {
			prem := model.Premium{Start: int64(account.Start), Finish: int64(account.Finish)}
			dat["premium"] = prem
		}
		dat["email"] = account.Email
		dat["id"] = account.ID
		dat["status"] = model.GetSPVal("status", uint16(account.Status))
		if account.SName != 0 {
			dat["sname"] = account.GetSname()
		}
		if account.FName != 0 {
			dat["fname"] = account.GetFname()
		}
		//dat["interests"] = account.Interests
		dat["birth"] = account.Birth
		out = append(out, dat)
	}
	resp["accounts"] = out
	bts, _ := json.Marshal(resp)
	return bts
}

type tempR struct {
	p  *model.User
	t  uint16
	dt float64
}
