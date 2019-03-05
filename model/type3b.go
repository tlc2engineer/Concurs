package model

/*Id3B - Хранение данных в 3 байтах*/
type Id3B []byte

/*Get - получение в uint32*/
func (data Id3B) Get() []uint32 {
	ln := len(data) / 3
	out := make([]uint32, ln)
	for i := 0; i < ln; i++ {
		out[i] = uint32(data[i*3]) | uint32(data[i*3+1])<<8 | uint32(data[i*3+2])<<16
	}
	return out
}

/*Len - длина данных*/
func (data Id3B) Len() int {
	return len(data) / 3
}

/*GetId - возврат элемента*/
func (data Id3B) GetId(num int) uint32 {
	return uint32(data[num*3]) | uint32(data[num*3+1])<<8 | uint32(data[num*3+2])<<16
}

/*Add - добавление id в список*/
func (data Id3B) Add(id uint32) Id3B {
	if data.IsExist(id) {
		return data
	}
	// добавление нового id в конец списка
	b0 := byte(id)
	b1 := byte(id >> 8)
	b2 := byte(id >> 16)
	data = append(data, b0, b1, b2)
	return data
}

/*IsExist - есть ли такой id*/
func (data Id3B) IsExist(id uint32) bool {
	for i := 0; i < data.Len(); i++ {
		if id == data.GetId(i) { // такой id  уже есть
			return true
		}
	}
	return false
}

/*RemoveId - удаление из списка*/
func (data Id3B) RemoveId(id uint32) Id3B {
	for i := 0; i < data.Len(); i++ {
		if id == data.GetId(i) { // такой id  уже есть
			ln := len(data)
			data = data[:i+copy(data[i:], data[i+3:])] //copy(data[:i*3], data[:i*3+3])
			data = data[:ln-3]
			return data
		}
	}
	return data
}

/*NewId3b - новые данные из массива uint32*/
func NewId3b(data []uint32) Id3B {
	ln := len(data) * 3
	out := make([]byte, ln)
	for i := 0; i < len(data); i++ {
		out[i*3] = byte(data[i])
		out[i*3+1] = byte(data[i] >> 8)
		out[i*3+2] = byte(data[i] >> 16)
	}
	return out
}

/*NewId3 - новые данные из id*/
func NewId3(id uint32) Id3B {
	out := make([]byte, 3)
	out[0] = byte(id)
	out[1] = byte(id >> 8)
	out[2] = byte(id >> 16)
	return out
}
