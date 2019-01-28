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
	//fmt.Println("Сортировано")
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
	fmt.Println("Start")

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
		for i >= 0 && user.ID < users[i].ID {
			i--
		}
		i = i + 1
		//вставка
		users = append(users, User{})
		copy(users[i+1:], users[i:]) // вставляем id перед i
		users[i] = user
		ts := users[i:]
		for j := range ts { //изменяем указатели
			MainMap[ts[j].ID] = &ts[j]
		}
		//fmt.Println("---Insert---")
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
