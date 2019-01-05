package main

import (
	"Concurs/handlers"
	"Concurs/model"
	"Concurs/util"

	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

const opt = "options.txt"
const num = 3
const base = "./data/"

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
	gen := len(os.Args) > 1 && os.Args[1] == "gen"
	accounts := make([]model.Account, 0)
	users := make([]model.User, 0, num*10000)
	for i := 1; i <= num; i++ {
		fmt.Println("Номер ", i)
		fname := fmt.Sprintf("%saccounts_%d.json", base, i)
		acc, err := getData(fname)
		if err != nil {
			panic(err)
		}

		acc = model.NormAll(acc)
		if gen {
			accounts = append(accounts, acc...)
		} else {
			// Добавление данных об аккаунте
			wb := model.DB.NewWriteBatch()
			//}
			for i := range acc {
				user := model.Conv(acc[i])
				if false {
					fmt.Println(user.City)
				}
				if model.Budger {
					err = wb.Set([]byte("l"+string(user.ID)), model.PackLSlice(acc[i].Likes), 0)
					if err != nil {
						fmt.Println(err)
					}
				} else {
					model.SetLikes(user.ID, model.PackLSlice(acc[i].Likes))
				}
				// добавление данных в карту кто-кого
				for _, like := range acc[i].Likes {
					model.AddWho(user.ID, like)
				}
				users = append(users, user)
			}
			err = wb.Flush()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	// Генерация sql
	if gen {
		fmt.Println("In SQL")
		err = util.CreateTables(accounts)
		if err != nil {
			panic(err)
		}
		return
	}
	model.SetUsers(users)
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
