package model

import (
	"Concurs/rgbtree"
	"fmt"
)

/*gInd  - количество свободных занятых и сложных в каком-то месте*/
type gInd struct {
	fman, cman, bman, fwom, cwom, bwom uint32 // свободные, сложные и занятые мужчины и женщины
}

func (gin *gInd) dec(sex bool, status int) {
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	switch num {
	case 0:
		gin.fwom = gin.fwom - 1
	case 1:
		gin.cwom = gin.cwom - 1
	case 2:
		gin.bwom = gin.bwom - 1
	case 3:
		gin.fman = gin.fman - 1
	case 4:
		gin.cman = gin.cman - 1
	case 5:
		gin.bman = gin.bman - 1
	}
}

func (gin *gInd) inc(sex bool, status int) {
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	switch num {
	case 0:
		gin.fwom = gin.fwom + 1
	case 1:
		gin.cwom = gin.cwom + 1
	case 2:
		gin.bwom = gin.bwom + 1
	case 3:
		gin.fman = gin.fman + 1
	case 4:
		gin.cman = gin.cman + 1
	case 5:
		gin.bman = gin.bman + 1
	}
}

var cityGMap = make(map[uint32]gInd)
var countryGMap = make(map[uint32]gInd)
var interGMap = make(map[uint32]gInd)
var cityInter = make(map[uint32]gInd)
var countryInter = make(map[uint32]gInd)
var globalGr = gInd{}

/*AddGIndex - добавление пользователя в индексы*/
func AddGIndex(user User) {
	city := user.City
	country := user.Country
	ints := user.Interests
	sex := user.Sex
	status := user.Status
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	cInd, ok := cityGMap[uint32(city)]
	if !ok {
		cityGMap[uint32(city)] = gInd{}
		cInd = cityGMap[uint32(city)]
	}
	pcInd := &cInd
	pcInd.inc(sex, int(status))
	coInd, ok := countryGMap[uint32(country)]
	if !ok {
		countryGMap[uint32(country)] = gInd{}
		coInd = countryGMap[uint32(country)]
	}
	pcoInd := &coInd
	pcoInd.inc(sex, int(status))
	pgl := &globalGr
	pgl.inc(sex, int(status))
	cityGMap[uint32(city)] = cInd
	countryGMap[uint32(country)] = coInd
	for _, inter := range ints {
		iind, ok := interGMap[uint32(inter)]
		if !ok {
			interGMap[uint32(inter)] = gInd{}
			iind = interGMap[uint32(inter)]
		}
		pind := &iind
		pind.inc(sex, int(status))
		interGMap[uint32(inter)] = iind
		//----City Inter--------------
		k := uint32(inter)<<16 | uint32(city)
		ciind, ok := cityInter[k]
		if !ok {
			cityInter[k] = gInd{}
			ciind = cityInter[k]
		}
		pcind := &ciind
		pcind.inc(sex, int(status))
		cityInter[k] = ciind
		//----Country Inter--------------
		k = uint32(inter)<<16 | uint32(country)
		ciind, ok = countryInter[k]
		if !ok {
			countryInter[k] = gInd{}
			ciind = countryInter[k]
		}
		pcind = &ciind
		pcind.inc(sex, int(status))
		countryInter[k] = ciind
	}
	//---------Индекс по году рождения-------------
	yi, ok := birthGI[user.getBYear()]
	if !ok {
		birthGI[user.getBYear()] = NewYGI()
		yi = birthGI[user.getBYear()]
	}
	yi.AddGIndex(user)
	birthGI[user.getBYear()] = yi
	//---------Индекс по дате присоединения-------------
	yi, ok = joinGI[user.getJYear()]
	if !ok {
		joinGI[user.getBYear()] = NewYGI()
		yi = joinGI[user.getBYear()]
	}
	yi.AddGIndex(user)
	joinGI[user.getJYear()] = yi

}

