package model

import (
	"fmt"
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
	// sort.Slice(acc, func(i, j int) bool {
	// 	return acc[i].ID > acc[j].ID
	// })
	// fmt.Println("Sorted")
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
	users = append(users, user)             // добавление в список
	MainMap[user.ID] = &users[len(users)-1] // указатель на последний элемент в карту
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
