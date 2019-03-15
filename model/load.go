package model

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

func loadLikeFactory() chan *likeData {
	ch := make(chan *likeData, 10)
	go func() {
		for ldata := range ch {
			tl := NormLikes(ldata.likes)
			SetLikes(uint32(ldata.id), PackLSlice(tl))
			for _, like := range tl {
				AddWho(uint32(ldata.id), like)
			}

		}
	}()
	return ch
}

type likeData struct {
	id    uint32
	likes []Like
}

/*AddUsers - Добавление пользователей из файла*/
func AddUsers(wg *sync.WaitGroup) chan []byte {
	out := make(chan []byte, 20)
	go func() {
		for bts := range out {
			bg := time.Now()
			data, _, _, err := jsonparser.Get(bts, "accounts")
			if err != nil {
				panic(err)
			}
			lfactory := loadLikeFactory()
			insertUser := func(user User) {
				ln := len(users)
				if ln > 0 && user.ID < users[ln-1].ID { // если больше последнего элемента добавляем в конец иначе вставка
					i := ln - 1 // начиная с последнего элемента
					for i >= 0 && user.ID < users[ln-1].ID {
						i--
					}
					//вставка
					users = append(users, User{})
					//fmt.Println(user.ID)
					copy(users[i+2:], users[i+1:]) // вставляем id перед i
					users[i+1] = user
				} else {
					users = append(users, user)
				}
			}
			_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				acc := new(Account)
				acc.UnmarshalJSON(value)
				ldata := new(likeData)
				ldata.id = uint32(acc.ID)
				ldata.likes = acc.Likes
				lfactory <- ldata
				user := Conv(*acc)
				insertUser(user)
			})
			if err != nil {
				panic(err)
			}
			close(lfactory)
			wg.Done()
			ReadData.Put(bts)
			fmt.Println("te", time.Since(bg))
			runtime.GC()
		}
	}()
	return out
}
