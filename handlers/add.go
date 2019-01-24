package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

/*LBtime - нижний предел даты рождения*/
var LBtime int64

/*HBtime - верхний предел даты рождения*/
var HBtime int64

/*LRTime - нижний предел регистрации*/
var LRTime int64

/*HRTime - верхний предел регистрации*/
var HRTime int64

/*LPTime - нижняя граница премиум*/
var LPTime int64

var zeroOut []byte

func init() {
	ttime, _ := time.Parse(time.RFC3339, "1950-01-01T00:00:00Z")
	LBtime = ttime.Unix()
	htime, _ := time.Parse(time.RFC3339, "2005-01-01T00:00:00Z")
	HBtime = htime.Unix()
	trtime, _ := time.Parse(time.RFC3339, "2011-01-01T00:00:00Z")
	LRTime = trtime.Unix()
	hrtime, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00Z")
	HRTime = hrtime.Unix()
	lptime, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00Z")
	LPTime = lptime.Unix()
	resp := make(map[string][]map[string]interface{})
	out := make([]map[string]interface{}, 0, 0)
	resp["accounts"] = out
	zeroOut, _ = json.Marshal(resp)
}

/*Add - добавление нового аккаунта*/
func Add(ctx *fasthttp.RequestCtx) {
	mutex := model.WrMutex
	mutex.Lock()
	defer mutex.Unlock()

	if !ctx.QueryArgs().Has("query_id") {
		ctx.SetStatusCode(400)
		return
	}
	data := ctx.PostBody()
	acc := new(model.Account)
	err := json.Unmarshal(data, acc)
	if err != nil {
		//fmt.Println("Ошибка распаковки", err)
		ctx.SetStatusCode(400)
		return
	}
	err = verifyAccount(acc)
	if err != nil {
		//fmt.Println("Ошибка верификации", err)
		ctx.SetStatusCode(400)
		return
	}
	likes := acc.Likes
	likes = model.NormLikes(likes)
	acc.Likes = likes
	inLikes := model.PackLSlice(likes)
	model.SetLikes(uint32(acc.ID), inLikes)
	model.AddWhos(uint32(acc.ID), likes)
	model.AddAcc(model.Conv(*acc))
	ctx.SetStatusCode(201) // все в норме
	ctx.Write([]byte(""))
	return
}

/*verifyAccount - проверка на достоверность*/
func verifyAccount(acc *model.Account) error {
	if acc.ID == 0 {
		return fmt.Errorf("Нет id")
	}
	// проверка id
	_, err := model.GetAccount(uint32(acc.ID))
	if err == nil {
		return fmt.Errorf("Такой id уже есть")
	}
	// проверка пола
	if (acc.Sex != "m" && acc.Sex != "f") || acc.Sex == "" {
		return fmt.Errorf("Неправильный пол " + acc.Sex)
	}
	// проверка email
	id := model.GetAccMail(acc.Email)
	if id != -1 {
		return fmt.Errorf("email уже существует " + acc.Email)
	}
	if len(acc.Email) > 100 {
		return fmt.Errorf("Большой email")
	}
	if len(acc.FName) > 50 || len(acc.SName) > 50 {
		return fmt.Errorf("Большая длина имени или фамилии")
	}
	if acc.Phone != "" {
		if len(acc.Phone) > 16 {
			return fmt.Errorf("Большая длина телефона")
		}
		id = model.GetAccPhone(acc.Phone)
		if id != -1 {
			return fmt.Errorf("Такой телефон уже существует " + acc.Phone)
		}

	}
	if acc.Birth == 0 {
		return fmt.Errorf("Нет даты рождения")
	}
	if acc.Birth < LBtime || acc.Birth > HBtime {
		return fmt.Errorf("Дата рождения вне предела")
	}
	if len(acc.City) > 50 || len(acc.Country) > 50 {
		return fmt.Errorf("Длина страны или города превышаетт предел")
	}
	if acc.Joined < LRTime || acc.Joined > HRTime {
		return fmt.Errorf("Дата регистрации вне предела")
	}
	// проверка статуса
	switch acc.Status {
	case "свободны":
	case "заняты":
	case "всё сложно":
	default:
		return fmt.Errorf("Неизвестный статус")
	}
	// проверка интересов
	for _, interes := range acc.Interests {
		if len(interes) > 100 {
			return fmt.Errorf("Слишком длинное поле интереса")
		}
	}
	// проверка
	if acc.Premium.Start > 0 && acc.Premium.Finish > 0 {
		if acc.Premium.Start < LPTime || acc.Premium.Finish < LPTime {
			return fmt.Errorf("Нарушена граница Premium")
		}
	}
	// проверка Like того что id существуют
	for _, like := range acc.Likes {
		id := like.ID
		_, err = model.GetAccount(uint32(id)) //id существуют
		if err != nil {
			return err
		}
	}
	return nil
}
