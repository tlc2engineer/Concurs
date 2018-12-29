package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
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
func Filter(w http.ResponseWriter, r *http.Request) {
	accounts := model.GetAccounts()
	parMap := make(map[string]sparam)
	var limit int
	limit = -1
	vars := r.URL.Query()
	for k := range vars {
		prm := vars.Get(k)
		if k == "query_id" {
			continue
		}
		if k == "limit" {
			val, err := strconv.Atoi(prm)
			if err == nil {
				limit = val
			} else {
				w.WriteHeader(400)
				return
			}
			continue
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
						w.WriteHeader(400)
						return
					}
				} else {
					w.WriteHeader(400)
					return
				}
				parMap[spar] = sp
			}
		}
		if !find { // параметр не найден
			fmt.Println("par not found " + k)
			w.WriteHeader(400)
			return
		}
	}
	// не найден limit ошибка
	if limit == -1 {
		fmt.Println("no limit", r.URL.Path)
		resp := make(map[string][]map[string]interface{})
		out := make([]map[string]interface{}, 0, len(accounts))
		resp["accounts"] = out
		bts, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(200)
		w.Write(bts)
		return
	}
	err := verifyFilter(parMap)
	if err != nil {
		fmt.Println("no verify " + err.Error())
		w.WriteHeader(400)
		return
	}
	resp := make([]model.Account, 0)
	filtFunc := make([]func(model.Account) bool, 0)
	var f func(model.Account) bool
	for k := range parMap {
		switch k {
		case "country", "email", "fname", "sname", "phone", "status", "city", "sex":
			f = func(par string) func(acc model.Account) bool {
				return func(acc model.Account) bool {
					return filterAcc(acc, par, parMap)
				}
			}(k)
		case "interests":
			f = func(acc model.Account) bool {
				return filterInterests(acc, "interests", parMap)
			}
		case "likes":
			f = func(acc model.Account) bool {
				return filterLikes(acc, "likes", parMap)
			}
		case "premium":
			f = func(acc model.Account) bool {
				return filterPremium(acc, "premium", parMap)
			}
		case "birth":
			f = func(acc model.Account) bool {
				return filterDate(acc, "birth", parMap)
			}
		case "joined":
			f = func(acc model.Account) bool {
				return filterDate(acc, "joined", parMap)
			}
		}
		filtFunc = append(filtFunc, f)
	}
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
	//fmt.Println(parMap)
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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(createFilterOutput(resp, fields))
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
func filterAcc(account model.Account, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	// если параметр не контролируется
	if par == "" {
		return true
	}
	pred := parMap[pname].pred
	accP := account.GetSParam(pname)
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
func filterInterests(account model.Account, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	interests := account.Interests
	pred := parMap[pname].pred
	pari := strings.Split(par, ",")
	switch pred {
	case "contains":
		for _, p := range pari {
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
			for _, p := range pari {
				if p == inter {
					return true
				}
			}
		}
		return false
	case "neq":
		for _, inter := range interests {
			if par != inter {
				return false
			}
		}
		return true
	}
	return true
}

/*filterLikes - фильтр лайков*/
func filterLikes(account model.Account, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	likes := account.Likes
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
func filterDate(account model.Account, pname string, parMap map[string]sparam) bool {
	par := parMap[pname].par
	if par == "" {
		return true
	}
	pred := parMap[pname].pred
	var date time.Time
	if pname == "birth" {
		date = time.Unix(account.Birth, 0).In(loc)
	}
	if pname == "joined" {
		date = time.Unix(account.Joined, 0).In(loc)
	}

	if len(par) == 1 && par == "" {
		return true
	}
	switch pred {
	case "lt":
		num, _ := strconv.ParseInt(par, 10, 0)
		if pname == "birth" {
			return account.Birth < num
		}
		return account.Joined < num
	case "gt":
		num, _ := strconv.ParseInt(par, 10, 0)
		if pname == "birth" {
			return account.Birth > num
		}
		return account.Joined > num
	case "year":

		year, _ := strconv.ParseInt(par, 10, 0)
		return year == int64(date.Year())
	}
	return true
}

/*filterPremium - фильтр по премиум*/
func filterPremium(account model.Account, pname string, parMap map[string]sparam) bool {
	premium := account.Premium
	pred := parMap[pname].pred
	par := parMap[pname].par
	switch pred {
	case "now":
		return account.IsPremium() //premium.Start < model.Now && premium.Finish > model.Now
	case "null":
		if par == "0" {
			return premium != model.Premium{}
		}
		if par == "1" {
			return premium == model.Premium{}
		}
	}
	return true
}

/*createFilterOutput - вывод фильтра*/
func createFilterOutput(accounts []model.Account, fields []string) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		dat["email"] = account.Email
		dat["id"] = account.ID
		if fields != nil {
			for _, field := range fields {
				switch field {
				case "sex":
					dat["sex"] = account.Sex
				case "fname":
					dat["fname"] = account.FName
				case "sname":
					dat["sname"] = account.SName
				case "phone":
					dat["phone"] = account.Phone
				case "city":
					dat["city"] = account.City
				case "country":
					dat["country"] = account.Country
				case "birth":
					dat["birth"] = account.Birth
				case "joined":
					dat["joined"] = account.Joined
				case "status":
					dat["status"] = account.Status
				case "premium":
					dat["premium"] = account.Premium
				}
			}
		}
		out = append(out, dat)
	}
	resp["accounts"] = out
	bts, _ := json.Marshal(resp)
	return bts
}
