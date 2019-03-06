package model

import (
	"Concurs/rgbtree"

	//"Concurs/model"

	"sort"
	"strconv"
	"strings"
)

/*GroupAgg - индексы по запросу group*/
func GroupAgg(mess []Mess, m *rgbtree.UTree, ff []func(User) bool, kfunc func(user User) []uint64) bool {
	indexes := make([]indexLogic, 0)
	// fmt.Println(mess)
	for _, v := range mess {
		par := v.Par
		switch par {
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
					return true
				}
				data := ids.Get()
				sort.Slice(data, func(i, j int) bool {
					return data[i] < data[j]
				})
				//indexD := IndSData(data)
				indexes = append(indexes, newSLogic(data))
			}
		case "city":
			city := v.Val
			num, ok := DataCity.Get(city)
			if !ok {
				return true
			}
			all, err := cityMap.GetAll(uint32(num))
			if err != nil {
				//fmt.Println("Нет такого города ", num)
				return true
			}
			indexes = append(indexes, newSLogic(all.GetAll()))
		case "country":
			country := v.Val
			num, ok := DataCountry.Get(country)
			if !ok {
				return true // пустой массив
			}
			all, err := countryMap.GetAll(uint32(num))
			if err != nil {
				//fmt.Println("Нет такой страны ", num)
				return true // пустой массив
			}
			indexes = append(indexes, newSLogic(all.GetAll()))
		case "interests":
			par := v.Val
			if par == "" {
				return true // пустой массив
			}
			pari := strings.Split(par, ",")
			dat := make([]uint16, len(pari))
			for i := range pari {
				dat[i], _ = DataInter.Get(pari[i])
			}
			logics := make([]*simpleIndexLogic, 0)
			for _, inter := range dat {
				idata, err := intMap.GetAll(uint32(inter))
				if err != nil {
					continue
				}
				sl := newSLogic(idata.GetAll())
				logics = append(logics, sl)
			}
			indexes = append(indexes, newCmplLog(logics))
		case "birth":
			sy := v.Val
			year, err := strconv.ParseInt(sy, 10, 0)
			//fmt.Println(year)
			if err != nil {
				return true // пустой массивs
			}
			ind, err := bYearIndex.GetAll(uint32(year))
			//fmt.Println(year, ind.Len())
			if err != nil {
				return true // пустой массивs
			}
			//fmt.Println(year, len(ind.GetAll()))
			indexes = append(indexes, newSLogic(ind.GetAll()))
		case "joined":
			sy := v.Val
			year, err := strconv.ParseInt(sy, 10, 0)
			if err != nil {
				return true // пустой массивs
			}
			ind, err := jYearIndex.GetAll(uint32(year))
			if err != nil {
				return true // пустой массивs
			}
			indexes = append(indexes, newSLogic(ind.GetAll()))
		}
	}
	//fmt.Println(len(indexes), indexes[0].len(), indexes[1].len())
	if len(indexes) == 0 {
		return false
	}
	if len(indexes) == 1 { // один индекс
		base := indexes[0]
	m2:
		for {
			i := base.nextIt()
			if i == -1 {
				break
			}
			acc := MainMap[uint32(i)]
			for _, fn := range ff {
				if !fn(*acc) {
					continue m2
				}
			}
			newSres := kfunc(*acc)
			for _, r := range newSres {
				count, ok := m.Get(r)
				if ok {
					count++
					m.Put(r, count)
					//m[r] = count
				} else {
					m.Put(r, 1)
				}
			}
		}
		return true
	}
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
		for _, ind := range indexes {
			if !ind.nextCheck(uint32(val)) {
				continue m1
			}
		}
		acc := *MainMap[uint32(val)]
		for _, fn := range ff {
			if !fn(acc) {
				continue m1
			}
		}
		newSres := kfunc(acc)
		for _, r := range newSres {
			count, ok := m.Get(r)
			if ok {
				count++
				m.Put(r, count)
			} else {
				m.Put(r, 1)
			}
		}

	}
	return true
}
