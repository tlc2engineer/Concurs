package model

import (
	"encoding/binary"
	"fmt"
)

/*Like - Лайк*/
type Like struct {
	Ts  float64 `json:"ts"`
	ID  int64   `json:"id"`
	Num uint8
}

/*LikePack - Упаковка Like*/
func LikePack(like Like) []byte {
	data := make([]byte, 8)
	id := uint32(like.ID)
	data[7] = byte(like.Num)
	data[0] = byte(id)
	data[1] = byte(id >> 8)
	data[2] = byte(id >> 16)
	ts := uint32(like.Ts)
	binary.LittleEndian.PutUint32(data[3:7], ts)
	return data
}

/*LikeUnPack - Распаковка Like*/
func LikeUnPack(data []byte) Like {
	like := Like{}
	var id uint32
	id = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 //int(data[0]) + int(data[1])<<8 + int(data[2])<<16
	ts := binary.LittleEndian.Uint32(data[3:7])
	like.ID = int64(id)
	like.Num = data[7]
	like.Ts = float64(ts)
	return like
}

/*PackLSlice - упаковка slice*/
func PackLSlice(likes []Like) []byte {
	ln := len(likes)
	out := make([]byte, ln*8)
	for i := 0; i < ln; i++ {
		copy(out[i*8:i*8+8], LikePack(likes[i]))
	}
	return out
}

/*UnPackLSlice - распаковка slice*/
func UnPackLSlice(data []byte) []Like {
	ln := len(data) / 8
	out := make([]Like, 0, ln)
	for i := 0; i < ln; i++ {
		like := LikeUnPack(data[i*8 : i*8+8])
		out = append(out, like)
	}
	return out
}

/*LikeByte - 8 байт на лайк*/
type LikeByte []byte

/*GetLikes - Получение лайков*/
func GetLikes(id uint32) []byte {
	return likesM[id]
}

/*SetLikes - установка лайков*/
func SetLikes(id uint32, data []byte) {
	likesM[id] = data
}

/*AddWho - добавление информации о том кто лайкал*/
func AddWho(id uint32, like Like) {
	_, err := GetWho(uint32(like.ID))
	if err != nil { // нет в карте
		SetWho(uint32(like.ID), NewId3(id))
	} else {
		w, _ := GetWho(uint32(like.ID))
		SetWho(uint32(like.ID), w.Add(id))
	}

}

/*AddWhos - добавление группы like о том кто лайкал*/
func AddWhos(id uint32, likes []Like) {
	for _, like := range likes {
		AddWho(id, like)
	}
}

/*GetWho - кого лайкали получить*/
func GetWho(id uint32) (Id3B, error) {
	//if !budger || true {
	data, ok := whoMap[id]
	if ok {
		return data, nil
	} else {
		return nil, fmt.Errorf("Нет")
	}
}

/*SetWho - Установить*/
func SetWho(id uint32, data Id3B) {
	whoMap[id] = data
}

func packWho(data []uint32) []byte {
	out := make([]byte, len(data)*4)
	for i := range data {
		binary.LittleEndian.PutUint32(out[i*4:(i+1)*4], data[i])
	}
	return out
}

func unPackWho(data []byte) []uint32 {
	out := make([]uint32, len(data)/4)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])

	}
	return out
}

// func likeTo64(like Like) uint64 {
// 	var out uint64
// 	id := like.ID
// 	num := like.Num
// 	ts := uint32(like.Ts)
// 	like := Like{}
// 	var id uint32
// 	id = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 //int(data[0]) + int(data[1])<<8 + int(data[2])<<16
// 	ts := binary.LittleEndian.Uint32(data[3:7])
// 	like.ID = int64(id)
// 	like.Num = data[7]
// 	like.Ts = float64(ts)
// 	return out
// }
