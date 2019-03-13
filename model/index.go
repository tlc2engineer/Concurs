package model

import (
	"fmt"
	"sort"
	"sync"
)

/*IndData - Данные индекса*/
type IndData interface {
	Len() int
	Remove(id uint32) IndData
	Add(id uint32) IndData
	IsExists(id uint32) bool
	GetAll() []uint32
	Intercept(idata IndData) IndData
	Union(idata IndData) IndData
}

/*IndSData - Данные индекса в списке*/
type IndSData []uint32

/*Len - длина*/
func (data IndSData) Len() int {
	return len(data)
}

/*Remove - удаление id из списка*/
func (data IndSData) Remove(id uint32) IndData {
	n := searchInSorted(data, id)
	if n == -1 {
		return data
	}
	// for i := range data {
	// 	if id == data[i] {
	// 		data = data[:i+copy(data[i:], data[i+1:])]
	// 		break
	// 	}
	// }
	data = data[:n+copy(data[n:], data[n+1:])]
	return data
}

/*Add - добавление id в список*/
func (data IndSData) Add(id uint32) IndData {
	ln := len(data)

	if ln > 0 && id > data[ln-1] { // добавляем в конец
		data = append(data, id)
	} else { // вставка
		n := searchInsPlace(data, id)
		switch n {
		case -2: // элемент найден
			return data
		case -1: // вставка в конец
			data = append(data, id)
		default: // вставка элемента на позицию n
			data = append(data, 0)
			copy(data[n+1:], data[n:]) // вставляем id перед i
			data[n] = id
		}
	}

	// if ln > 0 && id < data[ln-1] { // если больше последнего элемента добавляем в конец иначе вставка
	// 	i := ln - 1 // начиная с последнего элемента
	// 	for i >= 0 && id < data[i] {
	// 		i--
	// 	}
	// 	if i >= 0 && id == data[i] { // если тот же элемент
	// 		return data
	// 	}
	// 	//вставка
	// 	data = append(data, 0)
	// 	copy(data[i+2:], data[i+1:]) // вставляем id перед i
	// 	data[i+1] = id
	// } else {
	// 	data = append(data, id)
	// }
	return data
}

/*IsExists - id в списке*/
func (data IndSData) IsExists(id uint32) bool {
	for _, item := range data {
		if item == id {
			return true
		}
	}
	return false
}

/*GetAll - получение данных индекса в виде среза*/
func (data IndSData) GetAll() []uint32 {
	return data
}

/*Intercept - пересечение индексных данных*/
func (data IndSData) Intercept(idata IndData) IndData {
	out := make([]uint32, 0)
	mdata, ok := idata.(IndexMData) // если IndexMData
	if ok {
		for i := range data {
			if mdata.IsExists(data[i]) {
				out = append(out, data[i])
			}
		}
	} else { // пересечение двух отсортированных массивов
		out = InterceptSorted(data, idata.GetAll())
	}

	return IndSData(out)
}

/*Union - объединение индексных данных*/
func (data IndSData) Union(idata IndData) IndData {
	var out []uint32
	mdata, ok := idata.(IndexMData) // если IndexMData map
	if ok {
		out = make([]uint32, 0, len(data)+len(mdata))
		for i := range data {
			if !mdata.IsExists(data[i]) {
				out = append(out, data[i])
			}
		}
		for i := range mdata {
			out = append(out, i)
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i] < out[j]
		})

	} else {
		out = Union2Sorted(data, idata.GetAll())
	}

	return IndSData(out)
}

/*Index - интерфейс индекса*/
type Index interface {
	In(key uint32, id uint32) bool
	Has(key uint32) bool
	Add(key uint32, id uint32) bool
	Remove(key uint32, id uint32) bool
	GetAll(key uint32) (IndData, error)
}

/*IndexSlice - индекс на основе срезов*/
type IndexSlice map[uint32]IndData

/*NewIS - новый индекс*/
func NewIS() IndexSlice {
	return make(map[uint32]IndData)
}

/*Has - есть такой ключ*/
func (is IndexSlice) Has(key uint32) bool {
	_, ok := is[key]
	return ok
}

/*In - Есть ли такой id в заданном индексе*/
func (is IndexSlice) In(key uint32, id uint32) bool {
	data, ok := is[key]
	if !ok {
		return false
	}
	return data.IsExists(id)
}

/*Add - добавление нового значения*/
func (is IndexSlice) Add(key uint32, id uint32) bool {
	data, ok := is[key]
	if !ok {
		// idata := make([]uint32, 0, 10000)
		// idata = append(idata, id)
		is[key] = IndSData([]uint32{id})
		return false
	}
	is[key] = data.Add(id)
	return true
}

/*Remove - удаление из индекса*/
func (is IndexSlice) Remove(key uint32, id uint32) bool {
	data, ok := is[key]
	if !ok {
		return false
	}
	is[key] = data.Remove(id)
	return true
}

