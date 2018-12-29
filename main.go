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
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

const opt = "./data/options.txt"
const filename = "./data/accounts_1.json"
const filename_2 = "./data/accounts_2.json"
const filename_3 = "./data/accounts_3.json"

func main() {
	file, err := os.Open(opt)
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
	//router := mux.NewRouter()
	accounts := make([]model.Account, 0)
	acc, err := getData(filename)
	if err != nil {
		panic(err)
	}
	accounts = append(accounts, acc...)
	acc, err = getData(filename_2)
	if err != nil {
		panic(err)
	}
	accounts = append(accounts, acc...)
	acc, err = getData(filename_3)
	if err != nil {
		panic(err)
	}
	accounts = append(accounts, acc...)
	fmt.Println(len(accounts))
	if len(os.Args) > 1 && os.Args[1] == "gen" {
		err = util.CreateTables(accounts)
		if err != nil {
			panic(err)
		}
		return
	}
	accounts = model.NormAll(accounts) // нормализация
	model.SetAccounts(accounts)
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

/*Server - сервер*/
type Server struct {
}

/*ServeHTTP - роутер*/
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	meth := r.Method
	if meth == "GET" && strings.HasPrefix(path, "/accounts/filter/") {
		if len(strings.Split(path, "/")) > 4 {
			w.WriteHeader(404)
			return
		}
		//handlers.Filter(w, r)
	}
	if meth == "GET" && strings.HasPrefix(path, "/accounts/group/") {
		if len(strings.Split(path, "/")) > 4 {
			w.WriteHeader(404)
			return
		}
		//handlers.Group(w, r)
	}
	if meth == "GET" && strings.HasPrefix(path, "/accounts/") {
		//args := strings.Split(path, "/")
		// if args[3] == "recommend" {
		// 	id, err := strconv.Atoi(args[2])
		// 	if err != nil {
		// 		w.WriteHeader(400)
		// 		return
		// 	}
		// 	//handlers.Recommend(w, r, id)
		// }
	}
	if meth == "GET" && strings.HasPrefix(path, "/accounts/") {
		// args := strings.Split(path, "/")
		// if args[3] == "suggest" {
		// 	id, err := strconv.Atoi(args[2])
		// 	if err != nil {
		// 		w.WriteHeader(400)
		// 		return
		// 	}
		// 	//handlers.Suggest(w, r, id)
		// }
	}
	if meth == "POST" && strings.HasPrefix(path, "/accounts/new/") {
		//handlers.Add(w, r)
	}
	if meth == "POST" && strings.HasPrefix(path, "/accounts/") && !strings.Contains(path, "new") && !strings.Contains(path, "likes") {
		// args := strings.Split(path, "/")
		// id, err := strconv.Atoi(args[2])
		// if err != nil {
		// 	w.WriteHeader(404)
		// 	return
		// }
		//handlers.Update(w, r, id)
	}
	if meth == "POST" && strings.HasPrefix(path, "/accounts/likes/") {
		//handlers.AddLikes(w, r)
	}
	w.WriteHeader(404)
	return
}

func requestGet(ctx *fasthttp.RequestCtx) {
	//fmt.Fprintf(ctx, "Hello, world! Requested path is %q", ctx.Path())
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
