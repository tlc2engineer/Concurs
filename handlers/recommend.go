package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
)

var enParam = []string{"country", "city"}

/*Recommend - рекомендуемые id*/
func Recommend(w http.ResponseWriter, r *http.Request, id int) {
	//fmt.Printf("Recommend %d", id)
	vars := r.URL.Query()
	city := ""
	country := ""
	var limit = -1
	// получение параметров и верификация
	for k, v := range vars {
		switch k {
		case "city":
			city = v[0]
			if city == "" {
				w.WriteHeader(400)
				return
			}
		case "country":
			country = v[0]
			if country == "" {
				w.WriteHeader(400)
				return
			}
		case "limit":
			num, err := strconv.ParseInt(v[0], 10, 0)
			if err != nil {
				w.WriteHeader(400)
				return
			}
			limit = int(num)
			if limit < 0 {
				w.WriteHeader(400)
				return
			}
		case "query_id":
		default: // неизвестный параметр
			fmt.Println("Непонятный заголовок " + k)
			w.WriteHeader(400)
			return
		}
	}
	var account model.Account
	// находим аккаункт
	account, err := model.GetAccount(id)
	// Если нет такого аккаунта
	if err != nil {
		w.WriteHeader(404)
		return
	}
	// фильтрация
	filtered := filterRecommend(account, country, city)
	sort.Slice(filtered, func(i, j int) bool {
		f := filtered[i]
		s := filtered[j]
		// по премиум аккаунту
		if f.IsPremium() != s.IsPremium() {
			return f.IsPremium()
		}
		// по статусу
		if f.VStatus() != s.VStatus() {
			return f.VStatus() < s.VStatus()
		}
		// общие интересы
		commf := f.GetCommInt(account)
		comms := s.GetCommInt(account)
		if commf != comms {
			return commf > comms // у кого больше общих интересов
		}
		// по разнице в возрасте
		agef := math.Abs(float64(account.Birth - f.Birth))
		ages := math.Abs(float64(account.Birth - s.Birth))
		if agef != ages {
			return agef < ages
		}
		// по id
		return f.ID < s.ID
	})

	if len(filtered) > limit && limit != -1 {
		filtered = filtered[:limit]
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(recommendOutput(filtered))
}

func filterRecommend(account model.Account, country string, city string) []model.Account {
	accounts := model.GetAccounts()
	filtered := make([]model.Account, 0)
	sex := account.Sex
	rec := "f"
	if sex == "f" {
		rec = "m"
	}
	for _, acc := range accounts {
		if acc.GetCommInt(account) == 0 {
			continue
		}
		if acc.Sex != rec {
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

func recommendOutput(accounts []model.Account) []byte {
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, len(accounts))
	for _, account := range accounts {
		dat := make(map[string]interface{})
		if account.IsPremium() {
			dat["premium"] = account.Premium
		}
		dat["email"] = account.Email
		dat["id"] = account.ID
		dat["status"] = account.Status
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
