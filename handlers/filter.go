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

type sparam struct {
	par  string
	pred string
}

var sparams = []string{"email", "fname", "sname", "phone", "sex", "country", "city", "status", "interests", "birth", "premium", "likes"}

var predicts = []string{"any", "contains", "domain", "lt", "gt", "eq", "neq", "code", "null", "now", "starts", "year"}

var legalPred = map[string][]string{"email": []string{"lt", "gt", "domain"}, "fname": {"null", "any", "eq"},
	"sname": {"null", "starts", "eq"}, "phone": {"null", "code"}, "sex": {"eq"}, "country": {"null", "eq"},
	"city": {"null", "eq", "any"}, "status": {"eq", "neq"}, "interests": {"any", "contains"}, "birth": {"lt", "gt", "year"},
	"premium": {"now", "null"}, "likes": {"contains"}}

/*Filter - фильтрация аккаунтов*/
func Filter(ctx *fasthttp.RequestCtx) {
	parMap := make(map[string]sparam)
	var limit int
	limit = -1
	errFlag := false
	ctx.QueryArgs().VisitAll(func(kp, v []byte) {
		k := string(kp)
		prm := string(v)
		if k == "query_id" {
			return
		}
		if k == "limit" {
			val, err := strconv.Atoi(prm)
			if err == nil {
				limit = val
			} else {
				errFlag = true
			}
			return
		}
		find := false
		for _, spar := range sparams {
			if strings.HasPrefix(k, spar) {
				find = true
				sp := sparam{}
				sp.par = prm                  // значение параметра
				if strings.Contains(k, "_") { // если есть предикат
					args := strings.Split(k, "_")
					if len(args) == 2 && args[0] == spar { //два аргумента и первый это параметр
						sp.pred = args[1]
					} else {
						fmt.Println("no arguments")
						errFlag = true
					}
				} else {
					errFlag = true
				}
				parMap[spar] = sp
			}
		}
		if !find { // параметр не найден
			fmt.Println("par not found " + k)
			errFlag = true
		}

	})
	if errFlag {
		ctx.SetStatusCode(400)
		return
	}
	// не найден limit ошибка
	if limit == -1 {
		retZero(ctx)
		return
	}
	err := verifyFilter(parMap)
	if err != nil {
		fmt.Println("no verify " + err.Error())
		ctx.SetStatusCode(400)
		return
	}
	//---------------------------------------------------
	//accounts := model.GetAccounts()
	accounts := model.IndexAgg(toMess(parMap))
	//fmt.Println("--acc1--", len(accounts))
	resp := make([]model.User, 0)                // ответ
	filtFunc := make([]func(model.User) bool, 0) // список функций фильтрации
	var f func(model.User) bool                  // промежуточная переменная
	// установка фильтров
	noneFlag := false
	for k := range parMap {
		switch k {
		case "email", "sname", "fname", "phone":
			f = func(par string) func(acc model.User) bool {
				return func(acc model.User) bool {
					return filterAcc(acc, par, parMap)
				}
			}(k)
		case "city":
			f = func(acc model.User) bool {
				pred := parMap["city"].pred
				par := parMap["city"].par
				switch pred {
				case "null":
					if par == "1" {
						return acc.City == 0
					}
					if par == "0" {
						return acc.City != 0
					}
					/*
						case "eq":
							city, ok := model.DataCity[par]
							if !ok {
								return false //noneFlag = true
							}
							return acc.City == city
						case "any":
							if acc.City == 0 {
								return false
							}
							cities := strings.Split(par, ",")
							for _, city := range cities {
								if model.DataCity[city] == acc.City {
									return true
								}
							}
							return false
					*/
				}
				return true
			}
		case "country":
			f = func(acc model.User) bool {
				pred := parMap["country"].pred
				par := parMap["country"].par
				switch pred {
				case "null":
					// if par == "1" {
					// 	return acc.Country == 0
					// }
					if par == "0" {
						return acc.Country != 0
					}
					/*
						case "eq":
							return acc.Country == model.DataCountry[par]
					*/
				}
				return true
			}
		case "sex":
			f = func(acc model.User) bool {
				par := parMap["sex"].par
				return ((par == "m") && acc.Sex) || ((par == "f") && !acc.Sex)
			}
		case "status":
			f = func(acc model.User) bool {
				pred := parMap["status"].pred
				par := parMap["status"].par
				switch pred {
				case "eq":
					return acc.Status == model.DataStatus[par]
				case "neq":
					return acc.Status != model.DataStatus[par]
				}
				return false
			}
		case "interests":
			f = func(acc model.User) bool {
				return filterInterests(acc, "interests", parMap)
			}
		case "likes":
			//accounts = make([]model.User, 0)
			accmap := make(map[uint32]model.User)
			par := parMap["likes"].par
			if par == "" {
				continue
			}
			args := strings.Split(par, ",")
			for _, p := range args {
				num, _ := strconv.ParseInt(p, 10, 0)
				ids, err := model.GetWho(uint32(num))
				if err != nil {
					continue
				}
				//fmt.Println(ids)
				for i := 0; i < ids.Len(); i++ {
					idd := ids.GetId(i)
					accmap[idd], _ = model.GetAccount(uint32(idd))
					//accounts = append(accounts, acc)
				}
			}
			accounts = make([]model.User, 0)
			for _, acc := range accmap {
				accounts = append(accounts, acc)
			}
			//continue
			f = func(acc model.User) bool {
				return filterLikes(acc, "likes", parMap)
			}
		case "premium":
			f = func(acc model.User) bool {
				return filterPremium(acc, "premium", parMap)
			}
		case "birth":
			f = func(acc model.User) bool {
				return filterDate(acc, "birth", parMap)
			}
		case "joined":
			f = func(acc model.User) bool {
				return filterDate(acc, "joined", parMap)
			}
		}
		filtFunc = append(filtFunc, f)
	}
	if noneFlag {
		retZero(ctx)
		return
	}
	//fmt.Println("--acc--", len(accounts))
	// фильтрация
m1:
	for _, account := range accounts {

		for _, f := range filtFunc {
			if !f(account) {
				continue m1
			}
		}
		resp = append(resp, account)
	}
	fields := make([]string, 0)
	for k := range parMap {
		par := parMap[k]
		if !(par.pred == "null" && par.par == "1") { //_null=1
			fields = append(fields, k)
		}
	}
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].ID > resp[j].ID
	})
	if len(resp) > limit {
		resp = resp[:limit]
	}
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("charset", "UTF-8")
	ctx.SetStatusCode(200)
	ctx.Write((createFilterOutput(resp, fields)))
}

