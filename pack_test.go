package main

import (
	"Concurs/model"
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

/*
func TestSomething(t *testing.T) {

	v := 34.32244
	v1 := uint32(v)
	if v1 != 34 {
		t.Error("Не вышло")
	}
	now := uint32(time.Now().Unix())
	like := model.Like{ID: 1710, Ts: float64(now), Num: 2}
	data := model.LikePack(like)
	nlike := model.LikeUnPack(data)
	if like.Num != nlike.Num {
		t.Error("Не тот номер")
	}
	if like.ID != nlike.ID {
		t.Errorf("Не тот ID %d,%d", like.ID, nlike.ID)
	}
	if like.Ts != nlike.Ts {
		t.Errorf("Не тот Ts %f,%f", like.Ts, nlike.Ts)
	}
	likes := []model.Like{model.Like{ID: 15, Ts: float64(now), Num: 3}, model.Like{ID: 21, Ts: float64(now - 17), Num: 4}}
	data = model.PackLSlice(likes)
	nLikes := model.UnPackLSlice(data)
	if len(likes) != len(nLikes) {
		t.Errorf("Длина не совпадает")
	}
	for i := range likes {
		if likes[i].ID != nLikes[i].ID || likes[i].Ts != nLikes[i].Ts || likes[i].Num != nLikes[i].Num {
			t.Errorf("Ошибка упаковки")
		}
	}

}

func TestBudger(t *testing.T) {
	likes := model.UnPackLSlice(model.GetLikes(25006))
	for _, like := range likes {
		fmt.Println("--------", like.ID)
	}
}*/

/*
func TestBuger(t *testing.T) {
	bdata := make([]byte, 20000000)
	file, err := os.Open("./data/accounts_1.json")
	if err != nil {
		t.Errorf("Ошибка чтения")
	}
	n, err := file.Read(bdata)
	if err != nil {
		t.Errorf("Ошибка чтения 1")
	}
	bts := bdata[:n]
	data, tp, _, err := jsonparser.Get(bts, "accounts")
	if err != nil {
		t.Errorf("Ошибка чтения 2 %s", err)
	}
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		id, err := jsonparser.GetInt(value, "id")
		birth, err := jsonparser.GetInt(value, "birth")
		joined, err := jsonparser.GetInt(value, "joined")
		fname, err := jsonparser.GetString(value, "fname")
		sname, err := jsonparser.GetString(value, "sname")
		email, err := jsonparser.GetString(value, "email")
		phone, err := jsonparser.GetString(value, "phone")
		country, err := jsonparser.GetString(value, "country")
		city, err := jsonparser.GetString(value, "city")
		sex, err := jsonparser.GetString(value, "sex")
		status, err := jsonparser.GetString(value, "status")
		start, err := jsonparser.GetInt(value, "premium", "start")
		finish, err := jsonparser.GetInt(value, "premium", "finish")
		interests := make([]string, 0)
		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			interests = append(interests, string(utf8Unescaped(value)))
			//fmt.Printf("%s \n", "\u0410\u0432\u0430\u0442\u0430\u0440")
		}, "interests")
		likes := make([]model.Like, 0)
		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			lid, _ := jsonparser.GetInt(value, "id")
			ts, _ := jsonparser.GetInt(value, "ts")
			likes = append(likes, model.Like{ID: lid, Ts: float64(ts), Num: 0})
		}, "likes")
		fmt.Println(id, sex, fname, sname, email, phone, country, city, birth, joined, status, start, finish)
		fmt.Println(interests)
	})
	fmt.Println("-----", tp, len(data))
	//out := "\u554a\u0416\u0438\u0437\u043d\u044c \u0421\u043e\u043b\u043d\u0446\u0435 \u0411\u0443\u0440\u0433\u0435\u0440\u044b"
	//
	//fmt.Printf("%s /n", out)

}
*/
// хак для перевода экранированных строк вида "\u1234\u5678" в нормальный юникод
func utf8Unescaped(b []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte('"')
	buf.Write(b)
	buf.WriteByte('"')

	var s string
	json.Unmarshal(buf.Bytes(), &s)

	return []byte(s)
}

/*
func TestIndex(t *testing.T) {
	is := model.NewIS()
	is.Add(14, 22)
	is.Add(14, 41)
	is.Add(14, 17)
	is.Add(14, 144)
	idat, err := is.GetAll(14)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(idat.GetAll())
	is.Remove(14, 41)
	idat, err = is.GetAll(14)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(idat.GetAll())
	is.Remove(14, 17)
	idat, err = is.GetAll(14)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(idat.GetAll())
	// s1 := []uint32{0, 1, 3, 5, 7, 9, 14, 33, 36, 46, 50}
	// s2 := []uint32{0, 2, 4, 5, 7, 11, 14, 20, 25, 50}
	// res := model.Intercept(s1, s2)
	// fmt.Println(res)
	// res = model.UnionSorted(s1, s2)
	// fmt.Println(res)
	// ind1 := model.IndSData(s1)
	// ind2 := model.IndSData(s2)
	// ind2.Add(14)
	// ind3 := ind1.Intercept(ind2)
	// fmt.Println(ind3.GetAll())
	// ind4 := ind1.Union(ind2)
	// fmt.Println(ind4.GetAll())
	n1 := []uint32{29592, 29674, 29767, 29824, 29838, 29860, 29926, 29938, 29983, 29999}
	n2 := []uint32{29705, 29732, 29768, 29777, 29810, 29846, 29861, 29935, 29992, 29997}
	res := model.Union2Sorted(n1, n2)
	fmt.Println("R", res)
}
*/
/*
func TestRes(t *testing.T) {
	var num int
	prod := resource.New32Res(20)
	res, _ := prod.Get(3)

	if len(res) != 3 {
		t.Errorf("Ошибка")
	}
	res2, num2 := prod.Get(15)
	if len(res2) != 15 {
		t.Errorf("Ошибка")
	}
	res3, num := prod.Get(4)
	if num != -1 && res3 != nil {
		t.Errorf("Ошибка")
	}
	prod.Release(num2)
	res3, num4 := prod.Get(4)
	if num4 == -1 || len(res3) != 4 {
		t.Errorf("Ошибка")
	}
}
*/
/*
func TestJson(t *testing.T) {
	bts := make([]byte, 1000)
	bg := "{\"accounts\":["
	end := "]}"
	buff := bytes.NewBuffer(bts)
	buff.Reset()
	buff.WriteString(bg)
	dat := make(map[string]interface{})
	dat["email"] = "tlc2engineer@gmail.com"
	dat["id"] = "1201"
	dat["sex"] = "m"
	dat["city"] = "Алчевск"
	tmp := make([]byte, 100)
	tBuff := bytes.NewBuffer(tmp)
	tBuff.Reset()
	enc := json.NewEncoder(tBuff)
	err := enc.Encode(dat)
	if err != nil {
		t.Errorf("%s", err)
	}
	l := tBuff.Len()
	buff.Write(tmp[:l])
	//buff.WriteString(",")
	buff.WriteString(end)
	fmt.Println(string(bts[:buff.Len()]))
	fmt.Println(json.Valid(bts[:buff.Len()]))
}
*/
/*
func TestZip(t *testing.T) {
	//fnames := make([]string, 0)
	r, err := zip.OpenReader(base + "data.zip")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer r.Close()
	data := make([]byte, 0, 2000000)
	tmp := make([]byte, 32768)
	for _, f := range r.File {

		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		for {
			_, err := rc.Read(tmp)
			if err == io.EOF {
				data = append(data, tmp...)
				//fmt.Println("1", n)
				break
			}
			data = append(data, tmp...)
			//fmt.Println("1", n)

		}
		fmt.Println(len(data))
		rc.Close()
		fmt.Println()
	}
}
*/
/*
func TestSort(t *testing.T) {
	ids := []int{22, 3, 9, 456, 0}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	fmt.Println(ids)

}*/
/*
func TestLikes(t *testing.T) {
	like1 := model.Like{ID: 9034, Ts: 12344, Num: 4}
	like2 := model.Like{ID: 9035, Ts: 12344, Num: 2}
	like3 := model.Like{ID: 6000, Ts: 12344, Num: 2}
	like4 := model.Like{ID: 11000, Ts: 12344, Num: 2}
	likes := make([]model.Like, 0)
	likes = append(likes, like1, like2, like3, like4)
	sort.Slice(likes, func(i, j int) bool {
		return likes[i].ID < likes[j].ID
	})
	fmt.Println(likes)
	inLike := model.Like{ID: 20000, Ts: 12344, Num: 2}
	data := model.PackLSlice(likes)
	p := model.LikePack(inLike)
	ndata := make([]byte, len(data)+8)
	f := false
	for i := 0; i < len(data)/8; i++ {
		addr := i * 8
		id := uint32(data[addr]) | uint32(data[addr+1])<<8 | uint32(data[addr+2])<<16
		if uint32(inLike.ID) < id {
			fmt.Println(i)
			copy(ndata[:addr], data[:addr])
			copy(ndata[addr+8:], data[addr:])
			copy(ndata[addr:], p)
			// copy(data[i+8:], data[i:]) // вставляем id перед i
			// copy(data[i:i+8], p)       // вставка
			data = ndata
			f = true
			break
		}
	}
	fmt.Println(f)
	if !f { // добавление в конец списка
		copy(ndata[:len(data)], data)
		copy(ndata[len(data):], p)
		data = ndata
	}
	nlikes := model.UnPackLSlice(data)
	fmt.Println(nlikes)
}
*/

func contains(keys []string, ref []string) bool {

	for i := 0; i < len(keys); i++ {
		find := false
		for j := 0; j < len(ref); j++ {
			if ref[j] == keys[i] {
				find = true
				break
			}
		}
		if !find {
			return false
		}
	}
	return true
}

/*
func TestMap(t *testing.T) {
	nodes := make([]rgbtree.Node, 50)
	tree := rgbtree.NewUTree(nodes)
	tree.Put(22, 11)
	tree.Put(7, 48)
	tree.Put(40, 99)
	tree.Put(74, 1)
	//tree.Remove(40)
	fmt.Println(tree.String())

}
*/

func TestLoad(t *testing.T) {
	now := time.Now()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	ft := false
	//----------------------------------------
	file, err := os.Open(base + opt)
	if err != nil {
		model.Now = time.Now().Unix()
		ft = true
		//panic(err)
	}
	if !ft {
		reader := bufio.NewReader(file)
		line, _, err := reader.ReadLine()
		if err != nil {
			panic(err)
		}
		model.Now, err = strconv.ParseInt(string(line), 10, 0)
		if err != nil {
			panic(err)
		}
	}
	//------открываем zip-------------
	fnames := make([]string, 0)
	r, err := zip.OpenReader(base + dfname)
	if err != nil {
		panic(err)
	}
	for _, f := range r.File {
		fnames = append(fnames, f.Name)
	}
	num := len(fnames)
	bdata := make([]byte, 0, 20000000)
	//sdata := make([]byte, 0, 20000000)
	tmp := make([]byte, 32768)
	wg := &sync.WaitGroup{}
	uch := model.AddUsers(wg)
	for i := 1; i <= num; i++ {
		//fmt.Println("Номер ", i)
		fname := fmt.Sprintf("%saccounts_%d.json", base, i)
		for _, f := range r.File {
			if base+f.Name == fname {
				bdata = bdata[:0]
				rc, err := f.Open()
				if err != nil {
					panic("Ошибка чтения")
				}
				for {
					_, err := rc.Read(tmp)
					if err != nil {
						if err == io.EOF {
							bdata = append(bdata, tmp...)
							break
						}
						panic(err)
					}
					bdata = append(bdata, tmp...)
				}
				ndata := make([]byte, len(bdata))
				copy(ndata, bdata)
				wg.Add(1)
				uch <- ndata
			}
		}
	}
	r.Close()
	close(uch)
	wg.Wait()
	fmt.Println(time.Since(now))
}

/*
func TestNorm(t *testing.T) {
	likes := []model.Like{model.Like{Ts: 23344, ID: 2, Num: 1}, model.Like{Ts: 23344, ID: 1, Num: 1}, model.Like{Ts: 34344, ID: 1, Num: 1}, model.Like{Ts: 34344, ID: 1, Num: 1}}
	likes = model.NormLikes(likes)
	fmt.Println(likes)
}
*/
