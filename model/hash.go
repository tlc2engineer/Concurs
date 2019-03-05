package model

import (
	"sync"
)

/*CountryMap - страны*/
var countryMap Index = NewIS() //NewIS()

/*CityMap - города*/
var cityMap Index = NewIS()

/*fnameMap - города*/
var fnameMap Index = NewIS()

/*MainMap - главная карта*/
var MainMap = make(map[uint32]*User)

/*MailMap - email map*/
var MailMap = make(map[string]uint32)

/*PhoneMap - phone map*/
var PhoneMap = make(map[string]uint32)

var DomainMap = make(map[uint16][]uint32)

/*intMap - интересы*/
var intMap Index = NewIS()

var likesM = make(map[uint32][]byte)

/*WhoMap -  кто лайкал*/
var whoMap = make(map[uint32]Id3B)

/*codeIndex - индекс по коде*/
var codeIndex = NewIS()

/*fnameIndex - индекс по фамилии*/
var snameIndex = NewIS()

/*bYearIndex - индекс по году рождения*/
var bYearIndex = NewIS()

/*jYearIndex - индекс по году регистрации*/
var jYearIndex = NewIS()

var domIndex = NewIS()

/*имена по полу 0 - женский 1- мужской 2 - и то и то*/
var snmu = &sync.Mutex{}
var sexNames = make(map[uint16]byte)

func addSexName(name uint16, sex bool) {
	snmu.Lock()
	defer snmu.Unlock()
	//------Имена мужские и женские--------------
	if name != 0 {
		nsex, ok := sexNames[name]
		if ok { // имя уже есть в карте
			if (sex && nsex == 0) || (!sex && nsex == 1) { // а пол другой
				nsex = 2 // и тот и то
			}
		} else {
			if sex {
				sexNames[name] = 1
			} else {
				sexNames[name] = 0
			}
		}
	}
}

//---------Общие интересы-------------
// матрица общих интересов
var commonInt = make([][]bool, 100)
var ciMut = &sync.Mutex{}

func setCommonInt(interests []uint16) {
	ciMut.Lock()
	defer ciMut.Unlock()
	for _, i1 := range interests {
		for _, i2 := range interests {
			if i1 != i2 {
				commonInt[i1][i2] = true
				commonInt[i2][i1] = true
			}

		}
	}
}

/*exceptInterests - взаимоисключающие интересы(пиво футбол)*/
func exceptInterests(interests []uint16) bool {

	for _, i1 := range interests {
		for _, i2 := range interests {
			if i1 != i2 {
				if !commonInt[i1][i2] {
					//fmt.Println(i1, i2)
					return true
				}

			}

		}
	}
	return false
}