/*DeleteGIndex - удаление аккаунта из группового индекса*/
func DeleteGIndex(user User) {
	city := user.City
	country := user.Country
	ints := user.Interests
	sex := user.Sex
	status := user.Status
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	cInd := cityGMap[uint32(city)]
	pcInd := &cInd
	pcInd.dec(sex, int(status))
	cityGMap[uint32(city)] = cInd
	coInd := countryGMap[uint32(country)]
	pcoInd := &coInd
	pcoInd.dec(sex, int(status))
	countryGMap[uint32(country)] = coInd
	pgl := &globalGr
	pgl.dec(sex, int(status))
	for _, inter := range ints {
		iind := interGMap[uint32(inter)]
		pind := &iind
		pind.dec(sex, int(status))
		interGMap[uint32(inter)] = iind
		//----------------------------
		k := uint32(inter)<<16 | uint32(city)
		ciind, _ := cityInter[k]
		pcind := &ciind
		pcind.dec(sex, int(status))
		cityInter[k] = ciind
		//----------------------------
		k = uint32(inter)<<16 | uint32(country)
		ciind, _ = countryInter[k]
		pcind = &ciind
		pcind.dec(sex, int(status))
		countryInter[k] = ciind
	}
	//-------------------------------------------------
	yi := birthGI[user.getBYear()]
	yi.DeleteGIndex(user)
	birthGI[user.getBYear()] = yi
	//---------Индекс по дате присоединения-------------
	yi = joinGI[user.getJYear()]
	yi.DeleteGIndex(user)
	joinGI[user.getJYear()] = yi
	//--------------------------------------------------
}

/*GroupI - группировка по ключам город пол статус*/
func GroupI(keys []string, sex int, status int, m *rgbtree.UTree, country uint16, city uint16, interests []uint16) bool {
	var tm map[uint32]gInd
	tm = selMap(keys, country, city, interests)
	if tm == nil {
		return false
	}
	var buff = make([]byte, 0, 10)
	for k, v := range tm {
		k1 := uint16(k)
		var k2 uint16
		if k < 65536 {
			k2 = k1
		} else {
			k2 = uint16(k >> 16)
		}
		//-----фильтр по интересу
		if len(interests) != 0 {
			find := false
			for _, inter := range interests {
				if inter == k2 {
					find = true
					break
				}
			}
			if !find {
				continue
			}
		}
		//-----фильтр по городу---------
		if city != 0 && city != k1 {
			continue
		}
		//-----фильтр по стране--------
		if country != 0 && country != k1 {
			continue
		}
		//-----------------------------
		for i := 0; i < 6; i++ {
			if (sex == 0 && i > 2) || (sex == 1 && i < 3) { // фильтр по полу
				continue
			}
			if status != -1 && status != i%3 { // фильтр по статусу
				continue
			}
			buff = buff[:0]
			for _, key := range keys {
				switch key {
				case "country":
					country := k1
					b0 := byte(country)
					b1 := byte(country >> 8)
					buff = append(buff, b0, b1)
				case "interests":
					inter := k2
					b0 := byte(inter)
					b1 := byte(inter >> 8)
					buff = append(buff, b0, b1)
				case "city":
					city := k1
					b0 := byte(city)
					b1 := byte(city >> 8)
					buff = append(buff, b0, b1)
				case "status":
					buff = append(buff, byte(i%3))
				case "sex":
					if i < 3 {
						buff = append(buff, 0)
					} else {
						buff = append(buff, 1)
					}
				}
			}
			var out uint64
			// запаковка
			for j := 0; j < len(buff); j++ {
				out |= uint64(buff[j]) << (uint16(j) * 8)
			}
			switch i {
			case 0:
				if v.fwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.fwom))
				}
			case 1:
				if v.cwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.cwom))
				}
			case 2:
				if v.bwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.bwom))
				}
			case 3:
				if v.fman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.fman))
				}
			case 4:
				if v.cman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.cman))
				}
			case 5:
				if v.bman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.bman))
				}
			}
		}

	}
	return true
}

/*selMap - Выбор карты*/
func selMap(keys []string, country uint16, city uint16, interests []uint16) map[uint32]gInd {
	//---------------------------------------------------------
	if city != 0 && country != 0 {
		return nil
	}
	if containsAll(keys, []string{"city", "country"}) {
		fmt.Println(keys)
		return nil
	}
	//---------------------------------------------------------
	if city == 0 && country == 0 && len(interests) == 0 {
		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{0: globalGr}
		}
		if contains(keys, []string{"sex", "status", "city"}) {
			return cityGMap
		}
		if contains(keys, []string{"sex", "status", "country"}) {
			return countryGMap
		}
		if contains(keys, []string{"sex", "status", "interests"}) {
			return interGMap
		}
		if contains(keys, []string{"sex", "status", "interests", "city"}) {
			return cityInter
		}
		if contains(keys, []string{"sex", "status", "interests", "country"}) {
			return countryInter
		}
	}
	if country != 0 && city == 0 && len(interests) == 0 {
		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{uint32(country): countryGMap[uint32(country)]}
		}
	}
	if city != 0 && country == 0 && len(interests) == 0 {

		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{uint32(city): cityGMap[uint32(city)]}
		}
	}
	if len(interests) != 0 && country == 0 && city == 0 {
		if contains(keys, []string{"sex", "status"}) {
			out := map[uint32]gInd{}
			for _, inter := range interests {
				out[uint32(inter)] = interGMap[uint32(inter)]

			}
			return out
		}
	}
	if country == 0 && contains(keys, []string{"sex", "status", "city", "interests"}) {
		return cityInter
	}
	if city == 0 && contains(keys, []string{"sex", "status", "country", "interests"}) {
		return countryInter
	}
	return nil
}

