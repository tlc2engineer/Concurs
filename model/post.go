package model

import (
	"sort"
	"sync"
)

var lchan chan *LikeMess

/*UpdateChan - канал для обновления индексов*/
var UpdateChan = updateUser()

/*LikesTempB - pool для LTemp*/
var LikesTempB = sync.Pool{
	New: func() interface{} {
		return make([]LTemp, 300)
	},
}

/*UserPool - пул User*/
var UserPool = sync.Pool{
	New: func() interface{} {
		return new(User)
	},
}

/*GetLikeCh - получение канала загрузки Like*/
func GetLikeCh() chan *LikeMess {
	if lchan != nil {
		return lchan
	}
	return likeFactory()
}
func likeFactory() chan *LikeMess {
	ch := make(chan *LikeMess, 10)
	go func() {
		for mess := range ch {
			n := mess.Num
			switch n {
			case 0: // новые лайки с новым аккаунтом
				likes := mess.Likes
				id := mess.ID
				likes = NormLikes(likes)
				sort.Slice(likes, func(i, j int) bool {
					return likes[i].ID < likes[j].ID
				})
				inLikes := PackLSlice(likes)
				SetLikes(uint32(id), inLikes)
				AddWhos(uint32(id), likes)
			case 1: // список новых лайков
				likesT := mess.Ltemps
				for _, like := range likesT {
					id := like.Liker
					id2 := like.Likee
					ts := like.Ts
					data := GetLikes(uint32(id))
					// добавление
					found := false
					for i := 0; i < len(data)/8; i++ {
						addr := i * 8
						tid := uint32(data[addr]) | uint32(data[addr+1])<<8 | uint32(data[addr+2])<<16
						if tid == uint32(id2) { // уже есть лайк
							cnt := data[i*8+7]
							cnt++
							if cnt > 255 {
								cnt = 255
							}
							data[i*8+7] = cnt
							found = true
							break
						}
						if tid > uint32(id2) {
							found = true
							l := Like{Ts: float64(ts), ID: id2, Num: 1}
							p := LikePack(l)
							//ndata := make([]byte, len(data)+8)
							data = append(data, make([]byte, 8)...) // добавляем 8 байт
							//copy(ndata[:addr], data[:addr])
							copy(data[addr+8:], data[addr:])
							copy(data[addr:], p)

							AddWho(uint32(id), l)
							break
						}
					}
					if !found { // у аккаунта нет лайков на того же пользователя
						l := Like{Ts: float64(ts), ID: id2, Num: 1}
						p := LikePack(l)
						ndata := make([]byte, len(data)+8)
						if true { // добавление в конец списка
							copy(ndata[:len(data)], data)
							copy(ndata[len(data):], p)
							data = ndata
						}
						AddWho(uint32(id), l)
					}
					SetLikes(uint32(id), data)
				}
				LikesTempB.Put(likesT)
			case 2: //обновление
				likes := mess.Likes
				id := mess.ID
				likes = NormLikes(likes) // нормируем
				sort.Slice(likes, func(i, j int) bool {
					return likes[i].ID < likes[j].ID
				})
				SetLikes(uint32(id), PackLSlice(likes))
				AddWhos(uint32(id), likes)
				//Удалить старые лайки которых уже нет!
				oldLikes := GetLikes(uint32(id)) // старые лайки
				ids := make([]uint32, 0)         // список несовпадающих лайков
				for i := 0; i < len(oldLikes)/8; i++ {
					var idLike uint32 // старый id
					idLike = uint32(oldLikes[0]) | uint32(oldLikes[1])<<8 | uint32(oldLikes[2])<<16
					found := false
					for _, like := range likes {
						if uint32(like.ID) == idLike {
							found = true
							break
						}
					}
					if !found {
						ids = append(ids, idLike)
					}
				}
				// Цикл по номерам которые уже не предпочитает
				for _, tid := range ids {
					data, _ := GetWho(tid) // кто лайкал данный id
					SetWho(uint32(tid), data.RemoveId(uint32(id)))
				}
			}

		}
	}()
	return ch
}

/*LikeMess - сообщение для загрузки Like в память*/
type LikeMess struct {
	Num    int
	Likes  []Like
	Ltemps []LTemp
	ID     uint32
}

/*LTemp - промежуточная структура*/
type LTemp struct {
	Liker int64 `json:"liker"`
	Likee int64 `json:"likee"`
	Ts    int64 `json:"ts"`
}
type LikesT struct {
	Likes []LTemp `json:"likes"`
}

/*UserMess - сообщение содержит старый и обновленный аккаунт*/
type UserMess struct {
	OldUser *User
	NewUser *User
}

func updateUser() chan *UserMess {
	ch := make(chan *UserMess, 10)
	go func() {
		for mess := range ch {
			old := mess.OldUser
			nee := mess.NewUser
			defer UserPool.Put(nee)
			if old != nil {
				RemRecIndex(*old)
				DeleteGIndex(*old)
				defer UserPool.Put(old)
			}
			AddGIndex(*nee)
			AddRecIndex(*nee)

		}
	}()
	return ch
}
