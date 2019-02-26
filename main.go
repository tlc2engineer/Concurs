package main

import (
	"Concurs/handlers"
	"Concurs/model"
	"Concurs/util"
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

const opt = "options.txt"
const base = "./data/" //D:/install/elim_accounts_261218/data/data/" //"D:/install/elim_accounts_261218/data/data/" //"/home/sergey/Загрузки/data/data/"///home/sergey/Загрузки/test_accounts_220119/data/
const dfname = "data.zip"
const addr = ":8080"

func main() {
	gen := len(os.Args) > 1 && os.Args[1] == "gen"
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
					n, err := rc.Read(tmp)
					if err != nil {
						if err == io.EOF {
							bdata = append(bdata, tmp[:n]...)
							break
						}
						panic(err)
					}
					bdata = append(bdata, tmp...)
				}
				// sql флаг
				if gen {
					dat := make(map[string][]model.Account)
					err := json.Unmarshal(bdata, &dat)
					if err != nil {
						fmt.Println("json")
						panic(err)
					}
					err = util.CreateTables(dat["accounts"])
					if err != nil {
						panic(err)
					}
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
	model.SetUsers()
	//go clear()
	//debug.SetGCPercent(100)
	router := fasthttprouter.New()
	router.GET("/accounts/*path", requestGet)
	router.POST("/accounts/*path", requestPost)
	log.Fatal(fasthttp.ListenAndServe(addr, router.Handler))
}

func getData(fname string) ([]model.Account, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	dat := make(map[string][]model.Account)
	bts, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bts, &dat)
	if err != nil {
		return nil, err
	}
	accounts, ok := dat["accounts"]
	if !ok {
		return nil, err
	}
	return accounts, nil
}

func requestGet(ctx *fasthttp.RequestCtx) {
	path := ctx.UserValue("path").(string)
	args := strings.Split(path, "/")
	switch len(args) {
	case 3:
		switch args[1] {
		case "filter":
			handlers.Filter(ctx)
			return
		case "group":
			handlers.Group(ctx)
			return
		}
	case 4:
		id, err := strconv.ParseInt(args[1], 10, 0)
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		switch args[2] {
		case "recommend":
			handlers.Recommend(ctx, int(id))
			return
		case "suggest":
			handlers.Suggest(ctx, int(id))
			return
		}
	}
	ctx.SetStatusCode(404)
	return

}

func requestPost(ctx *fasthttp.RequestCtx) {
	path := ctx.UserValue("path").(string)
	args := strings.Split(path, "/")
	if len(args) != 3 {
		ctx.SetStatusCode(404)
		return
	}
	switch args[1] {
	case "likes":
		handlers.AddLikes(ctx)
		return
	case "new":
		handlers.Add(ctx)
		return
	default:
		id, err := strconv.ParseInt(args[1], 10, 0)
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		handlers.Update(ctx, int(id))
		return
	}
	ctx.SetStatusCode(404)
	return
}

//-----Очистка---------------

func clear() {
	//var on bool
	ms := &runtime.MemStats{}
	tick := time.Tick(time.Millisecond * 1000)
	for {
		select {
		case <-tick:
			runtime.ReadMemStats(ms)
			//sys := ms.TotalAlloc
			fmt.Println(ms.HeapAlloc)
			// if sys > 1800000000 && !on {
			// 	//fmt.Println("------GC Start-----")
			// 	on = true
			// 	go func(pon *bool) {
			// 		select {
			// 		case <-time.After(time.Millisecond * 2000):
			// 			*pon = false
			// 		}
			// 	}(&on)
			// 	runtime.GC()
			// }
		}
	}
}
