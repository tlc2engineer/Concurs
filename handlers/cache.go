package handlers

import (
	"Concurs/model"
	"sync"
)

var filterCache = make(map[int]fCache)
var accCashe = make(map[int][]uint32)
var muf = &sync.Mutex{}
var mu = &sync.Mutex{}
var gmu = &sync.Mutex{}
var rmu = &sync.Mutex{}
var groupCache = make(map[int][]res)
var recCache = make(map[int][]*model.User)

/*ClearCashe - очистка cashe*/
func ClearCashe() {
	filterCache = make(map[int]fCache)
	accCashe = make(map[int][]uint32)
	groupCache = make(map[int][]res)
	recCache = make(map[int][]*model.User)
}

func getFCache(qid int) (fCache, bool) {
	muf.Lock()
	defer muf.Unlock()
	data, ok := filterCache[qid]
	return data, ok
}

func setFCache(qid int, data fCache) {
	muf.Lock()
	defer muf.Unlock()
	filterCache[qid] = data
}

func getRCache(qid int) ([]*model.User, bool) {
	rmu.Lock()
	defer rmu.Unlock()
	data, ok := recCache[qid]
	return data, ok
}

func setRCache(qid int, data []*model.User) {
	rmu.Lock()
	defer rmu.Unlock()
	recCache[qid] = data
}

func getCache(qid int) ([]uint32, bool) {
	mu.Lock()
	defer mu.Unlock()
	data, ok := accCashe[qid]
	return data, ok
}

func setCache(qid int, data []uint32) {
	mu.Lock()
	defer mu.Unlock()
	accCashe[qid] = data
}

func getGroupCache(qid int) ([]res, bool) {
	gmu.Lock()
	defer gmu.Unlock()
	data, ok := groupCache[qid]
	return data, ok
}

func setGroupCache(qid int, data []res) {
	gmu.Lock()
	defer gmu.Unlock()
	groupCache[qid] = data
}

type fCache struct {
	users  []*model.User
	fields []string
}

/*makeHasRec - hash код для suggest recommend*/
func makeHasRec(id uint32, country uint16, city uint16) uint64 {
	return uint64(id)<<32 | uint64(country)<<16 | uint64(city)
}
