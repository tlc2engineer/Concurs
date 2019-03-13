package handlers

import (
	"Concurs/model"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"

	"github.com/valyala/fasthttp"
)

const queryParam = "query_id"

var wg = &sync.WaitGroup{}

/*Update - обновление аккаунта*/
func Update(ctx *fasthttp.RequestCtx, id int) {
	mutex.Lock()
	defer mutex.Unlock()
	if !ctx.QueryArgs().Has("query_id") {
		//fmt.Println("Нет query_id")
		ctx.SetStatusCode(400)
		return
	}
	paccount, err := model.GetAccountPointer(uint32(id))
	if err != nil {
		//fmt.Println("Нет такого аккаунта")
		ctx.SetStatusCode(404)
		return
	}
	data := ctx.PostBody()
	// dat := make(map[string]interface{})
	// err = json.Unmarshal(data, &dat)
	// if err != nil {
	// 	//fmt.Println(err)
	// 	ctx.SetStatusCode(400)
	// 	return
	// }
	//fmt.Println("Данные", dat)
	vmap := make(map[string]bool)
	// переменные
	var email, sname, fname, phone, sex, country, city, status string
	var birth, joined, start, finish int64
	var likes []model.Like
	var interests []string
	err = jsonparser.ObjectEach(data, func(bkey []byte, v []byte, dataTp jsonparser.ValueType, offset int) error {
		key, err := jsonparser.ParseString(bkey)
		switch key {
		case "email":
			email, _ = jsonparser.ParseString(v) //value
			if len(email) > 100 {
				return err
			}
			if model.IsMailExist(email) {
				return fmt.Errorf("Такой mail уже есть")
			}
			if !strings.Contains(email, "@") {
				return fmt.Errorf("Неправильный mail")
			}
		case "phone":
			phone, _ = jsonparser.ParseString(v)
			if len(phone) > 100 {
				return err
			}
		case "fname":
			fname, _ = jsonparser.ParseString(v)
			if len(fname) > 50 {
				return err
			}
		case "sname":
			sname, _ = jsonparser.ParseString(v)
			if len(sname) > 50 {
				return err
			}
		case "sex":
			sex, _ = jsonparser.ParseString(v)
			if sex != "m" && sex != "f" {
				return fmt.Errorf("Неправильный пол " + sex)
			}
		case "country":
			country, _ = jsonparser.ParseString(v)
			if len(country) > 50 {
				return err
			}
			//paccount.Country = country
		case "city":
			city, _ = jsonparser.ParseString(v)
			if len(city) > 50 {
				return err
			}
		case "status":
			status, _ = jsonparser.ParseString(v) //value
			switch status {
			case "свободны":
			case "заняты":
			case "всё сложно":
			default:
				return fmt.Errorf("Непонятный статус")
			}
		case "interests":
			interests = make([]string, 0)
			errFlag := false
			jsonparser.ArrayEach(v, func(value1 []byte, dataType jsonparser.ValueType, offset int, err error) {
				if err != nil {
					errFlag = true
				}
				interes, _ := jsonparser.ParseString(value1)
				if len(interes) > 100 {
					errFlag = true
				}
				interests = append(interests, interes)
			})
			//fmt.Println(interests)
			if errFlag {
				return fmt.Errorf("Неправильное значение интересов")
			}
		case "premium":
			checkCount := 0
			jsonparser.ObjectEach(v, func(bkey1 []byte, v1 []byte, dataTp jsonparser.ValueType, offset int) error {
				switch string(bkey1) {
				case "start":
					start, err = jsonparser.ParseInt(v1)
					if err != nil || int64(start) < LPTime {
						return fmt.Errorf("Неправильное значение премиум   %v", v)
					}
					checkCount++
				case "finish":
					finish, err = jsonparser.ParseInt(v1)
					if err != nil || int64(finish) < LPTime {
						return fmt.Errorf("Неправильное значение премиум   %v", v)
					}
					checkCount++
				default:
					return fmt.Errorf("Неправильное поле")
				}
				return nil
			})
			if checkCount != 2 {
				return fmt.Errorf("Неправильное значение premium")
			}
		case "birth":
			birth, err = jsonparser.ParseInt(v)
			//birth, err = verifyDPar(v, LBtime, HBtime)
			if err != nil {
				return err
			}
			if int64(birth) < LBtime || int64(birth) > HBtime {
				return fmt.Errorf("Неправильное предела даты")
			}
		case "joined":
			joined, err = jsonparser.ParseInt(v)
			if err != nil {
				return err
			}
			if int64(joined) < LRTime || int64(joined) > HRTime {
				return fmt.Errorf("Неправильное предела даты")
			}
		case "likes":
			likes = make([]model.Like, 0)
			errFlag := false
			jsonparser.ArrayEach(v, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				ts, err := jsonparser.GetInt(value, "ts")
				if err != nil {
					errFlag = true
				}
				id, err := jsonparser.GetInt(value, "id")
				if err != nil {
					errFlag = true
				}
				if !errFlag {
					likes = append(likes, model.Like{Ts: float64(ts), ID: id, Num: 1})
				}
			})
			if errFlag {
				return fmt.Errorf("Неправильное преобразование likes")
			}
		}
		vmap[key] = true
		return nil
	})
	if err != nil {
		//fmt.Println(err)
		ctx.SetStatusCode(400)
		return
	}
	wg.Add(2)
	// удаление старого индекса
	go func() {
		model.DeleteGIndex(*paccount)
		wg.Done()
	}()
	go func() {
		model.RemRecIndex(*paccount)
		wg.Done()
	}()
	wg.Wait()
	// Присвоение значений
	for k := range vmap {
		switch k {
		case "email":
			wg.Add(1)
			go func() {
				oldMail := paccount.Email
				paccount.Email = email
				model.UpdateEmail(email, oldMail)
				model.UpdateDomainInd(paccount.ID, oldMail, email)
				wg.Done()
			}()
		case "phone":
			old := paccount.Phone
			oldCode := model.GetCode(old)
			newCode := model.GetCode(phone)
			paccount.Phone = phone
			model.UpdatePhone(phone, old)
			model.UpdCode(paccount.ID, oldCode, newCode)
		case "fname":
			wg.Add(1)
			go func() {
				model.UpdFname(paccount.ID, paccount.FName, fname)
				paccount.SetFname(fname)
				wg.Done()
			}()
		case "sname":
			wg.Add(1)
			go func() {
				oldName := paccount.SName
				paccount.SetSname(sname)
				model.UpdSname(paccount.ID, oldName, sname)
				wg.Done()
			}()
		case "sex":
			bsex := false
			if sex == "m" {
				bsex = true
			}
			paccount.Sex = bsex
		case "birth":
			oldBirth := paccount.Birth
			oldDate := time.Unix(int64(oldBirth), 0).In(loc)
			oldYear := oldDate.Year()
			newDate := time.Unix(int64(birth), 0).In(loc)
			newYear := newDate.Year()
			paccount.Birth = uint32(birth)
			model.UpdateBYear(uint32(paccount.ID), uint32(oldYear), uint32(newYear))
		case "country":
			wg.Add(1)
			go func() {
				model.UpdICountry(paccount.ID, paccount.Country, country) // обновить индекс
				paccount.Country = model.DataCountry.GetOrAdd(country)
				wg.Done()
			}()
		case "city":
			wg.Add(1)
			go func() {
				model.UpdICity(paccount.ID, paccount.City, city) // обновить индекс
				paccount.City = model.DataCity.GetOrAdd(city)
				wg.Done()
			}()
		case "joined":
			oldBirth := paccount.Joined
			oldDate := time.Unix(int64(oldBirth), 0).In(loc)
			oldYear := oldDate.Year()
			newDate := time.Unix(int64(joined), 0).In(loc)
			newYear := newDate.Year()
			paccount.Joined = uint32(joined)
			model.UpdateJYear(uint32(paccount.ID), uint32(oldYear), uint32(newYear))
		case "status":
			paccount.Status = model.DataStatus[status]
		case "interests":
			wg.Add(1)
			go func() {
				oldI := paccount.Interests                   // старые интересы
				model.UpdInter(paccount.ID, oldI, interests) // обновить индекс
				paccount.Interests = model.GetInterests(interests)
				wg.Done()
			}()
		case "premium":
			paccount.Start = uint32(start)
			paccount.Finish = uint32(finish)
		case "likes":
			wg.Add(1)
			go func() {
				likes := model.NormLikes(likes) // нормируем
				sort.Slice(likes, func(i, j int) bool {
					return likes[i].ID < likes[j].ID
				})
				model.SetLikes(uint32(id), model.PackLSlice(likes))
				model.AddWhos(uint32(id), likes)
				//Удалить старые лайки которых уже нет!
				oldLikes := model.GetLikes(uint32(id)) // старые лайки
				ids := make([]uint32, 0)               // список несовпадающих лайков
				for i := 0; i < len(oldLikes)/8; i++ {
					var idLike uint32 // старый id
					idLike = uint32(oldLikes[0]) | uint32(oldLikes[1])<<8 | uint32(oldLikes[2])<<16
					found := false
					for _, like := range likes {
						if uint32(like.ID) == idLike {
							found = true
							break
						}
					}
					if !found {
						ids = append(ids, idLike)
					}
				}
				// Цикл по номерам которые уже не предпочитает
				for _, tid := range ids {
					data, _ := model.GetWho(tid) // кто лайкал данный id
					model.SetWho(uint32(tid), data.RemoveId(uint32(id)))
				}
				wg.Done()
			}()

		}
	}
	wg.Wait()
	wg.Add(2)
	func() {
		model.AddGIndex(*paccount)
		wg.Done()
	}()
	func() {
		model.AddRecIndex(*paccount)
		wg.Done()
	}()
	wg.Wait()
	ctx.SetStatusCode(202) // все в норме
	ctx.Write([]byte(""))
	return
}

/*verSPar - проверка строкового параметра по длине*/
func verSPar(v interface{}, lim int) (string, error) {
	val, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("не строка")
	}
	if len(val) > lim {
		return "", fmt.Errorf("превышение длины")
	}
	return val, nil
}

/*verifyDPar - проверка даты*/
func verifyDPar(v interface{}, l int64, h int64) (int64, error) {
	val, ok := v.(int64)
	if !ok {
		return -1, fmt.Errorf("Неправильное значение даты")
	}
	if l != -1 && val < l {
		return -1, fmt.Errorf("Нарушение нижнего предела даты")
	}
	if h != -1 && val > h {
		return -1, fmt.Errorf("Нарушение верхнего предела даты")
	}
	return val, nil
}
