package model

import (
	"fmt"
	"sync"
)

var WrMutex = new(sync.Mutex)
var accounts []Account

/*GetAccounts - Получение списка*/
func GetAccounts() []Account {
	return accounts
}

/*SetAccounts - установка списка*/
func SetAccounts(acc []Account) {
	accounts = make([]Account, 0, len(acc)*2) //двойная емкость
	accounts = append(accounts, acc...)
	for i := range accounts {
		id := accounts[i].ID
		pacc := &accounts[i]
		MainMap[id] = pacc
		MailMap[pacc.Email] = uint32(pacc.ID)
		if pacc.Phone != "" {
			PhoneMap[pacc.Phone] = uint32(pacc.ID)
		}
	}
}

/*GetAccount - получение значения аккаунта*/
func GetAccount(id int) (Account, error) {
	acc, ok := MainMap[id]
	if !ok {
		return Account{}, fmt.Errorf("Нет аккаунта %d", id)
	}
	return *acc, nil
}

/*AddAcc - добавление элемента*/
func AddAcc(account Account) {
	accounts = append(accounts, account)             // добавление в список
	MainMap[account.ID] = &accounts[len(accounts)-1] // указатель на последний элемент в карту
	MailMap[account.Email] = uint32(accounts[len(accounts)-1].ID)
	if account.Phone != "" {
		PhoneMap[account.Phone] = uint32(accounts[len(accounts)-1].ID)
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
func GetAccountPointer(id int) (*Account, error) {
	accp, ok := MainMap[id]
	if !ok {
		return &Account{}, fmt.Errorf("Нет аккаунта")
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