/*GetAll - все значения*/
func (is IndexSlice) GetAll(key uint32) (IndData, error) {
	data, ok := is[key]
	if !ok {
		return nil, fmt.Errorf("Нет ключа")
	}
	return data, nil
}

/*Update - обновление индекса для id*/
func (is IndexSlice) Update(id uint32, old uint32, nee uint32, wg *sync.WaitGroup) {
	if old == nee {
		return
	}
	wg.Add(2)
	var odata, ndata IndData
	odata, _ = is[old]
	go func() {
		odata = odata.Remove(id)
		wg.Done()
	}()
	ndata, ok := is[nee]
	go func() {
		if ok {
			ndata = ndata.Add(id)
		} else {
			ndata = IndSData([]uint32{id})
		}
		wg.Done()
	}()
	wg.Wait()
	is[old] = odata
	is[nee] = ndata
}

/*UpdateMany - изменение многих значений (для интересов)*/
func (is IndexSlice) UpdateMany(id uint32, old []uint16, nee []uint16, wg *sync.WaitGroup, lock *sync.Mutex) {
	for _, o := range old { // цикл удаления
		wg.Add(1)
		lock.Lock()
		r, _ := is[uint32(o)]
		lock.Unlock()
		go func(data IndData, n uint32) { // удаление из индекса
			data = data.Remove(id)
			lock.Lock()
			is[n] = data
			lock.Unlock()
			wg.Done()
		}(r, uint32(o))
	}
	for _, ne := range nee { // цикл добавления
		lock.Lock()
		add, ok := is[uint32(ne)]
		lock.Unlock()
		if ok {
			wg.Add(1)
			go func(data IndData, n uint32) {
				data = data.Add(id)
				lock.Lock()
				is[n] = data
				lock.Unlock()
				wg.Done()
			}(add, uint32(ne))
		} else {
			is[uint32(ne)] = IndSData([]uint32{id})
		}
	}
	wg.Wait() // ожидание окончания всех операций
}

/*IndexMData - индексные данные на основе карты*/
type IndexMData map[uint32]bool

/*Len - длина данных*/
func (imd IndexMData) Len() int {
	return len(imd)
}

/*Remove - удаление данных*/
func (imd IndexMData) Remove(id uint32) IndData {
	delete(imd, id)
	return imd
}

/*Add - добавление данных*/
func (imd IndexMData) Add(id uint32) IndData {
	imd[id] = false
	return imd
}

/*IsExists - наличие элемента*/
func (imd IndexMData) IsExists(id uint32) bool {
	_, ok := imd[id]
	return ok
}

/*GetAll - вывод всех данных в виде среза*/
func (imd IndexMData) GetAll() []uint32 {
	out := make([]uint32, 0, len(imd))
	for k := range imd {
		out = append(out, k)
	}
	return out
}

/*Intercept - пересечение двух данных*/
func (imd IndexMData) Intercept(ind IndData) IndData {
	m := make(map[uint32]bool)
	for k := range imd {
		if ind.IsExists(k) {
			m[k] = false
		}
	}
	return IndexMData(m)
}

/*Union - объединение двух данных*/
func (imd IndexMData) Union(ind IndData) IndData {
	m := make(map[uint32]bool)
	for k := range imd {
		m[k] = false
	}
	data, ok := ind.(IndSData)
	if ok {
		for _, k := range data {
			m[k] = false
		}
	}
	dm, ok := ind.(IndexMData)
	if ok {
		for k := range dm {
			m[k] = false
		}
	}

	return IndexMData(m)
}

/*IndexMap - индекс на основе map*/
type IndexMap map[uint32]IndexMData

/*In - вхождение*/
func (im IndexMap) In(key uint32, id uint32) bool {
	row, ok := im[key]
	if !ok {
		return false
	}
	_, ok = row[id]
	return ok
}

/*NewMI - новый индекс*/
func NewMI() IndexMap {
	return make(map[uint32]IndexMData)
}

/*Has - наличие ключа*/
func (im IndexMap) Has(key uint32) bool {
	_, ok := im[key]
	return ok
}

/*Add - Добавление элемента*/
func (im IndexMap) Add(key uint32, id uint32) bool {
	data, ok := im[key]
	if !ok {
		mp := make(map[uint32]bool)
		mp[id] = false
		im[key] = mp
		return false
	}
	data[id] = false
	im[key] = data
	return true
}

/*Remove - Удаление элемента*/
func (im IndexMap) Remove(key uint32, id uint32) bool {
	data, ok := im[key]
	if !ok { // ничего не удаляем
		return false
	}
	delete(data, id)
	return true
}

/*GetAll - вывод всех элементов в виде slice*/
func (im IndexMap) GetAll(key uint32) (IndData, error) {
	data, ok := im[key]
	if !ok {
		return nil, fmt.Errorf("Нет ключа")
	}
	return data, nil
}
