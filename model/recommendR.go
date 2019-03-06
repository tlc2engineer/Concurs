package model

import (
	"math"
	"sort"
	"sync"
)

var bUbuff = sync.Pool{
	New: func() interface{} {
		return make([]*User, 0, 20000)
	},
}

var recSortPool = sync.Pool{
	New: func() interface{} {
		return make([]recSortedEl, 0, 20000)
	},
}

/*GetFPointers - возвращает список указактелей рекомендованных пользователей*/
func GetFPointers(id uint32, country uint16, city uint16, limit int, out []*User) []*User {
	// indData := selRecFilter(id, country, city) // фильтрация по городу выгоднее
	// if indData != nil {
	// 	return altFilter(id, indData, limit, out)
	// }
	acc := GetUser(uint32(id)) //acc := MainMap[id]
	out = out[:0]
	//out := make([]*User, 0)
	intrst := acc.Interests
	sex := acc.Sex
	rec := !sex
	addS := 0
	if rec { // если  мужчины
		addS = 6
	}
	tmp := bUbuff.Get().([]*User)
	defer bUbuff.Put(tmp)
	for i := 0; i < 6; i++ {
		indexes := make([]indexLogic, 0)
		logics := make([]*simpleIndexLogic, 0)
		tmp = tmp[:0]
		num := i + addS                // номер корзины
		for _, inter := range intrst { // составляем сумарный индекс
			//fmt.Println(num, inter)
			bucket := recIndex[uint32(inter)]
			idata := bucket[num]
			sl := newSLogic(idata.GetAll())
			logics = append(logics, sl)
		}
		//complLog := newCmplLog(logics)
		indexes = append(indexes, newCmplLog(logics))
		if country != 0 {
			all, err := countryMap.GetAll(uint32(country))
			if err != nil {
				return nil // пустой массив
			}
			indexes = append(indexes, newSLogic(all.GetAll()))
		}
		if city != 0 {
			all, err := cityMap.GetAll(uint32(city))
			if err != nil {
				return nil // пустой массив
			}
			indexes = append(indexes, newSLogic(all.GetAll()))
		}
		// если нет интересов
		if len(indexes) == 0 {
			return nil
		}
		// один интерес
		var base indexLogic
		if len(indexes) == 1 {
			base = indexes[0]
			for {
				i := base.nextIt()
				if i == -1 {
					break
				}
				acc := GetUser(uint32(i)) //MainMap[uint32(i)]
				tmp = append(tmp, acc)
			}
		} else {
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
				for _, ind := range indexes {
					if !ind.nextCheck(uint32(val)) {
						continue m1
					}
				}
				acc := GetUser(uint32(val)) // MainMap[uint32(val)]
				tmp = append(tmp, acc)
			}
		}
		sortedTmps := recSortPool.Get().([]recSortedEl)
		sortedTmps = sortedTmps[:len(tmp)]
		//sortedTmps := make([]recSortedEl, len(tmp))
		for i := range tmp {
			u := tmp[i]
			sortedTmps[i].commonInt = u.GetCommInt(*acc)
			sortedTmps[i].dt = math.Abs(float64(int64(acc.Birth) - int64(u.Birth)))
			sortedTmps[i].u = u
			// sortedTmps = append(sortedTmps, recSortedEl{
			// 	commonInt: u.GetCommInt(*acc),
			// 	dt:        math.Abs(float64(int64(acc.Birth) - int64(u.Birth))),
			// 	u:         u,
			// })
		}
		sort.Slice(sortedTmps, func(i, j int) bool {
			f := sortedTmps[i]
			s := sortedTmps[j]
			if f.commonInt != s.commonInt {
				return f.commonInt > s.commonInt
			}
			if f.dt != s.dt {
				return f.dt < s.dt
			}
			return f.u.ID < s.u.ID
		})
		// sort.Slice(tmp, func(i, j int) bool {
		// 	f := tmp[i]
		// 	s := tmp[j]
		// 	commf := f.GetCommInt(*acc)
		// 	comms := s.GetCommInt(*acc)
		// 	if commf != comms {
		// 		return commf > comms // у кого больше общих интересов
		// 	}
		// 	// по разнице в возрасте
		// 	agef := math.Abs(float64(int64(acc.Birth) - int64(f.Birth)))
		// 	ages := math.Abs(float64(int64(acc.Birth) - int64(s.Birth)))
		// 	if agef != ages {
		// 		return agef < ages
		// 	}
		// 	// по id
		// 	return f.ID < s.ID

		// })
		for i := range sortedTmps {
			out = append(out, sortedTmps[i].u)
			if len(out) >= limit {
				recSortPool.Put(sortedTmps)
				return out[:limit]
			}
		}
		recSortPool.Put(sortedTmps)
		// if len(tmp) <= limit {
		// 	out = append(out, tmp...)
		// } else {
		// 	out = append(out, tmp[:limit]...)
		// }
		// if len(out) >= limit {
		// 	return out[:limit]
		// }
	}
	return out
}

/*selRecFilter - выбор фильтра, есть ли смысл брать индекс по стране и городу*/
func selRecFilter(id uint32, country uint16, city uint16) IndData {
	if country == 0 && city == 0 {
		return nil
	}
	acc := GetUser(uint32(id)) //acc := MainMap[id]
	intrst := acc.Interests
	ilen := 0                      // суммарная длина индексов по интересам
	for _, inter := range intrst { // считаем
		ind, err := intMap.GetAll(uint32(inter))
		if err != nil {
			return nil
		}
		ilen += ind.Len()
	}
	var index IndData // индекс страна или город
	if city != 0 {
		//fmt.Println("city")
		index, _ = cityMap.GetAll(uint32(city))
	} else {
		index, _ = countryMap.GetAll(uint32(country))
	}
	if index.Len() < ilen/10 {
		return index
	}
	return nil
}

/*altFilter - если брать индекс по городу или стране а по интересам фильтровать*/
func altFilter(id uint32, indat IndData, limit int, out []*User) []*User {
	out = out[:0]
	acc := GetUser(uint32(id))
	intrst := acc.Interests
	sex := acc.Sex
	rec := !sex
	//fmt.Println(len(indat.GetAll()))
next:
	for _, idUser := range indat.GetAll() {
		user := GetUser(uint32(idUser)) // пользователь
		if user.Sex == rec {            //нужный пол
			for _, interes := range intrst { //интересы
				for _, uinter := range user.Interests {
					if interes == uinter { // хотя бы один интерес совпадает
						out = append(out, user)
						continue next
					}
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		f := out[i]
		s := out[j]
		if f.IsPremium() != s.IsPremium() {
			return f.IsPremium()
		}
		if f.Status != s.Status {
			return f.Status < s.Status
		}
		commf := f.GetCommInt(*acc)
		comms := s.GetCommInt(*acc)
		if commf != comms {
			return commf > comms // у кого больше общих интересов
		}
		// по разнице в возрасте
		agef := math.Abs(float64(int64(acc.Birth) - int64(f.Birth)))
		ages := math.Abs(float64(int64(acc.Birth) - int64(s.Birth)))
		if agef != ages {
			return agef < ages
		}
		// по id
		return f.ID < s.ID

	})

	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

type recSortedEl struct {
	commonInt int
	dt        float64
	u         *User
}