/*verifyFilter - проверка строки запроса*/
func verifyFilter(params map[string]sparam) error {
	for k, v := range params {
		par := v.par
		pred := v.pred
		find := false
		for _, pr := range predicts {
			if pr == pred {
				find = true
				break
			}
		}
		if !find {
			fmt.Println("predict not found " + pred)
			return fmt.Errorf("Predict not found in predict list")
		}
		legPredict := legalPred[k]
		find = false
		for _, pr := range legPredict {
			if pr == pred {
				find = true
			}
		}
		if !find {
			fmt.Println("predict not found in legal list " + pred)
			return fmt.Errorf("Predict not found in legal predict list")
		}
		switch k {
		case "birth":
			switch pred {
			case "lt":
				fallthrough
			case "gt":
				num, err := strconv.ParseInt(par, 10, 0)
				if err != nil {
					return fmt.Errorf("Illegal birth")
				}
				birth := time.Unix(num, 0).In(loc)
				if birth.Year() < 1950 || birth.Year() > 2005 {
					return fmt.Errorf("Illegal birth year")
				}
			case "year":
				year, err := strconv.ParseInt(par, 10, 0)
				if err != nil {
					return fmt.Errorf("Illegal birth year")
				}
				if year < 1950 || year > 2005 {
					return fmt.Errorf("Illegal birth year")
				}
			}
		case "joined":
			num, err := strconv.ParseInt(par, 10, 0)
			if err != nil {
				return fmt.Errorf("Illegal joined")
			}
			joined := time.Unix(num, 0).In(loc)
			if joined.Year() > 2005 || joined.Year() > 2018 {
				return fmt.Errorf("Illegal joined year")
			}
		case "sex":
			if par != "f" && par != "m" {
				return fmt.Errorf("Illegal sex")
			}
		case "likes":
			lks := strings.Split(par, ",")
			for _, p := range lks {
				_, err := strconv.ParseInt(p, 10, 0)
				if err != nil {
					return fmt.Errorf("Illegal like " + p)
				}
			}
		}
	}
	return nil
}

/*filterAcc - фильтр строчного параметра*/
func filterAcc(account model.User, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	// если параметр не контролируется
	if par == "" {
		return true
	}
	pred := parMap[pname].pred
	accP := ""
	switch pname {
	case "email":
		accP = account.Email
	case "fname":
		accP = account.GetFname()
	case "sname":
		accP = account.GetSname()
	case "phone":
		accP = account.Phone
	}
	switch pred {
	case "eq":
		return accP == par
	case "null":
		if par == "1" {
			return accP == ""
		}
		if par == "0" {
			return accP != ""
		}
	case "code":
		return strings.Contains(accP, "("+par+")")
	case "domain":
		return strings.Contains(accP, par)
	case "any":
		cities := strings.Split(par, ",")
		for _, city := range cities {
			if city == accP {
				return true
			}
		}
		return false
	case "none":
		return accP == par
	case "neq":
		return par != accP
	case "gt":
		return strings.Compare(accP, par) > 0
	case "lt":
		return strings.Compare(accP, par) < 0
	case "starts":
		return strings.HasPrefix(accP, par)
	}
	return true
}

