package handlers

import (
	"Concurs/model"
	"bytes"
	"fmt"

	"github.com/buger/jsonparser"

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
	likeData, _, _, err := jsonparser.Get(data, ("likes"))
	if err != nil {
		ctx.SetStatusCode(400)
		fmt.Println("Not likes")
		return
	}
	likesT := likesTempB.Get().([]Temp)
	likesT = likesT[:0]
	defer likesTempB.Put(likesT)
	errFlag := false
	jsonparser.ArrayEach(likeData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//"likee":22727,"ts":1492927159,"liker":21102
		if err != nil {
			errFlag = true
		}
		if !errFlag {
			likee, err := jsonparser.GetInt(value, "likee") //int64(getInt(value, "likee"))
			if err != nil {
				errFlag = true
			}
			ts, err := jsonparser.GetInt(value, "ts")
			if err != nil {
				errFlag = true
			}
			liker, err := jsonparser.GetInt(value, "liker")
			if err != nil {
				errFlag = true
			}
			// fmt.Println(string(value))
			// fmt.Println(likee, ts, liker)
			likesT = append(likesT, Temp{liker, likee, float64(ts)})
		}
	})
	if errFlag {
		ctx.SetStatusCode(400)
		return
	}
	// проверка что есть такие like id
	for _, like := range likesT {
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
type Temp struct {
	Liker int64   `json:"liker"`
	Likee int64   `json:"likee"`
	Ts    float64 `json:"ts"`
}

/*getInt - получение целого*/
func getInt(val []byte, name string) uint32 {
	fbyte := []byte(name)
	fbyte = append(fbyte, 34, 58) //":
	ind := bytes.Index(val, fbyte)
	if ind == -1 {
		return 0
	}
	i := ind + 3 + len(name)
	beg := i
	var out int
	for val[i] >= 48 && val[i] < 59 {
		i++
	}
	end := i
	mul := 1
	for i := end - 1; i >= beg; i-- {
		out += int(val[i]-48) * mul
		mul = mul * 10
	}
	return uint32(out)
}
