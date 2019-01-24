package main

import (
	"Concurs/handlers"
	"Concurs/model"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

const opt = "options.txt"
const num = 130
const base = "D:/install/elim_accounts_261218/data/data/" //"D:/install/elim_accounts_261218/data/data/" //"D:/install/elim_accounts_261218/data/data/" //"/home/sergey/Загрузки/data/data/"

func main() {

	//----------------------------------------
	file, err := os.Open(base + opt)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	line, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}
	model.Now, err = strconv.ParseInt(string(line), 10, 0)

	if err != nil {
		panic(err)
	}
	// gen := len(os.Args) > 1 && os.Args[1] == "gen"
	// accounts := make([]model.Account, 0)
	bdata := make([]byte, 20000000)
	//users := make([]model.User, 0, num*10000) //num*10000
	for i := 1; i <= num; i++ {
		fmt.Println("Номер ", i)
		fname := fmt.Sprintf("%saccounts_%d.json", base, i)
		file, err := os.Open(fname)
		if err != nil {
			panic("Ошибка чтения")
		}
		n, err := file.Read(bdata)
		if err != nil {
			panic("Ошибка чтения 1")
		}
		bts := bdata[:n]
		err = model.AddUsers(bts)
		if err != nil {
			panic(err)
		}

	}

	/* Генерация sql
	if gen {
		fmt.Println("In SQL")
		err = util.CreateTables(accounts)
		if err != nil {
			panic(err)
		}
		return
	}
	*/
	model.SetUsers()
	//go clear()
	debug.SetGCPercent(50)
	router := fasthttprouter.New()
	router.GET("/accounts/*path", requestGet)
	router.POST("/accounts/*path", requestPost)
	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
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
	var on bool
	ms := &runtime.MemStats{}
	tick := time.Tick(time.Millisecond * 200)
	for {
		select {
		case <-tick:
			runtime.ReadMemStats(ms)
			sys := ms.Sys
			if sys > 1800000000 && !on {
				//fmt.Println("------GC Start-----")
				on = true
				go func(pon *bool) {
					select {
					case <-time.After(time.Millisecond * 2000):
						*pon = false
					}
				}(&on)
				runtime.GC()
			}
			// all := ms.HeapAlloc
			// fmt.Println("h all", all)
			// fmt.Println("idle", ms.HeapIdle)
			// fmt.Println("inuse", ms.HeapInuse)
			// fmt.Println(ms.Alloc)
			// fmt.Println(ms.Frees)
			// fmt.Println(ms.TotalAlloc)
			// fmt.Println(ms.Sys)
		}
	}
}