func contains(keys []string, ref []string) bool {

	for i := 0; i < len(keys); i++ {
		find := false
		for j := 0; j < len(ref); j++ {
			if ref[j] == keys[i] {
				find = true
				break
			}
		}
		if !find {
			return false
		}
	}
	return true
}

func containsAll(keys []string, ref []string) bool {

	for i := 0; i < len(ref); i++ {
		find := false
		for j := 0; j < len(keys); j++ {
			if ref[i] == keys[j] {
				find = true
				break
			}
		}
		if !find {
			return false
		}
	}
	return true
}

//---------------------------------------------------------------------
//---------------------------------------------------------------------
var birthGI = make(map[int]*yearGIndex) // рождения
var joinGI = make(map[int]*yearGIndex)  // joined

/*yearGIndex - групповой индекс за год*/
type yearGIndex struct {
	cityMap    map[uint32]gInd
	countryMap map[uint32]gInd
	interMap   map[uint32]gInd
	globalGr   gInd
}

/*NewYGI - новый годовой индекс*/
func NewYGI() *yearGIndex {
	return &yearGIndex{cityMap: make(map[uint32]gInd), countryMap: make(map[uint32]gInd), interMap: make(map[uint32]gInd), globalGr: gInd{}}
}

/*AddGIndex - добавление пользователя в индекс*/
func (yi *yearGIndex) AddGIndex(user User) {
	city := user.City
	country := user.Country
	ints := user.Interests
	sex := user.Sex
	status := user.Status
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	cInd, ok := yi.cityMap[uint32(city)]
	if !ok {
		yi.cityMap[uint32(city)] = gInd{}
		cInd = yi.cityMap[uint32(city)]
	}
	pcInd := &cInd
	pcInd.inc(sex, int(status))
	coInd, ok := yi.countryMap[uint32(country)]
	if !ok {
		yi.countryMap[uint32(country)] = gInd{}
		coInd = yi.countryMap[uint32(country)]
	}
	pcoInd := &coInd
	pcoInd.inc(sex, int(status))
	pgl := &yi.globalGr // глобальный
	pgl.inc(sex, int(status))
	yi.cityMap[uint32(city)] = cInd
	yi.countryMap[uint32(country)] = coInd
	//-------Интересы----------------
	for _, inter := range ints {
		iind, ok := yi.interMap[uint32(inter)]
		if !ok {
			yi.interMap[uint32(inter)] = gInd{}
			iind = yi.interMap[uint32(inter)]
		}
		pind := &iind
		pind.inc(sex, int(status))
		yi.interMap[uint32(inter)] = iind
	}

}

/*DeleteGIndex - удаление аккаунта из группового индекса*/
func (yi *yearGIndex) DeleteGIndex(user User) {
	city := user.City
	country := user.Country
	ints := user.Interests
	sex := user.Sex
	status := user.Status
	var num int
	if sex {
		num = 3
	}
	num += int(status)
	//-----------------------------
	cInd := yi.cityMap[uint32(city)]
	pcInd := &cInd
	pcInd.dec(sex, int(status))
	yi.cityMap[uint32(city)] = cInd
	//----------------------------
	coInd := yi.countryMap[uint32(country)]
	pcoInd := &coInd
	pcoInd.dec(sex, int(status))
	yi.countryMap[uint32(country)] = coInd
	//-----------------------------
	pgl := &yi.globalGr
	pgl.dec(sex, int(status))
	//----------------------------
	for _, inter := range ints {
		iind := yi.interMap[uint32(inter)]
		pind := &iind
		pind.dec(sex, int(status))
		yi.interMap[uint32(inter)] = iind
	}
}

