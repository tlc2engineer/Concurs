package handlers

import (
	"Concurs/model"
	"encoding/json"
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
	var account model.Account
	// находим аккаункт
	account, err := model.GetAccount(id)
	// Если нет такого аккаунта
	if err != nil {
		ctx.SetStatusCode(404)
		return
	}
	//fmt.Println(account.SName, account.FName, account.Sex)
	// фильтрация по стране  полу городу
	filtered := filterSuggest(account, country, city)
	// сортировка по предпочтениям
	sort.Slice(filtered, func(i, j int) bool {
		f := filtered[i]
		s := filtered[j]
		return f.Suggest(account) > s.Suggest(account)
	})
	// составление карты предпочтений самого пользователя
	idMap := make(map[int]bool)
	lids := account.FilterLike()
	for _, lid := range lids {
		idMap[int(lid)] = false
	}
	sugg := getSuggestAcc(filtered, idMap, limit, city, country)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(suggestOutput(sugg))
}

/*suggestOutput - вывод данных*/
func suggestOutput(accounts []model.Account) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		dat["email"] = account.Email
		dat["id"] = account.ID
		dat["status"] = account.Status
		if account.SName != "" {
			dat["sname"] = account.SName
		}
		if account.FName != "" {
			dat["fname"] = account.FName
		}
		out = append(out, dat)
	}
	resp["accounts"] = out
	bts, _ := json.Marshal(resp)
	return bts
}

/*filterSuggest - фильтрация пользователей по полу,стране,городу*/
func filterSuggest(account model.Account, country string, city string) []model.Account {
	accounts := model.GetAccounts()
	filtered := make([]model.Account, 0)
	likes := account.Likes
	sex := account.Sex
	for _, acc := range accounts {
		found := false
		alikes := acc.Likes
	m:
		for _, like := range likes {
			for _, alike := range alikes {
				if like.ID == alike.ID {
					found = true
					break m
				}
			}
		}
		if !found {
			continue
		}
		if acc.Sex != sex {
			continue
		}
		if country != "" {
			if acc.Country != country {
				continue
			}
		}
		if city != "" {
			if acc.City != city {
				continue
			}
		}
		filtered = append(filtered, acc)
	}
	return filtered

}

/*getSuggestAcc - аккаунты которые любят пользователи с близкими симпатиями*/
func getSuggestAcc(sugg []model.Account, exclID map[int]bool, limit int, city string, country string) []model.Account {
	ret := make([]model.Account, 0) // возвращаемое значение
	for i := range sugg {
		ids := sugg[i].FilterLike()               // id предпочитает данный пользователь
		tmp := make([]model.Account, 0, len(ids)) // временный срез для id пользователя которые не предпочитает целевой
		for _, id := range ids {                  // id которые предпочитал пользователь
			_, ok := exclID[int(id)] // фильтрация id которые предпочитает целевой пользователь
			if ok {
				continue
			}
			acc, _ := model.GetAccount(int(id))
			// if city != "" && acc.City != city {
			// 	continue
			// }
			// if country != "" && acc.Country != country {
			// 	continue
			// }

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
