package handlers

import (
	"Concurs/model"
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
)

/*AddLikes - новые предпочтения*/
func AddLikes(ctx *fasthttp.RequestCtx) {
	mutex := model.WrMutex
	mutex.Lock()
	defer mutex.Unlock()
	if !ctx.QueryArgs().Has(queryParam) {
		ctx.SetStatusCode(400)
		return
	}
	data := ctx.PostBody()
	vmap := make(map[string][]Temp)
	err := json.Unmarshal(data, &vmap)
	if err != nil {
		fmt.Println("Ошибка распаковки")
		ctx.SetStatusCode(400)
		return
	}
	likesT, ok := vmap["likes"]
	if !ok {
		fmt.Println("Ошибка распаковки 1")
		ctx.SetStatusCode(400)
		return
	}
	// проверка что есть такие like id
	for _, like := range likesT {
		id := like.Likee
		_, err := model.GetAccountPointer(uint32(id))
		if err != nil {
			fmt.Println("Нет такого id кто %d", id)
			ctx.SetStatusCode(400)
			return
		}
		id = like.Liker
		_, err = model.GetAccountPointer(uint32(id))
		if err != nil {
			fmt.Println("Нет такого id кого")
			ctx.SetStatusCode(400)
			return
		}
	}
	// добавление like
	for _, like := range likesT {
		id := like.Liker
		id2 := like.Likee
		ts := like.Ts
		data := model.GetLikes(uint32(id))
		// добавление
		found := false
		for i := 0; i < len(data)/8; i++ {
			l := model.LikeUnPack(data[i*8 : i*8+8])
			if l.ID == id2 { // уже есть лайк
				cnt := data[i*8+7]
				cnt++
				data[i*8+7] = cnt
				found = true
				break
			}
		}

		if !found { // у аккаунта нет лайков на того же пользователя
			l := model.Like{Ts: float64(ts), ID: id2, Num: 1}
			p := model.LikePack(l)
			model.AddWho(uint32(id), l)
			data = append(data, p...)
			model.SetLikes(uint32(id), data)
		}
		//model.LikesMap[uint32(id)] = data
	}
	// окончание
	ctx.SetStatusCode(202) // все в норме
	ctx.Write([]byte(""))
	return

}

/*Temp - промежуточная структура*/
type Temp struct {
	Liker int64   `json:"liker"`
	Likee int64   `json:"likee"`
	Ts    float64 `json:"ts"`
}
