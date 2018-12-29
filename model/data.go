package model

import (
	"fmt"
	"sync"
)

var WrMutex = new(sync.Mutex)
var accounts []Account
var accMap = make(map[int]*Account)
var accMailMap = make(map[string]int)
var accPhone = make(map[string]int)

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
		accMap[id] = pacc
		accMailMap[pacc.Email] = pacc.ID
		if pacc.Phone != "" {
			accPhone[pacc.Phone] = pacc.ID
		}
	}
}

/*GetAccount - получение значения аккаунта*/
func GetAccount(id int) (Account, error) {
	acc, ok := accMap[id]
	if !ok {
		return Account{}, fmt.Errorf("Нет аккаунта %d", id)
	}
	return *acc, nil
}

/*AddAcc - добавление элемента*/
func AddAcc(account Account) {
	accounts = append(accounts, account)            // добавление в список
	accMap[account.ID] = &accounts[len(accounts)-1] // указатель на последний элемент в карту
	accMailMap[account.Email] = accounts[len(accounts)-1].ID
	if account.Phone != "" {
		accPhone[account.Phone] = accounts[len(accounts)-1].ID
	}
	return
}

/*GetAccMail - получение аккаунта по email*/
func GetAccMail(email string) int {
	id, ok := accMailMap[email]
	if !ok {
		return -1
	}
	return id
}

/*GetAccPhone -  аккаунт по телефону*/
func GetAccPhone(phone string) int {
	id, ok := accPhone[phone]
	if !ok {
		return -1
	}
	return id
}

/*GetAccountPointer - получение указателя на аккаунт*/
func GetAccountPointer(id int) (*Account, error) {
	accp, ok := accMap[id]
	if !ok {
		return &Account{}, fmt.Errorf("Нет аккаунта")
	}
	return accp, nil
}

/*UpdateEmail - обновить карту email*/
func UpdateEmail(n, old string) {
	id := accMailMap[old]
	delete(accMailMap, old)
	accMailMap[n] = id
}

/*UpdatePhone - обновить карту email*/
func UpdatePhone(n, old string) {
	id := accPhone[old]
	delete(accPhone, old)
	accPhone[n] = id
}

/*IsMailExist - email существует*/
func IsMailExist(mail string) bool {
	_, exist := accMailMap[mail]
	return exist
}
