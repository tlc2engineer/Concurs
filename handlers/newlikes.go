package handlers

import (
	"Concurs/model"

	"github.com/valyala/fasthttp"
)

/*AddLikes - новые предпочтения*/
func AddLikes(ctx *fasthttp.RequestCtx) {
	mutex.Lock()
	defer mutex.Unlock()

	if !ctx.QueryArgs().Has(queryParam) {
		ctx.SetStatusCode(400)
		return
	}
	data := ctx.PostBody()
	likes := new(model.LikesT)
	likesT := model.LikesTempB.Get().([]model.LTemp)
	likesT = likesT[:0]
	likes.Likes = likesT
	//defer model.LikesTempB.Put(likesT)
	err := likes.UnmarshalJSON(data)
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}
	likesT = likes.Likes
	// проверка что есть такие like id
	for _, like := range likesT {
		//fmt.Println(like)
		if like.Ts <= 0 {
			ctx.SetStatusCode(400)
			return
		}
		id := like.Likee
		_, err := model.GetAccountPointer(uint32(id))
		if err != nil {
			//fmt.Println("Нет такого id кто %d", id)
			ctx.SetStatusCode(400)
			return
		}
		id = like.Liker
		_, err = model.GetAccountPointer(uint32(id))
		if err != nil {
			//fmt.Println("Нет такого id кого")
			ctx.SetStatusCode(400)
			return
		}
	}
	//------------------
	lch := model.LikeCh
	lmess := model.LikeMess{
		Num:    1,
		Ltemps: likesT,
	}
	lch <- &lmess
	//------------------
	/* добавление like
	for _, like := range likesT {
		id := like.Liker
		id2 := like.Likee
		ts := like.Ts
		data = model.GetLikes(uint32(id))
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
				l := model.Like{Ts: float64(ts), ID: id2, Num: 1}
				p := model.LikePack(l)
				//ndata := make([]byte, len(data)+8)
				data = append(data, make([]byte, 8)...) // добавляем 8 байт
				//copy(ndata[:addr], data[:addr])
				copy(data[addr+8:], data[addr:])
				copy(data[addr:], p)

				model.AddWho(uint32(id), l)
				break
			}
		}
		if !found { // у аккаунта нет лайков на того же пользователя
			l := model.Like{Ts: float64(ts), ID: id2, Num: 1}
			// lks := model.UnPackLSlice(data)
			// lks = append(lks, l)
			// sort.Slice(lks, func(i, j int) bool {
			// 	return lks[i].ID < lks[j].ID
			// })
			// data = model.PackLSlice(lks)
			p := model.LikePack(l)
			ndata := make([]byte, len(data)+8)
			if true { // добавление в конец списка
				copy(ndata[:len(data)], data)
				copy(ndata[len(data):], p)
				data = ndata
			}
			model.AddWho(uint32(id), l)
		}
		model.SetLikes(uint32(id), data)
	}*/

	// окончание
	ctx.SetStatusCode(202) // все в норме
	ctx.Write([]byte(""))
	return

}

func add(like model.LTemp) error {
	id := like.Liker
	id2 := like.Likee
	ts := like.Ts
	data := model.GetLikes(uint32(id))
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
			l := model.Like{Ts: float64(ts), ID: id2, Num: 1}
			p := model.LikePack(l)
			//ndata := make([]byte, len(data)+8)
			data = append(data, make([]byte, 8)...) // добавляем 8 байт
			//copy(ndata[:addr], data[:addr])
			copy(data[addr+8:], data[addr:])
			copy(data[addr:], p)

			model.AddWho(uint32(id), l)
			break
		}
	}
	if !found { // у аккаунта нет лайков на того же пользователя
		l := model.Like{Ts: float64(ts), ID: id2, Num: 1}
		// lks := model.UnPackLSlice(data)
		// lks = append(lks, l)
		// sort.Slice(lks, func(i, j int) bool {
		// 	return lks[i].ID < lks[j].ID
		// })
		// data = model.PackLSlice(lks)
		p := model.LikePack(l)
		ndata := make([]byte, len(data)+8)
		if true { // добавление в конец списка
			copy(ndata[:len(data)], data)
			copy(ndata[len(data):], p)
			data = ndata
		}
		model.AddWho(uint32(id), l)
	}
	model.SetLikes(uint32(id), data)
	return nil
}
