package model

import (
	"fmt"
	"sort"
	"sync"
)

var WrMutex = new(sync.Mutex)
var users = make([]User, 0)

/*GetAccounts - Получение списка*/
func GetAccounts() []User {
	return users
}

/*SetUsers - установка списка*/
func SetUsers() {
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	fmt.Println("Сортировано")
	// //users = make([]User, 0, len(acc)*2) //двойная емкость
	// users = acc

	for i := range users {
		id := users[i].ID
		pacc := &users[i]
		MainMap[id] = pacc
		MailMap[pacc.Email] = uint32(pacc.ID)
		if pacc.Phone != "" {
			PhoneMap[pacc.Phone] = uint32(pacc.ID)
		}
	}
	fmt.Println("Окончание упаковки")
	fmt.Println("Мужчин:", globalGr.bman+globalGr.fman+globalGr.cman)
	fmt.Println("Женщин:", globalGr.cwom+globalGr.bwom+globalGr.fwom)
	free := 0
	busy := 0
	complex := 0
	for _, v := range cityGMap {
		busy += int(v.bman) + int(v.bwom)
		free += int(v.fman) + int(v.fwom)
		complex += int(v.cman) + int(v.cwom)
	}
	fmt.Println("G Status:", globalGr.fman+globalGr.fwom, globalGr.cman+globalGr.cwom, globalGr.bman+globalGr.bwom)
	fmt.Println("Status:", free, complex, busy)
	num, ok := DataInter.Get("Путешествия")
	if !ok {
		fmt.Println("Путешествия не найдены")
	} else {
		fmt.Println(interGMap[uint32(num)])
	}
}

/*GetAccount - получение значения аккаунта*/
func GetAccount(id uint32) (User, error) {
	acc, ok := MainMap[id]
	if !ok {
		return User{}, fmt.Errorf("Нет аккаунта %d", id)
	}
	return *acc, nil
}

/*AddAcc - добавление элемента*/
func AddAcc(user User) {
	ln := len(users)
	if ln > 0 && user.ID < users[ln-1].ID { // если больше последнего элемента добавляем в конец иначе вставка
		i := ln - 1 // начиная с последнего элемента
		for i >= 0 && user.ID < users[ln-1].ID {
			i--
		}
		//вставка
		users = append(users, User{})
		copy(users[i+2:], users[i+1:]) // вставляем id перед i
		users[i+1] = user
		for j := range users[i:] { //изменяем указатели
			MainMap[users[j].ID] = &users[j]
		}
		fmt.Println("---Insert---")
	} else {
		users = append(users, user)
		MainMap[user.ID] = &users[len(users)-1] // указатель на последний элемент в карту
	} // добавление в список

	MailMap[user.Email] = uint32(users[len(users)-1].ID)
	if user.Phone != "" {
		PhoneMap[user.Phone] = uint32(users[len(users)-1].ID)
	}
	return
}

/*GetAccMail - получение аккаунта по email*/
func GetAccMail(email string) int {
	id, ok := MailMap[email]
	if !ok {
		return -1
	}
	return int(id)
}

/*GetAccPhone -  аккаунт по телефону*/
func GetAccPhone(phone string) int {
	id, ok := PhoneMap[phone]
	if !ok {
		return -1
	}
	return int(id)
}

/*GetAccountPointer - получение указателя на аккаунт*/
func GetAccountPointer(id uint32) (*User, error) {
	accp, ok := MainMap[id]
	if !ok {
		return &User{}, fmt.Errorf("Нет аккаунта")
	}
	return accp, nil
}

/*UpdateEmail - обновить карту email*/
func UpdateEmail(n, old string) {
	id := MailMap[old]
	delete(MailMap, old)
	MailMap[n] = id
}

/*UpdatePhone - обновить карту email*/
func UpdatePhone(n, old string) {
	id := PhoneMap[old]
	delete(PhoneMap, old)
	PhoneMap[n] = id
}

/*IsMailExist - email существует*/
func IsMailExist(mail string) bool {
	_, exist := MailMap[mail]
	return exist
}