/*filterInterests - фильтр интересов*/
func filterInterests(account model.User, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	interests := account.Interests
	pred := parMap[pname].pred
	pari := strings.Split(par, ",")
	dat := make([]uint16, len(pari))
	for i := range pari {
		dat[i], _ = model.DataInter.Get(pari[i])
	}
	switch pred {
	case "contains":
		for _, p := range dat {
			find := false
			for _, inter := range interests {
				if p == inter {
					find = true
					break
				}
			}
			if !find {
				return false
			}
		}
		return true
	case "any":

		for _, inter := range interests {
			for _, p := range dat {
				if p == inter {
					return true
				}
			}
		}
		return false
	case "neq":
		for _, inter := range interests {
			v, _ := model.DataInter.Get(par)
			if v != inter {
				return false
			}
		}
		return true
	}
	return true
}

/*filterLikes - фильтр лайков*/
func filterLikes(account model.User, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	id := account.ID
	likes := model.UnPackLSlice(model.GetLikes(id))
	lnums := make([]int64, 0, len(likes))
	args := strings.Split(par, ",")
	for _, p := range args {
		num, _ := strconv.ParseInt(p, 10, 0)
		lnums = append(lnums, num)
	}
	// нужно чтобы все совпали
	for _, num := range lnums {
		found := false
		for _, like := range likes {
			if like.ID == num {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

/*filterDate - фильтр по дате*/
func filterDate(account model.User, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	pred := parMap[pname].pred
	var date time.Time
	if pname == "birth" {
		date = time.Unix(int64(account.Birth), 0).In(loc)
	}
	if pname == "joined" {
		date = time.Unix(int64(account.Joined), 0).In(loc)
	}

	if len(par) == 1 && par == "" {
		return true
	}
	switch pred {
	case "lt":
		num, _ := strconv.ParseInt(par, 10, 0)
		if pname == "birth" {
			return int64(account.Birth) < num
		}
		return int64(account.Joined) < num
	case "gt":
		num, _ := strconv.ParseInt(par, 10, 0)
		if pname == "birth" {
			return int64(account.Birth) > num
		}
		return int64(account.Joined) > num
	case "year":

		year, _ := strconv.ParseInt(par, 10, 0)
		return year == int64(date.Year())
	}
	return true
}

/*filterPremium - фильтр по премиум*/
func filterPremium(account model.User, pname string, parMap map[string]sparam) bool {
	pred := parMap[pname].pred
	par := parMap[pname].par
	switch pred {
	case "now":
		start := time.Unix(int64(account.Start), 0).In(model.Loc)
		finish := time.Unix(int64(account.Finish), 0).In(model.Loc)
		now := time.Unix(model.Now, 0).In(model.Loc)
		//acc.mutex.Unlock()
		return now.After(start) && now.Before(finish)
	case "null":
		if par == "0" {
			return !(account.Start == 0 && account.Finish == 0)
		}
		if par == "1" {
			return account.Start == 0 && account.Finish == 0
		}
	}
	return true
}

/*createFilterOutput - вывод фильтра*/
func createFilterOutput(accounts []model.User, fields []string) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		dat["email"] = account.Email
		dat["id"] = account.ID
		//dat["sname"] = account.GetSname()
		if fields != nil {
			for _, field := range fields {
				switch field {
				case "sex":
					osex := "f"
					if account.Sex {
						osex = "m"
					}
					dat["sex"] = osex
				case "fname":
					dat["fname"] = account.GetFname()
				case "sname":
					dat["sname"] = account.GetSname()
				case "phone":
					dat["phone"] = account.Phone
				case "city":
					for k, v := range model.DataCity.GetMap() {
						if v == account.City {
							dat["city"] = k
						}
					}
				case "country":
					for k, v := range model.DataCountry.GetMap() {
						if v == account.Country {
							dat["country"] = k
						}
					}
				case "birth":
					dat["birth"] = account.Birth
				case "joined":
					dat["joined"] = account.Joined
				case "status":
					for k, v := range model.DataStatus {
						if v == account.Status {
							dat["status"] = k
						}
					}
				case "premium":
					prem := model.Premium{
						Start:  int64(account.Start),
						Finish: int64(account.Finish),
					}
					dat["premium"] = prem
				}
			}
		}
		out = append(out, dat)
	}
	resp["accounts"] = out
	bts, _ := json.Marshal(resp)
	return bts
}

func toMess(m map[string]sparam) []model.Mess {
	out := make([]model.Mess, len(m))
	for k, v := range m {
		mess := model.Mess{Par: k, Val: v.par, Act: v.pred}
		out = append(out, mess)
	}
	return out
}
