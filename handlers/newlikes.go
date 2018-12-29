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
		_, err := model.GetAccountPointer(int(id))
		if err != nil {
			fmt.Println("Нет такого id кто %d", id)
			ctx.SetStatusCode(400)
			return
		}
		id = like.Liker
		_, err = model.GetAccountPointer(int(id))
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
		pacc, _ := model.GetAccountPointer(int(id))
		// добавление
		likes := pacc.Likes
		found := false
		for i := range likes {
			if likes[i].ID == id2 { // уже есть лайк
				nts := (ts + likes[i].Ts*float64(likes[i].Num)) / float64(1+likes[i].Num) // новый ts
				likes[i].Num = likes[i].Num + 1
				likes[i].Ts = nts
				found = true
				break
			}
		}
		if !found { // у аккаунта нет лайков на того же пользователя
			likes = append(likes, model.Like{ID: id2, Ts: ts, Num: 1})
		}
		pacc.Likes = likes
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
