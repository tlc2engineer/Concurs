package util

import (
	"Concurs/model"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

const connStr = "user=postgres dbname=concurs sslmode=disable password=m171079 host=localhost"

/*CreateTables - заполнение sql*/
func CreateTables(data []model.Account) error {
	fmt.Println("Создание таблиц", len(data))
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}
	likes := make([]tlike, 0)
	count := 0
	icount := 0
	imap := make(map[string]int)
	values := []interface{}{}
	insertAcc := `INSERT INTO public.accounts(
		id, email, fname, sname, phone, sex, country, city, joined, interests, start, finish, status, birth)
		VALUES `
	for i, acc := range data {
		for _, like := range acc.Likes {
			likes = append(likes, tlike{int64(acc.ID), like.ID, like.Ts})
		}
		interests := acc.Interests
		islice := make([]int, 0)
		for _, inter := range interests {
			num, ok := imap[inter]
			if !ok {
				imap[inter] = icount
				islice = append(islice, icount)
				icount++
			} else {
				islice = append(islice, num)
			}
		}
		status := 0
		if acc.Status == "всё сложно" {
			status = 1
		}
		if acc.Status == "заняты" {
			status = 2
		}
		values = append(values, acc.ID, acc.Email, acc.FName, acc.SName, acc.Phone, acc.Sex, acc.Country, acc.City, time.Unix(acc.Joined, 0), intarray(islice),
			time.Unix(acc.Premium.Start, 0), time.Unix(acc.Premium.Finish, 0), status, time.Unix(acc.Birth, 0))
		numFields := 14 // the number of fields you are inserting
		n := count * numFields
		insertAcc += `(`
		for j := 0; j < numFields; j++ {
			insertAcc += `$` + strconv.Itoa(n+j+1) + `,`
		}
		insertAcc = insertAcc[:len(insertAcc)-1] + `),`
		count++
		if count == 1000 || i == len(data)-1 {
			insertAcc = insertAcc[:len(insertAcc)-1]
			//fmt.Println(insertAcc)
			res, err := db.Exec(insertAcc, values...)
			if err != nil {
				fmt.Println(insertAcc)
				panic(err)
			}
			fmt.Println("Новые 1000", res)
			count = 0
			values = []interface{}{}
			insertAcc = `INSERT INTO public.accounts(
			id, email, fname, sname, phone, sex, country, city, joined, interests, start, finish, status, birth)
			VALUES `
		}
	}
	// Вывод интересов
	query := `INSERT INTO public.interests(
	id, interes)
	VALUES ($1, $2);`
	for k, v := range imap {
		res, err := db.Exec(query, v, k)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)

	}
	// Вывод лайков
	count = 0
	values = []interface{}{}
	query = `INSERT INTO public.likes(
		id, pid, tm)
		VALUES  `
	for i, like := range likes {
		values = append(values, like.id, like.tid, like.ts)
		numFields := 3 // the number of fields you are inserting
		n := count * numFields
		query += `(`
		for j := 0; j < numFields; j++ {
			query += `$` + strconv.Itoa(n+j+1) + `,`
		}
		query = query[:len(query)-1] + `),`
		count++
		if count == 1000 || i == len(likes)-1 {
			query = query[:len(query)-1] // убираем запятую
			res, err := db.Exec(query, values...)
			if err != nil {
				fmt.Println(query)
				panic(err)
			}
			fmt.Println("Новые 1000 like", res)
			count = 0
			values = []interface{}{}
			query = `INSERT INTO public.likes(
				id, pid, tm)
				VALUES  `
		}
	}

	return nil
}

type intarray []int

func (a intarray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	ret := "{"
	for _, i := range a {
		ret += fmt.Sprintf("%d,", i)
	}
	ret = ret[:len(ret)-1]
	ret += "}"
	return ret, nil
}

type tlike struct {
	id  int64
	tid int64
	ts  float64
}
