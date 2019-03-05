package model

/* 12 корзин
0 - свободные женщины прем
1 - сложно женщины прем
2 - занятые женщ. прем.
3 - свободные женщины
4 - сложно женщины
5 - занятые женщ
6 - свободные мужчины прем
7 - сложно мужчины прем
8 - заняты мужчины прем
9 - свободные мужчины
10 - сложно мужчины
11 - заняты мужчины
*/

var recIndex = make(map[uint32][]IndSData)

/*AddRecIndex - добавление индекса*/
func AddRecIndex(user User) {
	num := rBNum(user)
	ints := user.Interests
	for _, inum := range ints {
		ni, ok := recIndex[uint32(inum)]
		if !ok { // делаем новую корзину с интересом
			data := make([]IndSData, 12)
			for i := 0; i < 12; i++ {
				data[i] = IndSData(make([]uint32, 0))
			}
			recIndex[uint32(inum)] = data
			ni = recIndex[uint32(inum)]
		}
		bucket := ni[num]
		b2 := bucket.Add(user.ID)
		bucket = b2.(IndSData)
		//fmt.Println(bucket.Add(user.ID).Len(), user.ID)
		ni[num] = bucket

	}
}

/*RemRecIndex - удаление индекса*/
func RemRecIndex(user User) {
	ints := user.Interests
	num := rBNum(user)
	for _, inum := range ints {
		ni := recIndex[uint32(inum)]
		bucket := ni[num]
		b2 := bucket.Remove(user.ID)
		bucket = b2.(IndSData)
		ni[num] = bucket
	}

}

/*rBNum - вычисление номера*/
func rBNum(user User) int {
	var num int
	if !user.IsPremium() {
		num += 3
	}
	if user.Sex {
		num += 6
	}
	num += int(user.Status)
	return num
}
