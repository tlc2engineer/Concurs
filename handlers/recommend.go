package handlers

import (
	"Concurs/model"
	"encoding/json"
	"math"
	"sort"
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
			_, ok := model.DataCity[city]
			if !ok {
				noneFlag = true
			}
		case "country":
			country = val
			if country == "" {
				errFlag = true
			}
			_, ok := model.DataCountry[country]
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
	var account model.User
	// находим аккаункт
	account, err := model.GetAccount(uint32(id))
	// Если нет такого аккаунта
	if err != nil {
		ctx.SetStatusCode(404)
		return
	}
	// фильтрация
	filtered := filterRecommend(account, model.DataCountry[country], model.DataCity[city])
	sort.Slice(filtered, func(i, j int) bool {
		f := filtered[i]
		s := filtered[j]
		// по премиум аккаунту
		if f.IsPremium() != s.IsPremium() {
			return f.IsPremium()
		}
		// по статусу
		if f.Status != s.Status {
			return f.Status < s.Status
		}
		// общие интересы
		commf := f.GetCommInt(account)
		comms := s.GetCommInt(account)
		if commf != comms {
			return commf > comms // у кого больше общих интересов
		}
		// по разнице в возрасте
		agef := math.Abs(float64(int64(account.Birth) - int64(f.Birth)))
		ages := math.Abs(float64(int64(account.Birth) - int64(s.Birth)))
		if agef != ages {
			return agef < ages
		}
		// по id
		return f.ID < s.ID
	})

	if limit > 0 && len(filtered) > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	// fmt.Println("Accouunt", account.Interests, account.ID, time.Unix(int64(account.Birth), 0))
	// for _, f := range filtered {
	// 	fmt.Println(model.GetSPVal("city", f.City), f.Interests, f.ID, f.IsPremium(), f.GetCommInt(account),
	// 		int64(account.Birth)-int64(f.Birth), time.Unix(int64(f.Birth), 0), account.Birth, f.Birth)
	// }
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(recommendOutput(filtered))
}

func filterRecommend(account model.User, country uint16, city uint16) []model.User {
	accounts := model.GetAccounts()
	filtered := make([]model.User, 0)
	sex := account.Sex
	rec := !sex // противоположный пол
	for _, acc := range accounts {
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
		if acc.GetCommInt(account) == 0 {
			continue
		}

		filtered = append(filtered, acc)
	}
	return filtered

}

func recommendOutput(accounts []model.User) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		if account.IsPremium() {
			prem := model.Premium{Start: int64(account.Start), Finish: int64(account.Finish)}
			dat["premium"] = prem
		}
		dat["email"] = account.Email
		dat["id"] = account.ID
		dat["status"] = model.GetSPVal("status", uint16(account.Status))
		if account.SName != "" {
			dat["sname"] = account.SName
		}
		if account.FName != "" {
			dat["fname"] = account.FName
		}
		//dat["interests"] = account.Interests
		dat["birth"] = account.Birth
		out = append(out, dat)
	}
	resp["accounts"] = out
	bts, _ := json.Marshal(resp)
	return bts
}
