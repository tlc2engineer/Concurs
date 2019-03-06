package model

import (
	"bytes"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

/*likesLoader - Загрузка like*/
func likesLoader(wg *sync.WaitGroup) chan *lInsHelp {
	ch := make(chan *lInsHelp, 100)
	var lmu = &sync.Mutex{}
	const n = 5 // число каналов
	//tmp1Likes := make([]Like, 0, 100) // для вычислений
	chans := make([]chan *lInsHelp, 0)
	for i := 0; i < n; i++ {
		chans = append(chans, addOne(wg, lmu))
	}
	go func() {
		count := 0
		for h := range ch { // принимаем контейнер
			ch := chans[count]
			ch <- h
			count++
			if count >= n {
				count = 0
			}
		}
		for i := 0; i < n; i++ {
			close(chans[i])
		}
	}()
	return ch
}

func addOne(wg *sync.WaitGroup, lmu *sync.Mutex) chan *lInsHelp {
	out := make(chan *lInsHelp, 3)
	tmpLikes := likeBuff.Get().([]Like)
	go func() {
		for h := range out {
			id := h.id
			data := h.data
			tmpLikes = tmpLikes[:0]
			//digits := make([]byte, 0, 20)
			jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				//fmt.Println(string(value))
				lid := int64(getInt(value, "id"))
				ts := int64(getInt(value, "ts"))
				tmpLikes = append(tmpLikes, Like{ID: lid, Ts: float64(ts), Num: 0})
			})
			tl := NormLikes(tmpLikes)
			lmu.Lock()
			SetLikes(uint32(id), PackLSlice(tl))
			for _, like := range tl {
				AddWho(uint32(id), like)
			}
			lmu.Unlock()
			wg.Done()
		}
		likeBuff.Put(tmpLikes)
	}()
	return out
}

/*lInsHelp - вспомагательный тип для загрузки like*/
type lInsHelp struct {
	data []byte
	id   uint32
}

/*lInsHelp - вспомагательный тип для загрузки User*/
type uInsHelp struct {
	id                                                     uint32
	country, city, email, phone, sex, status, sname, fname string
	interests                                              []string
	birth, joined, start, finish                           uint32
}

func userLoader(wg *sync.WaitGroup) chan *uInsHelp {
	ch := make(chan *uInsHelp, 100)
	go func() {
		for h := range ch {
			id := h.id
			country := h.country
			city := h.city
			fname := h.fname
			sname := h.sname
			interests := h.interests
			birth := h.birth
			joined := h.joined
			phone := h.phone
			email := h.email
			sex := h.sex
			status := h.status
			finish := h.finish
			start := h.start
			//---------------------------
			countryV := DataCountry.GetOrAdd(country)
			countryMap.Add(uint32(countryV), uint32(id))
			//-----------------------------------
			cityV := DataCity.GetOrAdd(city)
			cityMap.Add(uint32(cityV), uint32(id))
			//-------------------------------------
			fnameV := DataFname.GetOrAdd(fname)
			fnameMap.Add(uint32(fnameV), uint32(id))
			//--------------------------------------
			snameV, ok := DataSname[sname]
			if !ok {
				if sname == "" {
					DataSname[sname] = 0
					RSname[0] = ""
				} else {
					ln := uint32(len(DataSname) + 1)
					DataSname[sname] = ln
					RSname[ln] = sname
					snameV = ln

				}
			}
			snameIndex.Add(snameV, uint32(id))
			//-------------------------------------------
			interestsV := GetInterests(interests)
			for _, inter := range interestsV {
				intMap.Add(uint32(inter), uint32(id))
			}
			//-------------------------------------------
			code := GetCode(phone)
			codeIndex.Add(uint32(code), uint32(id))
			//-------------------------------------------
			if birth != 0 {
				date := time.Unix(int64(birth), 0).In(Loc)
				year := date.Year()
				bYearIndex.Add(uint32(year), uint32(id))
			}
			if joined != 0 {
				date := time.Unix(int64(joined), 0).In(Loc)
				year := date.Year()
				jYearIndex.Add(uint32(year), uint32(id))
			}
			//--------------------------------------------
			domain := getDomain(email)
			domIndex.Add(uint32(domain), uint32(id))
			//---------------------------------------------
			user := User{
				ID:        uint32(id),
				Email:     email,
				Domain:    domain,
				FName:     fnameV,
				SName:     snameV,
				Phone:     phone,
				Code:      code,
				Sex:       sex == "m",
				Birth:     uint32(birth),
				Country:   countryV,
				City:      cityV,
				Joined:    uint32(joined),
				Status:    DataStatus[status],
				Interests: interestsV,
				Start:     uint32(start),
				Finish:    uint32(finish),
			}
			//-------------------------------
			AddGIndex(user)
			AddRecIndex(user)
			ln := len(users)
			//---------Имя и пол------------------------
			addSexName(fnameV, sex == "m")
			//---------Общие интересы-------------------------
			setCommonInt(interestsV)
			//-------------------------------------------
			if ln > 0 && user.ID < users[ln-1].ID { // если больше последнего элемента добавляем в конец иначе вставка
				i := ln - 1 // начиная с последнего элемента
				for i >= 0 && user.ID < users[ln-1].ID {
					i--
				}
				//вставка
				users = append(users, User{})
				copy(users[i+2:], users[i+1:]) // вставляем id перед i
				users[i+1] = user
			} else {
				users = append(users, user)
			}
			//ch <- h // отдаем контейнер обратно
			wg.Done()
		}
	}()
	return ch
}

