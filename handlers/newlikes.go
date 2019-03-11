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
	likes := new(LikesT)
	likesT := likesTempB.Get().([]LTemp)
	likesT = likesT[:0]
	likes.Likes = likesT
	defer likesTempB.Put(likesT)
	err := likes.UnmarshalJSON(data)
	//likeData, _, _, err := jsonparser.Get(data, ("likes"))
	//likesT := likes.Likes
	if err != nil {
		ctx.SetStatusCode(400)
		//fmt.Println("Not likes")
		return
	}
	likesT = likes.Likes
	//fmt.Println(cap(likesT), likesT)
	//fmt.Println(likesT, cap(likesT))
	//fmt.Println("likes", len(likesT))
	// likesT := likesTempB.Get().([]lTemp)
	// likesT = likesT[:0]
	// defer likesTempB.Put(likesT)
	// errFlag := false
	// jsonparser.ArrayEach(likeData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	// 	//"likee":22727,"ts":1492927159,"liker":21102
	// 	if err != nil {
	// 		errFlag = true
	// 	}
	// 	if !errFlag {
	// 		likee, err := jsonparser.GetInt(value, "likee") //int64(getInt(value, "likee"))
	// 		if err != nil {
	// 			errFlag = true
	// 		}
	// 		ts, err := jsonparser.GetInt(value, "ts")
	// 		if err != nil {
	// 			errFlag = true
	// 		}
	// 		liker, err := jsonparser.GetInt(value, "liker")
	// 		if err != nil {
	// 			errFlag = true
	// 		}
	// 		// fmt.Println(string(value))
	// 		// fmt.Println(likee, ts, liker)
	// 		likesT = append(likesT, lTemp{liker, likee, float64(ts)})
	// 	}
	// })
	// if errFlag {
	// 	ctx.SetStatusCode(400)
	// 	return
	// }
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
	// добавление like
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
	}

	// окончание
	ctx.SetStatusCode(202) // все в норме
	ctx.Write([]byte(""))
	return

}

/*Temp - промежуточная структура*/
type LTemp struct {
	Liker int64 `json:"liker"`
	Likee int64 `json:"likee"`
	Ts    int64 `json:"ts"`
}
type LikesT struct {
	Likes []LTemp `json:"likes"`
}
