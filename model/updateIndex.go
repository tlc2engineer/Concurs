package model

/*UpdICountry - обновление индекса по стране*/
func UpdICountry(id uint32, oldCountry uint16, newCountry string) {
	countryMap.Remove(uint32(oldCountry), id)                    //номер старой страны убираем
	countryMap.Add(uint32(DataCountry.GetOrAdd(newCountry)), id) // номер новой страны добавляем
}

/*UpdICity - обновление индекса по городу*/
func UpdICity(id uint32, oldCity uint16, newCity string) {
	cityMap.Remove(uint32(oldCity), id)                 //номер старой страны убираем
	cityMap.Add(uint32(DataCity.GetOrAdd(newCity)), id) // номер новой страны добавляем
}

/*UpdInter - обновление индекса по интересам*/
func UpdInter(id uint32, oldInterests []uint16, newInterests []string) {
	for _, item := range oldInterests { // удаление старых индексов
		intMap.Remove(uint32(item), id)
	}
	for _, item := range newInterests { // добавление новых индексов
		intMap.Add(uint32(DataInter.GetOrAdd(item)), id)
	}
}

/*UpdFname - обновление имени*/
func UpdFname(id uint32, oldName uint16, newName string) {
	fnameMap.Remove(uint32(oldName), id)
	fnameMap.Add(uint32(DataFname.GetOrAdd(newName)), id)
}

/*UpdSname - обновление фамилии*/
func UpdSname(id uint32, oldName uint32, newName string) {
	snameIndex.Remove(uint32(oldName), id)
	snameIndex.Add(uint32(DataSname[newName]), id)
}

/*UpdCode - обновление кода*/
func UpdCode(id uint32, oldCode uint16, newCode uint16) {
	codeIndex.Remove(uint32(oldCode), id)
	codeIndex.Add(uint32(newCode), id)
}

/*UpdateBYear - обновление индекса по году рождения*/
func UpdateBYear(id uint32, oldYear uint32, newYear uint32) {
	bYearIndex.Remove(oldYear, id)
	bYearIndex.Add(newYear, id)
}

/*UpdateJYear - Обновлеие индекса по году регистрации*/
func UpdateJYear(id uint32, oldYear uint32, newYear uint32) {
	jYearIndex.Remove(oldYear, id)
	jYearIndex.Add(newYear, id)
}

/*UpdateDomainInd - обновление индекса по домену*/
func UpdateDomainInd(id uint32, OldEmail, newEmail string) {
	oldD := getDomain(OldEmail)
	newD := getDomain(newEmail)
	domIndex.Remove(uint32(oldD), id)
	domIndex.Add(uint32(newD), id)
}
