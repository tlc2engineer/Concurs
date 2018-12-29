package model

/*GetID - Получить ID*/
func GetID(acc Account) int {
	return acc.ID
}

/*GetSParam - получаем строковой параметр*/
func (acc Account) GetSParam(name string) string {
	switch name {
	case "email":
		return acc.Email
	case "fname":
		return acc.FName
	case "sname":
		return acc.SName
	case "phone":
		return acc.Phone
	case "sex":
		return acc.Sex
	case "country":
		return acc.Country
	case "city":
		return acc.City
	case "status":
		return acc.Status
	}
	return "error"
}
