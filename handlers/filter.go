package handlers

import (
	"Concurs/model"
	"bytes"
	"fmt"
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

var uBuff = make([]model.User, 0, 1000)
var bTs = make([]byte, 10000)
var buff = bytes.NewBuffer(bTs)
var out = make([]map[string]interface{}, 0, 1000)

/*Filter - фильтрация аккаунтов*/
func Filter(ctx *fasthttp.RequestCtx) {
	parMap := make(map[string]sparam)
	var limit int
	limit = -1
	errFlag := false
	//fasthttp.AcquireRequest
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

	filtFunc := make([]func(model.User) bool, 0) // список функций фильтрации
	var f func(model.User) bool                  // промежуточная переменная
	// установка фильтров
	noneFlag := false
	for k := range parMap {
		switch k {
		case "sname":
			par := parMap["sname"].par
			if par == "" {
				continue
			}
			pred := parMap["sname"].pred
			switch pred {
			case "null":
				if par == "0" {
					f = func(acc model.User) bool {
						return acc.SName != 0
					}
				}
				if par == "1" {
					f = func(acc model.User) bool {
						return acc.SName == 0
					}
				}
			default:
				continue
			}
			// f = func(par string) func(acc model.User) bool {
			// 	return func(acc model.User) bool {
			// 		return filterAcc(acc, par, parMap)
			// 	}
			// }(k)

		case "fname":
			par := parMap["fname"].par
			if par == "" {
				continue
			}
			pred := parMap["fname"].pred
			switch pred {
			case "null":
				if par == "0" {
					f = func(acc model.User) bool {
						return acc.FName != 0
					}
				}
				if par == "1" {
					f = func(acc model.User) bool {
						return acc.FName == 0
					}
				}
			default:
				continue
			}
		case "phone":
			par := parMap["phone"].par
			if par == "" {
				continue
			}
			pred := parMap["phone"].pred
			switch pred {
			case "null":
				if par == "1" {
					f = func(acc model.User) bool {
						return acc.Phone == ""
					}
				}
				if par == "0" {
					f = func(acc model.User) bool {
						return acc.Phone != ""
					}
				}
			default:
				continue
			}
		case "email":
			mail := parMap["email"].par
			// если параметр не контролируется
			if mail == "" {
				continue
			}
			pred := parMap["email"].pred
			switch pred {
			case "lt":
				f = func(acc model.User) bool {
					return strings.Compare(acc.Email, mail) < 0
				}
			case "gt":
				f = func(acc model.User) bool {
					return strings.Compare(acc.Email, mail) > 0
				}
			default:
				continue
			}
		case "city":
			pred := parMap["city"].pred
			par := parMap["city"].par
			switch pred {
			case "null":
				if par == "1" {
					f = func(acc model.User) bool {
						return acc.City == 0
					}
				}
				if par == "0" {
					f = func(acc model.User) bool {
						return acc.City != 0
					}
				}
			default:
				continue
			}

		case "country":
			pred := parMap["country"].pred
			par := parMap["country"].par
			if pred == "null" && par == "0" {
				f = func(acc model.User) bool {
					return acc.Country != 0
				}
			} else {
				continue
			}
		case "sex":
			par := parMap["sex"].par
			if par == "m" {
				f = func(acc model.User) bool {
					return acc.Sex
				}
			} else {
				f = func(acc model.User) bool {
					return !acc.Sex
				}
			}
		case "status":
			pred := parMap["status"].pred
			par := parMap["status"].par
			stat := model.DataStatus[par]
			if pred == "eq" {
				f = func(acc model.User) bool {
					return acc.Status == stat
				}
			}
			if pred == "neq" {
				f = func(acc model.User) bool {
					return acc.Status != stat
				}
			}
		case "interests":
			par := parMap["interests"].par
			if par == "" {
				continue
			}
			pred := parMap["interests"].pred
			pari := strings.Split(par, ",")
			dat := make([]uint16, len(pari))
			for i := range pari {
				dat[i], _ = model.DataInter.Get(pari[i])
			}
			switch pred {
			case "contains":
				continue
				/*
					f = func(acc model.User) bool {
						interests := acc.Interests
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
					}*/
			case "any":
				continue /*
					f = func(acc model.User) bool {
						interests := acc.Interests
						for _, inter := range interests {
							for _, p := range dat {
								if p == inter {
									return true
								}
							}
						}
						return false
					}*/
			case "neq":
				v, _ := model.DataInter.Get(par)
				f = func(acc model.User) bool {
					interests := acc.Interests
					for _, inter := range interests {
						if v != inter {
							return false
						}
					}
					return true
				}
			}
		case "likes":
			continue
		case "premium":
			pred := parMap["premium"].pred
			par := parMap["premium"].par
			switch pred {
			case "now":
				f = func(acc model.User) bool {
					return acc.Start < uint32(model.Now) && acc.Finish > uint32(model.Now) //now.After(start) && now.Before(finish)
				}
			case "null":
				if par == "0" {
					f = func(acc model.User) bool {
						return !(acc.Start == 0 && acc.Finish == 0)
					}
				}
				if par == "1" {
					f = func(acc model.User) bool {
						return (acc.Start == 0 && acc.Finish == 0)
					}
				}
			default:
				continue
			}
		case "birth":
			par := parMap["birth"].par
			if par == "" {
				continue
			}
			pred := parMap["birth"].pred
			switch pred {
			case "lt":
				num, _ := strconv.ParseInt(par, 10, 0)
				f = func(acc model.User) bool {
					return int64(acc.Birth) < num
				}
			case "gt":
				num, _ := strconv.ParseInt(par, 10, 0)
				f = func(acc model.User) bool {
					return int64(acc.Birth) > num
				}
			default:
				continue
			}
		case "joined":
			par := parMap["joined"].par
			if par == "" {
				continue
			}
			pred := parMap["joined"].pred
			switch pred {
			case "lt":
				num, _ := strconv.ParseInt(par, 10, 0)
				f = func(acc model.User) bool {
					return int64(acc.Joined) < num
				}
			case "gt":
				num, _ := strconv.ParseInt(par, 10, 0)
				f = func(acc model.User) bool {
					return int64(acc.Joined) > num
				}
			default:
				continue
			}
		}
		filtFunc = append(filtFunc, f)
	}
	if noneFlag {
		retZero(ctx)
		return
	}
	var resp []model.User
	accounts := model.IndexAgg(toMess(parMap), filtFunc, limit)
	if accounts == nil { // общий цикл
		uBuff = uBuff[:0]
		resp = uBuff
		accounts = model.GetAccounts()
		ln := len(accounts)
	m1:
		for i := ln - 1; i >= 0; i-- {

			for _, f := range filtFunc {
				if !f(accounts[i]) {
					continue m1
				}
			}
			resp = append(resp, accounts[i])
			if len(resp) >= limit { // все
				break
			}
		}
	} else {
		resp = accounts
	}
	fields := make([]string, 0)
	for k := range parMap {
		par := parMap[k]
		if !(par.pred == "null" && par.par == "1") { //_null=1
			if k != "id" && k != "email" && k != "interests" && k != "likes" {
				fields = append(fields, k)
			}
		}
	}
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

/*filterDate - фильтр по дате
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
}*/

/*filterPremium - фильтр по премиум
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
}*/

/*createFilterOutput - вывод фильтра*/
func createFilterOutput(accounts []model.User, fields []string) []byte {
	bg := "{\"accounts\":["
	end := "]}"
	// resp := make(map[string][]map[string]interface{})
	// out = out[:0]
	// tmp := make([]byte, 1000)
	// tmpBuff := bytes.NewBuffer(tmp)
	//dat := make(map[string]interface{})
	buff.Reset()
	buff.WriteString(bg)
	//enc := json.NewEncoder(tmpBuff)
	//dat := make(map[string]interface{})
	for i, account := range accounts {
		buff.WriteString("{")
		//dat["email"] = account.Email
		buff.WriteString(fmt.Sprintf("\"email\":\"%s\",\"id\":%d", account.Email, account.ID))
		if len(fields) > 0 {
			buff.WriteString(",")
		}
		// dat["id"] = account.ID
		// buff.WriteString(fmt.Sprintf("\"email\":\"%s\"", account.Email))
		//dat["sname"] = account.GetSname()
		if fields != nil {
			for m, field := range fields {
				switch field {
				case "sex":
					osex := "f"
					if account.Sex {
						osex = "m"
					}
					buff.WriteString(fmt.Sprintf("\"sex\":\"%s\"", osex))
					//dat["sex"] = osex
				case "fname":
					buff.WriteString(fmt.Sprintf("\"fname\":\"%s\"", account.GetFname()))
					//dat["fname"] = account.GetFname()
				case "sname":
					buff.WriteString(fmt.Sprintf("\"sname\":\"%s\"", account.GetSname()))
					//dat["sname"] = account.GetSname()
				case "phone":
					buff.WriteString(fmt.Sprintf("\"phone\":\"%s\"", account.Phone))
					//dat["phone"] = account.Phone
				case "city":
					city := model.DataCity.GetRev(account.City)
					buff.WriteString(fmt.Sprintf("\"city\":\"%s\"", city))
					// for k, v := range model.DataCity.GetMap() {
					// 	if v == account.City {
					// 		dat["city"] = k
					// 	}
					// }
				case "country":
					country := model.DataCountry.GetRev(account.Country)
					buff.WriteString(fmt.Sprintf("\"country\":\"%s\"", country))
					// for k, v := range model.DataCountry.GetMap() {
					// 	if v == account.Country {
					// 		dat["country"] = k
					// 	}
					// }
				case "birth":
					buff.WriteString(fmt.Sprintf("\"birth\":%d", account.Birth))
					//dat["birth"] = account.Birth
				case "joined":
					buff.WriteString(fmt.Sprintf("\"joined\":%d", account.Joined))
					//dat["joined"] = account.Joined
				case "status":
					var status string
					for k, v := range model.DataStatus {
						if v == account.Status {
							status = k
						}
					}
					buff.WriteString(fmt.Sprintf("\"status\":\"%s\"", status))
				case "premium":
					buff.WriteString(fmt.Sprintf("\"premium\":{\"start\":%d,\"finish\":%d}", account.Start, account.Finish))
					// prem := model.Premium{
					// 	Start:  int64(account.Start),
					// 	Finish: int64(account.Finish),
					// }
					// dat["premium"] = prem
				}
				if m != len(fields)-1 {
					buff.WriteString(",")
				}
			}
		}
		buff.WriteString("}")
		// tmpBuff.Reset()
		// enc.Encode(dat)
		// buff.Write(tmp[:tmpBuff.Len()])
		if i != (len(accounts) - 1) {
			buff.WriteString(",")
		}
		//out = append(out, dat)
	}
	buff.WriteString(end)
	//fmt.Println(buff.Len())
	// valid := json.Valid(bTs[:buff.Len()])
	// if !valid {
	// 	fmt.Println("Not valid")
	// }
	// //return make([]byte, 0)
	// resp["accounts"] = out
	// buff.Reset()
	// enc := json.NewEncoder(buff)
	// enc.Encode(resp)
	return bTs[:buff.Len()]
}

func toMess(m map[string]sparam) []model.Mess {
	out := make([]model.Mess, len(m))
	for k, v := range m {
		mess := model.Mess{Par: k, Val: v.par, Act: v.pred}
		out = append(out, mess)
	}
	return out
}
