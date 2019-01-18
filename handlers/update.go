package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

const queryParam = "query_id"

/*Update - обновление аккаунта*/
func Update(ctx *fasthttp.RequestCtx, id int) {
	mutex := model.WrMutex
	mutex.Lock()
	defer mutex.Unlock()
	if !ctx.QueryArgs().Has("query_id") {
		fmt.Println("Нет query_id")
		ctx.SetStatusCode(400)
		return
	}
	paccount, err := model.GetAccountPointer(uint32(id))
	if err != nil {
		fmt.Println("Нет такого аккаунта")
		ctx.SetStatusCode(404)
		return
	}
	data := ctx.PostBody()
	dat := make(map[string]interface{})
	err = json.Unmarshal(data, &dat)
	if err != nil {
		fmt.Println(err)
		ctx.SetStatusCode(400)
		return
	}
	//fmt.Println("Данные", dat)
	vmap := make(map[string]bool)
	// переменные
	var email, sname, fname, phone, sex, country, city, status string
	var birth, joined, start, finish int64
	var likes []model.Like
	var interests []string

	verify := func(dat map[string]interface{}) error {
		for k, v := range dat {
			switch k {
			case "email":
				email, err = verSPar(v, 100)
				if err != nil {
					return err
				}
				if model.IsMailExist(email) {
					return fmt.Errorf("Такой mail уже есть")
				}
				if !strings.Contains(email, "@") {
					return fmt.Errorf("Неправильный mail")
				}
			case "phone":
				phone, err = verSPar(v, 100)
				if err != nil {
					return err
				}
				//model.UpdatePhone(phone, paccount.Phone)
			case "fname":
				fname, err = verSPar(v, 50)
				if err != nil {
					return err
				}
				//paccount.FName = fname
			case "sname":
				sname, err = verSPar(v, 50)
				if err != nil {
					return err
				}
				//paccount.SName = sname
			case "sex":
				sex, err = verSPar(v, 1)
				if err != nil {
					return err
				}
				if sex != "m" && sex != "f" {
					return fmt.Errorf("Неправильный пол " + sex)
				}
			case "birth":
				birth, err = verifyDPar(v, LBtime, HBtime)
				if err != nil {
					return err
				}
				//paccount.Birth = birth
			case "country":
				country, err = verSPar(v, 50)
				if err != nil {
					return err
				}
				//paccount.Country = country
			case "city":
				city, err = verSPar(v, 50)
				if err != nil {
					return err
				}
				//paccount.City = city
			case "joined":
				joined, err = verifyDPar(v, LRTime, HRTime)
				if err != nil {
					return err
				}
				//paccount.Joined = joined
			case "status":
				status, err = verSPar(v, 50)
				if err != nil {
					return err
				}
				switch status {
				case "свободны":
				case "заняты":
				case "всё сложно":
				default:
					return fmt.Errorf("Непонятный статус")
				}
				//paccount.Status = status
			case "interests":
				dat, ok := v.([]interface{})
				if !ok {
					return fmt.Errorf("Неправильное значение интересов %v", v)
				}
				interests = make([]string, 0)
				for i := range dat {
					s, ok := dat[i].(string)
					if !ok {
						return fmt.Errorf("Неправильное значение интересов %v", v)
					}
					if len(s) > 100 {
						return fmt.Errorf("Превышение длины интереса")
					}
					interests = append(interests, s)
				}
			case "premium":
				dat, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("Неправильное значение премиум 1  %v", v)
				}
				sval, ok := dat["start"]
				if !ok {
					return fmt.Errorf("Неправильное значение премиум 2  %v", v)
				}
				fval, ok := dat["finish"]
				if !ok {
					return fmt.Errorf("Неправильное значение премиум 3  %v", v)
				}

				tstart, ok := sval.(float64)
				if !ok {
					return fmt.Errorf("Неправильное значение премиум 4  %v", v)
				}
				tfinish, ok := fval.(float64)
				if !ok {
					return fmt.Errorf("Неправильное значение премиум 5  %v", v)
				}

				start = int64(tstart)
				finish = int64(tfinish)

				if start < LPTime {
					return fmt.Errorf("Неправильное значение премиум 6")
				}
				if finish < LPTime {
					return fmt.Errorf("Неправильное значение премиум 7")
				}
			case "likes":
				out := make([]model.Like, 0)
				likesMap, ok := v.([]map[string]interface{})
				if !ok {
					return fmt.Errorf("Неправильное преобразование likes")
				}
				for _, like := range likesMap {
					idv, ok := like["id"]
					if !ok {
						return fmt.Errorf("Неправильное преобразование likes")
					}
					id, ok := idv.(int64)
					if !ok {
						return fmt.Errorf("Неправильное преобразование likes")
					}
					tsv, ok := like["ts"]
					if !ok {
						return fmt.Errorf("Неправильное преобразование likes")
					}
					ts, ok := tsv.(float64)
					if !ok {
						return fmt.Errorf("Неправильное преобразование likes")
					}
					nlike := model.Like{Ts: ts, ID: id}
					out = append(out, nlike)
				}
				likes = out
			default:
				return fmt.Errorf("Неизвестное поле")

			}
			vmap[k] = true
		}
		return nil
	}
	err = verify(dat)
	if err != nil {
		fmt.Println(err)
		ctx.SetStatusCode(400)
		return
	}
	// Присвоение значений
	for k := range vmap {
		switch k {
		case "email":
			oldMail := paccount.Email
			paccount.Email = email
			model.UpdateEmail(email, oldMail)
		case "phone":
			old := paccount.Phone
			paccount.Phone = phone
			model.UpdatePhone(phone, old)
		case "fname":
			model.UpdFname(paccount.ID, paccount.FName, fname)
			paccount.SetFname(fname)
		case "sname":
			paccount.SetSname(sname)
			model.UpdSname(paccount.ID, paccount.SName, sname)
		case "sex":
			bsex := false
			if sex == "m" {
				bsex = true
			}
			paccount.Sex = bsex
		case "birth":
			paccount.Birth = uint32(birth)
		case "country":
			model.UpdICountry(paccount.ID, paccount.Country, country) // обновить индекс
			paccount.Country = model.DataCountry.GetOrAdd(country)
		case "city":
			model.UpdICity(paccount.ID, paccount.City, city) // обновить индекс
			paccount.City = model.DataCity.GetOrAdd(city)
		case "joined":
			paccount.Joined = uint32(joined)
		case "status":
			paccount.Status = model.DataStatus[status]
		case "interests":
			oldI := paccount.Interests                   // старые интересы
			model.UpdInter(paccount.ID, oldI, interests) // обновить индекс
			paccount.Interests = model.GetInterests(interests)
		case "premium":
			paccount.Start = uint32(start)
			paccount.Finish = uint32(finish)
		case "likes":
			likes := model.NormLikes(likes) // нормируем
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
				if tid == 31629 {
					fmt.Println("Вот он!")
				}
				data, _ := model.GetWho(tid) // кто лайкал данный id
				model.SetWho(uint32(tid), data.RemoveId(uint32(id)))
			}

		}
	}
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