func parseAccountLoader(chelp chan *uInsHelp) chan []byte {
	cbyte := make(chan []byte, 10)
	parseFunc := func() (chan []byte, chan *uInsHelp) {
		in := make(chan []byte)
		out := make(chan *uInsHelp)
		go func() {
			for value := range in {
				//getEmail(value)
				id, _ := jsonparser.GetInt(value, "id")
				birth := getInt(value, "birth") //jsonparser.GetInt(value, "birth")
				// b2, _ := jsonparser.GetInt(value, "birth")
				// if uint32(b2) != birth && id == 1 {
				// 	fmt.Println("compare", b2, birth)
				// 	//panic("Not")
				// }
				joined := getInt(value, "joined") //jsonparser.GetInt(value, "joined")
				fname, _ := jsonparser.GetString(value, "fname")
				sname, _ := jsonparser.GetString(value, "sname")
				//jsonparser.GetString(value, "email")
				email := getStringV(value, "email")
				//phone, _ := jsonparser.GetString(value, "phone")
				phone := getStringV(value, "phone")
				country, _ := jsonparser.GetString(value, "country")
				city, _ := jsonparser.GetString(value, "city")
				sex := getStringV(value, "sex") //jsonparser.GetString(value, "sex")
				status, _ := jsonparser.GetString(value, "status")
				start := getInt(value, "start")   //jsonparser.GetInt(value, "premium", "start")
				finish := getInt(value, "finish") //jsonparser.GetInt(value, "premium", "finish")
				interestsTmp := make([]string, 0)
				jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					s, _ := jsonparser.ParseString(value)
					interestsTmp = append(interestsTmp, s)
				}, "interests")
				user := uInsHelp{}
				user.interests = interestsTmp
				user.id = uint32(id)
				user.email = email
				user.phone = phone
				user.birth = uint32(birth)
				user.joined = uint32(joined)
				user.fname = fname
				user.sname = sname
				user.city = city
				user.country = country
				user.sex = sex
				user.status = status
				user.start = uint32(start)
				user.finish = uint32(finish)

				out <- &user
			}
			close(out)
		}()

		return in, out
	}
	in1, out1 := parseFunc()
	in2, out2 := parseFunc()
	in3, out3 := parseFunc()
	in4, out4 := parseFunc()
	go func() {
		count := 0
		for h := range cbyte {
			switch count {
			case 0:
				in1 <- h
				count = 1
			case 1:
				in2 <- h
				count = 2
			case 2:
				in3 <- h
				count = 3
			case 3:
				in4 <- h
				val1 := <-out1
				val2 := <-out2
				val3 := <-out3
				val4 := <-out4
				chelp <- val1
				chelp <- val2
				chelp <- val3
				chelp <- val4
				count = 0
			}
		}
		close(chelp)
	}()
	// go func() {
	// 	parseFunc()
	// }()
	// go func() {
	// 	parseFunc()
	// }()

	return cbyte
}

/*getId - получение ID*/
func getStringV(val []byte, name string) string {
	fbyte := []byte(name)
	fbyte = append(fbyte, 34, 58) //":
	ind := bytes.Index(val, fbyte)
	if ind == -1 {
		return ""
	}
	i := ind + 4 + len(name)
	bg := i
	for {
		if val[i] == 34 {
			break
		}
		i++
	}
	s := string(val[bg:i])
	//fmt.Println(s)
	return s

}

/*getInt - получение целого*/
func getInt(val []byte, name string) uint32 {
	fbyte := []byte(name)
	fbyte = append(fbyte, 34, 58) //":
	ind := bytes.Index(val, fbyte)
	if ind == -1 {
		return 0
	}
	i := ind + 3 + len(name)
	beg := i
	var out int
	for val[i] >= 48 && val[i] < 59 {
		i++
	}
	end := i
	mul := 1
	for i := end - 1; i >= beg; i-- {
		out += int(val[i]-48) * mul
		mul = mul * 10
	}
	return uint32(out)
}
