package model

import (
	"sort"
	"strconv"
	"strings"
)

/*Mess - сообщение из обработчика*/
type Mess struct {
	Par string
	Val string
	Act string
}

const cnt = 100

/*IndexAgg - возвращает список элементов в зависимости от индексов*/
func IndexAgg(mess []Mess, fncs []func(User) bool, limit int, uBuff []*User) []*User {
	//------------------------------
	indexes := make([]indexLogic, 0)
	for _, v := range mess {
		par := v.Par
		switch par {
		case "city":
			switch v.Act {
			case "eq":
				city := v.Val
				num, ok := DataCity.Get(city)
				if !ok {
					return make([]*User, 0) // пустой массив
				}
				all, err := cityMap.GetAll(uint32(num))
				if err != nil {
					//fmt.Println("Нет такого города ", num)
					return make([]*User, 0) // пустой массив
				}
				indexes = append(indexes, newSLogic(all.GetAll()))
			case "any":
				cities := strings.Split(v.Val, ",") // имена городов
				logics := make([]*simpleIndexLogic, 0)
				for _, city := range cities {
					num, ok := DataCity.Get(city)
					if !ok {
						continue
					}
					all, _ := cityMap.GetAll(uint32(num))
					sl := newSLogic(all.GetAll())
					logics = append(logics, sl)
				}
				indexes = append(indexes, newCmplLog(logics))
			case "null":
				if v.Val == "1" { // все без города
					all, _ := cityMap.GetAll(0)
					indexes = append(indexes, newSLogic(all.GetAll()))
				}
			}
		case "likes":
			// карта лайков
			par := v.Val //параметр
			if par == "" {
				continue
			}
			args := strings.Split(par, ",")
			for _, p := range args {
				num, _ := strconv.ParseInt(p, 10, 0)
				ids, err := GetWho(uint32(num))
				if err != nil {
					//fmt.Println("Нет такого города ", num)
					return make([]*User, 0) // пустой массив
				}
				data := ids.Get()
				sort.Slice(data, func(i, j int) bool {
					return data[i] < data[j]
				})
				//indexD := IndSData(data)
				indexes = append(indexes, newSLogic(data))
			}
		case "country":
			switch v.Act {
			case "eq":
				country := v.Val
				num, ok := DataCountry.Get(country)
				if !ok {
					return make([]*User, 0) // пустой массив
				}
				all, err := countryMap.GetAll(uint32(num))
				if err != nil {
					//fmt.Println("Нет такой страны ", num)
					return make([]*User, 0) // пустой массив
				}
				indexes = append(indexes, newSLogic(all.GetAll()))
			case "null":
				if v.Val == "1" { // все без города
					all, _ := countryMap.GetAll(0)
					indexes = append(indexes, newSLogic(all.GetAll()))
				}
			}
		case "interests":
			par := v.Val
			if par == "" {
				return make([]*User, 0) // пустой массив
			}
			pari := strings.Split(par, ",")
			dat := make([]uint16, len(pari))
			for i := range pari {
				dat[i], _ = DataInter.Get(pari[i])
			}
			switch v.Act {
			case "contains":
				if len(dat) > 1 && exceptInterests(dat) { // взаимоисключающие интересы
					return make([]*User, 0) // пустой массив
				}
				for _, inter := range dat {
					idata, err := intMap.GetAll(uint32(inter))
					if err != nil {
						return make([]*User, 0)
					}
					indexes = append(indexes, newSLogic(idata.GetAll()))
				}
			case "any":
				logics := make([]*simpleIndexLogic, 0)
				for _, inter := range dat {
					idata, err := intMap.GetAll(uint32(inter))
					if err != nil {
						return make([]*User, 0)
					}
					sl := newSLogic(idata.GetAll())
					logics = append(logics, sl)
				}
				indexes = append(indexes, newCmplLog(logics))
			}
		case "fname":
			switch v.Act {
			case "eq":
				fname := v.Val
				num, ok := DataFname.Get(fname)
				if !ok {
					return make([]*User, 0) // пустой массив
				}
				for _, msg := range mess { // есть ли проверка по полу
					if msg.Par == "sex" {
						sex := msg.Val == "m"
						sn, ok := sexNames[num] // смотрим в словаре
						if !ok {
							return make([]*User, 0)
						}
						if sex && sn == 0 || !sex && sn == 1 { //не тот пол
							return make([]*User, 0)
						}
						break
					}
				}
				all, err := fnameMap.GetAll(uint32(num))
				if err != nil {
					return make([]*User, 0) // пустой массив
				}
				indexes = append(indexes, newSLogic(all.GetAll()))
			case "any":
				names := strings.Split(v.Val, ",") // имена
				logics := make([]*simpleIndexLogic, 0)
				checkSex := false
				sex := false
				sexValid := false
				//fmt.Println(mess)
				for _, msg := range mess { // есть ли проверка по полу
					if msg.Par == "sex" {
						checkSex = true
						sex = msg.Val == "m"
						break
					}
				}
				for _, name := range names {
					num, ok := DataFname.Get(name)
					if !ok {
						return make([]*User, 0)
					}
					if checkSex && !sexValid {
						sn, ok := sexNames[num] // смотрим в словаре
						if !ok {
							return make([]*User, 0)
						}
						if sex && sn == 1 || !sex && sn == 0 { //тот пол
							sexValid = true
						}
					}
					all, _ := fnameMap.GetAll(uint32(num))
					sl := newSLogic(all.GetAll())
					logics = append(logics, sl)
				}
				if checkSex && !sexValid { // пол имени и параметр sex не совпадают
					return make([]*User, 0)
				}
				indexes = append(indexes, newCmplLog(logics))
			case "null":
				if v.Val == "1" { // все без имени
					all, _ := fnameMap.GetAll(0)
					indexes = append(indexes, newSLogic(all.GetAll()))
				}
			}

		case "sname":
			switch v.Act {
			case "eq":
				sname := v.Val
				num, ok := DataSname[sname]
				if !ok {
					return make([]*User, 0) // пустой массив
				}
				all, err := snameIndex.GetAll(uint32(num))
				if err != nil {
					//fmt.Println("Нет такого города ", num)
					return make([]*User, 0) // пустой массив
				}
				indexes = append(indexes, newSLogic(all.GetAll()))
			case "starts":
				pref := v.Val
				nums := make([]uint32, 0)
				for sname := range DataSname {
					if strings.HasPrefix(sname, pref) {
						nums = append(nums, DataSname[sname])
					}
				}
				logics := make([]*simpleIndexLogic, 0)
				for _, num := range nums {
					ind, err := snameIndex.GetAll(num)
					if err != nil {
						return make([]*User, 0) // пустой массив
					}
					data := ind.GetAll()
					sl := newSLogic(data)
					logics = append(logics, sl)

				}
				indexes = append(indexes, newCmplLog(logics))
			case "null":
				if v.Val == "1" { // все без имени
					all, _ := snameIndex.GetAll(0)
					indexes = append(indexes, newSLogic(all.GetAll()))
				}
			}
		case "phone":
			switch v.Act {
			case "code":
				scode := v.Val
				code, err := strconv.ParseInt(scode, 10, 0)
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				ind, _ := codeIndex.GetAll(uint32(code))
				indexes = append(indexes, newSLogic(ind.GetAll()))
			}

		case "birth":
			switch v.Act {
			case "year":
				sy := v.Val
				year, err := strconv.ParseInt(sy, 10, 0)
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				ind, err := bYearIndex.GetAll(uint32(year))
				//fmt.Println(year, ind.Len())
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				indexes = append(indexes, newSLogic(ind.GetAll()))
			}

		case "joined":
			switch v.Act {
			case "year":
				sy := v.Val
				year, err := strconv.ParseInt(sy, 10, 0)
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				ind, err := jYearIndex.GetAll(uint32(year))
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				indexes = append(indexes, newSLogic(ind.GetAll()))
			}
		case "email":
			switch v.Act {
			case "domain":
				dom := v.Val
				num, ok := DataDomain.Get(dom)
				if !ok {
					return make([]*User, 0) // пустой массивs
				}
				ind, err := domIndex.GetAll(uint32(num))
				if err != nil {
					return make([]*User, 0) // пустой массивs
				}
				indexes = append(indexes, newSLogic(ind.GetAll()))
			}

		}
	}

	if len(indexes) == 0 { // нет индексов
		return nil
	}
	uBuff = uBuff[:0]
	_users := uBuff       // make([]User, 0, limit)
	if len(indexes) > 1 { // больше одного индекса
		var base indexLogic
		minLen := indexes[0].len()
		bnum := 0
		for i := 0; i < len(indexes); i++ { // находим индекс с минимальной длиной
			if indexes[i].len() < minLen {
				minLen = indexes[i].len()
				bnum = i
			}
		}
		base = indexes[bnum]                                            // выходной срез максимальная длина минимальый индекс
		indexes = indexes[:bnum+copy(indexes[bnum:], indexes[bnum+1:])] // удаляем базовый индекс из списка
	m1:
		for {
			val := base.nextIt()
			if val == -1 {
				break
			}
			for i := 0; i < len(indexes); i++ {
				if !indexes[i].nextCheck(uint32(val)) {
					continue m1
				}
			}
			_user := GetUser(uint32(val))
			for _, fn := range fncs {
				if !fn(*_user) {
					continue m1
				}
			}
			_users = append(_users, _user)
			if len(_users) >= limit {
				break
			}
		}
		return _users

	}
m2:
	for {
		val := indexes[0].nextIt()
		if val == -1 {
			break
		}
		_user := GetUser(uint32(val))
		for _, fn := range fncs {
			if !fn(*_user) {
				continue m2
			}
		}
		_users = append(_users, _user)
		if len(_users) >= limit {
			break
		}
	}
	return _users

}