/*GroupI - группировка по ключам город пол статус*/
func (yi *yearGIndex) GroupI(keys []string, sex int, status int, m *rgbtree.UTree, country uint16, city uint16, interests []uint16) bool {
	var tm map[uint32]gInd
	tm = yi.selMap(keys, country, city, interests)
	if tm == nil {
		return false
	}
	//fmt.Println(tm)
	for k, v := range tm {
		k1 := uint16(k)
		var k2 uint16
		if k < 65536 {
			k2 = k1
		} else {
			k2 = uint16(k >> 16)
		}
		//-----фильтр по интересу
		if len(interests) != 0 {
			find := false
			for _, inter := range interests {
				if inter == k2 {
					find = true
					break
				}
			}
			if !find {
				continue
			}
		}
		//-----фильтр по городу---------
		if city != 0 && city != k1 {
			continue
		}
		//-----фильтр по стране--------
		if country != 0 && country != k1 {
			continue
		}
		//-----------------------------
		for i := 0; i < 6; i++ {
			if (sex == 0 && i > 2) || (sex == 1 && i < 3) { // фильтр по полу
				continue
			}
			if status != -1 && status != i%3 { // фильтр по статусу
				continue
			}
			var buff = make([]byte, 0)
			for _, key := range keys {
				switch key {
				case "country":
					country := k1
					b0 := byte(country)
					b1 := byte(country >> 8)
					buff = append(buff, b0, b1)
				case "interests":
					inter := k2
					b0 := byte(inter)
					b1 := byte(inter >> 8)
					buff = append(buff, b0, b1)
				case "city":
					city := k1
					b0 := byte(city)
					b1 := byte(city >> 8)
					buff = append(buff, b0, b1)
				case "status":
					buff = append(buff, byte(i%3))
				case "sex":
					if i < 3 {
						buff = append(buff, 0)
					} else {
						buff = append(buff, 1)
					}
				}
			}
			var out uint64
			// запаковка
			for j := 0; j < len(buff); j++ {
				out |= uint64(buff[j]) << (uint16(j) * 8)
			}
			//fmt.Println("out", out, v)
			switch i {
			case 0:
				if v.fwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.fwom))
				}
			case 1:
				if v.cwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.cwom))
				}
			case 2:
				if v.bwom > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.bwom))
				}
			case 3:
				if v.fman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.fman))
				}
			case 4:
				if v.cman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.cman))
				}
			case 5:
				if v.bman > 0 {
					val, _ := m.Get(out)
					m.Put(out, val+int(v.bman))
				}
			}
		}

	}
	//fmt.Println(m.Get(0))
	return true
}

/*selMap - Выбор карты для года*/
func (yi *yearGIndex) selMap(keys []string, country uint16, city uint16, interests []uint16) map[uint32]gInd {
	//---------------------------------------------------------
	if city != 0 && country != 0 {
		return nil
	}
	if containsAll(keys, []string{"city", "country"}) {
		//fmt.Println(keys)
		return nil
	}
	//---------------------------------------------------------
	if city == 0 && country == 0 && len(interests) == 0 {
		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{0: yi.globalGr}
		}
		if contains(keys, []string{"sex", "status", "city"}) {
			return yi.cityMap
		}
		if contains(keys, []string{"sex", "status", "country"}) {
			return yi.countryMap
		}
		if contains(keys, []string{"sex", "status", "interests"}) {
			return yi.interMap
		}
	}
	if country != 0 && city == 0 && len(interests) == 0 {
		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{uint32(country): yi.countryMap[uint32(country)]}
		}
	}
	if city != 0 && country == 0 && len(interests) == 0 {
		if contains(keys, []string{"sex", "status"}) {
			return map[uint32]gInd{uint32(city): yi.cityMap[uint32(city)]}
		}
	}
	if len(interests) != 0 && country == 0 && city == 0 {
		if contains(keys, []string{"sex", "status"}) {
			out := map[uint32]gInd{}
			for _, inter := range interests {
				out[uint32(inter)] = yi.interMap[uint32(inter)]

			}
			return out
		}
	}
	return nil
}

/*GBirthY - индекс по году рождения*/
func GBirthY(keys []string, sex int, status int, m *rgbtree.UTree, country uint16, city uint16, interests []uint16, year int) bool {
	yi, ok := birthGI[year]
	if !ok {
		return false
	}
	return yi.GroupI(keys, sex, status, m, country, city, interests)
}

/*GJoinY - индекс по году присоединения*/
func GJoinY(keys []string, sex int, status int, m *rgbtree.UTree, country uint16, city uint16, interests []uint16, year int) bool {
	yi, ok := joinGI[year]
	if !ok {
		return false
	}
	return yi.GroupI(keys, sex, status, m, country, city, interests)
}
