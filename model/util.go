package model

import (
	"sync"
)

var likeBuff = sync.Pool{
	New: func() interface{} {
		return make([]Like, 0, 100)
	},
}

/*Intercept - пересечение двух срезов на основе наименьшего*/
func Intercept(arr1 []uint32, arr2 []uint32) []uint32 {
	out := make([]uint32, 0)
	var first = arr1
	var sec = arr2
	if len(arr2) < len(arr1) { //выбираем наименьший массив за исходный
		first = arr2
		sec = arr1
	}
	for i := 0; i < len(first); i++ {
		for j := 0; j < len(sec); j++ {
			if first[i] == sec[j] {
				out = append(out, first[i])
			}
		}
	}
	return out
}

/*union - объединение двух срезов с уникальностью элементов*/
func union(arr1 []uint32, arr2 []uint32) []uint32 {
	m := make(map[uint32]bool) // карта уникальных элементов
	for _, item := range arr1 {
		_, ok := m[item] // если нет такого id добавляем
		if !ok {
			m[item] = false
		}
	}
	for _, item := range arr2 {
		_, ok := m[item] // если нет такого id добавляем
		if !ok {
			m[item] = false
		}
	}
	out := make([]uint32, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

/*uniq - функция уникальности*/
func uniq(data []uint32) []uint32 {
	m := make(map[uint32]bool)
	//ex := make([]int, 0)
	for i := 0; i < len(data); i++ {
		_, ok := m[data[i]] // есть ли такой элемент
		if !ok {            // если есть добавляем в исключения
			m[data[i]] = false
		}
	}
	cnt := 0
	for i := range m {
		data[cnt] = i
		cnt++
	}
	return data[:cnt]
}

/*InterceptSorted - пересечение двух отсортированных slice uint32*/
func InterceptSorted(s1, s2 []uint32) []uint32 {
	mlen := len(s1)
	init := s1
	second := s2
	if len(s2) < mlen {
		mlen = len(s2)
		init = s2
		second = s1
	}
	out := make([]uint32, 0, mlen)
	var m int
	for i := 0; i < len(init); i++ {
		for j := m; j < len(second); j++ {
			if init[i] < second[j] { // элемент нет такого элемента
				m = j
				break
			}
			if init[i] == second[j] { // элемент найден
				out = append(out, init[i])
				m = j + 1
				break
			}
		}
	}
	return out
}

/*Union2Sorted - объединение двух отсортированных slice uint32*/
func Union2Sorted(arr1, arr2 []uint32) []uint32 {
	ln1 := len(arr1)
	ln2 := len(arr2)
	if ln1 == 0 {
		return arr2
	}
	if ln2 == 0 {
		return arr1
	}
	out := make([]uint32, 0, ln1+ln2)
	i := 0
	j := 0
	for i < ln1 && j < ln2 {
		if arr1[i] == arr2[j] {
			out = append(out, arr1[i])
			j++
			i++
		} else if arr1[i] > arr2[j] {
			out = append(out, arr2[j])
			j++
		} else {
			out = append(out, arr1[i])
			i++
		}
		if i == ln1 && j < ln2 {
			out = append(out, arr2[j:]...)
			break
		}
		if j == ln2 && i < ln1 {
			out = append(out, arr1[i:]...)
			break
		}
	}
	return out
}

/*UnionRSorted - объединение двух отсортированных slice uint32 данные из ресурсов*/
func UnionRSorted(out, arr1, arr2 []uint32) []uint32 {
	ln1 := len(arr1)
	ln2 := len(arr2)
	if ln1 == 0 {
		return arr2
	}
	if ln2 == 0 {
		return arr1
	}
	if cap(out) < ln1+ln2 {
		out = make([]uint32, ln1+ln2)
	}
	out = out[:0]
	i := 0
	j := 0
	for i < ln1 && j < ln2 {
		if arr1[i] == arr2[j] {
			out = append(out, arr1[i])
			j++
			i++
		} else if arr1[i] > arr2[j] {
			out = append(out, arr2[j])
			j++
		} else {
			out = append(out, arr1[i])
			i++
		}
		if i == ln1 && j < ln2 {
			out = append(out, arr2[j:]...)
			break
		}
		if j == ln2 && i < ln1 {
			out = append(out, arr1[i:]...)
			break
		}
	}
	return out
}

/*InterceptRSorted - пересечение двух отсортированных slice uint32, данные из ресурса*/
func InterceptRSorted(out, s1, s2 []uint32) []uint32 {
	mlen := len(s1)
	init := s1
	second := s2
	if len(s2) < mlen {
		mlen = len(s2)
		init = s2
		second = s1
	}
	if cap(out) < mlen {
		out = make([]uint32, mlen)
	}
	out = out[:0]
	var m int
	for i := 0; i < len(init); i++ {
		for j := m; j < len(second); j++ {
			if init[i] < second[j] { // элемент нет такого элемента
				m = j
				break
			}
			if init[i] == second[j] { // элемент найден
				out = append(out, init[i])
				m = j + 1
				break
			}
		}
	}
	return out
}

/*searchInSorted - поиск id в сортированном срезе. Возвращает номер в срезе. Если не находит -1.*/
func searchInSorted(data []uint32, id uint32) int {
	ln := len(data)
	if ln == 1 {
		return 0
	}
	first := data[0]             // первый
	last := data[ln-1]           // последний
	if id > last || id < first { // не входит в пределы
		panic("Error 1") //return -1
	}
	proc := float64(id-first) / float64(last-first)
	init := int(float64(ln) * proc) // начало поиска
	if init > ln-1 {
		init = ln - 1
	}
	i := init
	var tooMany = false   // слишком много
	var tooLittle = false // слишком мало
	for {
		if (tooLittle && tooMany) || i < 0 || i > len(data)-1 {
			panic("Error 2")
			//return -1
		}
		if data[i] < id {
			tooLittle = true
			i++
		} else if data[i] > id {
			tooMany = true
			i--
		} else {
			return i
		}
	}
	// for i := range data {
	// 	if id == data[i] {
	// 		return i
	// 	}
	// }
	//return -1
}

/*searchInsPlace - поиск места для вставки*/
func searchInsPlace(data []uint32, id uint32) int {
	ln := len(data) // длина данных
	if ln == 0 {    // если пусто вставка в конец списка
		return -1
	}
	first := data[0]   // первый
	last := data[ln-1] // последний
	if id < first {    // вставка в начало списка
		return 0
	}
	if id > last { // вставка в конец списка
		return -1
	}
	proc := float64(id-first) / float64(last-first)
	init := int(float64(ln) * proc) // начало поиска
	if init > ln-1 {
		init = ln - 1
	}
	i := init
	for {
		if i != 0 && (data[i] > id && data[i-1] < id) { // id меньше этого но больше предыдущего
			return i
		} else if data[i] < id {
			i++
		} else if data[i] > id {
			i--
		} else {
			return -2 // нашли равный элемент
		}
	}

}
