package handlers

import (
	"Concurs/model"
	"encoding/json"
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

/*Group - группировка*/
func Group(ctx *fasthttp.RequestCtx) {
	//vars := r.URL.Query()
	//Ключи группировки и их верификация
	vkey := string(ctx.QueryArgs().Peek("keys"))
	if vkey == "" {
		fmt.Println("no keys")
		ctx.SetStatusCode(400)
		return
	}
	keys := strings.Split(vkey, ",")
	if len(keys) == 0 {
		fmt.Println("no keys")
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
			fmt.Println("illegal key " + key)
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
	// Фильтрация
	filtered := make([]model.User, 0)
	accounts := model.GetAccounts()
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
	for _, account := range accounts {
		cityP, ok := actParams["city"]
		if ok {
			if account.City != model.DataCity[cityP.sval] {
				continue
			}
		}
		countryP, ok := actParams["country"]
		if ok {
			if account.Country != model.DataCountry[countryP.sval] {
				continue
			}
		}
		if likesP, ok := actParams["likes"]; ok {
			likeID := likesP.ival
			found := false
			id := account.ID
			for _, like := range model.UnPackLSlice(model.GetLikes(id)) {
				if like.ID == likeID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if interestP, ok := actParams["interests"]; ok {
			interestV := interestP.sval
			found := false
			for _, interest := range account.Interests {
				if interest == model.DataInter[interestV] {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if sexP, ok := actParams["sex"]; ok {
			if account.Sex == false && sexP.sval == "m" || account.Sex == true && sexP.sval == "f" {
				continue
			}

		}
		if statusP, ok := actParams["status"]; ok {
			if account.Status != model.DataStatus[statusP.sval] {
				continue
			}
		}
		if joinedP, ok := actParams["joined"]; ok {
			year := int(joinedP.ival)
			joinDate := time.Unix(int64(account.Joined), 0).In(loc)
			if joinDate.Year() != year {
				continue
			}
		}
		if birthP, ok := actParams["birth"]; ok {
			year := int(birthP.ival)
			birthDate := time.Unix(int64(account.Birth), 0).In(loc)
			if birthDate.Year() != year {
				continue
			}
		}
		filtered = append(filtered, account)
	}
	// Группировка
	gres := []groupRes{groupRes{params: map[string]uint16{}, accounts: filtered}} // результат
	for _, key := range keys {
		if key != "interests" {
			gres = groupResults(key, gres)
		} else {
			gres = groupInterests(gres)
		}
	}
	filGres := make([]groupRes, 0, len(gres))
	for i := range gres {
		gres[i].count = len(gres[i].accounts)
		gres[i].accounts = nil
		if len(gres[i].params) == len(keys) {
			filGres = append(filGres, gres[i])
		}
	}

	if order == 1 {
		sort.Slice(filGres, func(i, j int) bool {
			if filGres[i].count != filGres[j].count {
				return filGres[i].count < filGres[j].count
			}
			for _, key := range keys {
				if filGres[i].params[key] != filGres[j].params[key] {
					return strings.Compare(model.GetSPVal(key, filGres[i].params[key]), model.GetSPVal(key, filGres[j].params[key])) < 0
				}
			}
			return false
		})
	}
	if order == -1 {
		sort.Slice(filGres, func(i, j int) bool {
			if filGres[i].count != filGres[j].count {
				return filGres[i].count > filGres[j].count
			}
			for _, key := range keys {
				if filGres[i].params[key] != filGres[j].params[key] {
					if filGres[i].params[key] != filGres[j].params[key] {
						return strings.Compare(model.GetSPVal(key, filGres[j].params[key]), model.GetSPVal(key, filGres[i].params[key])) < 0
					}
				}
			}
			return false

		})
	}

	if len(filGres) > limit {
		filGres = filGres[:limit]
	}
	// Вывод
	bts := createGroupOutput(filGres, keys)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write(bts)
}

/*createGroupOutput -вывод данных*/
func createGroupOutput(res []groupRes, keys []string) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(res))
	for _, r := range res {
		if len(r.params) == len(keys) {
			dat := make(map[string]interface{})
			dat["count"] = r.count
			for _, key := range keys {
				if !(r.params[key] == 0 && (key == "city" || key == "country")) {
					switch key {
					case "sex", "city", "country", "status":
						dat[key] = model.GetSPVal(key, r.params[key])
						//case "interests":

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
