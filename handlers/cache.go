package handlers

import (
	"Concurs/model"
	"bytes"
	"sort"
	"sync"
)

var filterCache = make(map[int]fCache)
var suggCashe = make(map[uint64][]uint32)
var groupCache = make(map[string][]res)
var recCache = make(map[uint64][]*model.User)
var muf = &sync.Mutex{}
var mu = &sync.Mutex{}
var gmu = &sync.Mutex{}
var rmu = &sync.Mutex{}

/*ClearCashe - очистка cashe*/
func ClearCashe() {
	filterCache = make(map[int]fCache)
	suggCashe = make(map[uint64][]uint32)
	groupCache = make(map[string][]res)
	recCache = make(map[uint64][]*model.User)
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

func getRCache(qid uint64) ([]*model.User, bool) {
	rmu.Lock()
	defer rmu.Unlock()
	data, ok := recCache[qid]
	return data, ok
}

func setRCache(qid uint64, data []*model.User) {
	rmu.Lock()
	defer rmu.Unlock()
	recCache[qid] = data
}

func getCache(qid uint64) ([]uint32, bool) {
	mu.Lock()
	defer mu.Unlock()
	data, ok := suggCashe[qid]
	return data, ok
}

func setCache(qid uint64, data []uint32) {
	mu.Lock()
	defer mu.Unlock()
	suggCashe[qid] = data
}

func getGroupCache(key string) ([]res, bool) {
	gmu.Lock()
	defer gmu.Unlock()
	data, ok := groupCache[key]
	return data, ok
}

func setGroupCache(key string, data []res) {
	gmu.Lock()
	defer gmu.Unlock()
	groupCache[key] = data
}

type fCache struct {
	users  []*model.User
	fields []string
}

/*makeHasRec - hash код для suggest recommend*/
func makeHasRec(id uint32, country uint16, city uint16, limit int) uint64 {
	return uint64(id)<<32 | uint64(country)<<16 | uint64(city) | uint64(limit)<<56
}

func makeGroupHash(keys []string, params map[string]groupParam) string {
	dataHash := make([]byte, 200)
	bb := bytes.NewBuffer(dataHash)
	bb.Reset()
	for _, key := range keys {
		bb.WriteString(key)
	}
	kparams := make([]string, 0, len(params))
	for k := range params {
		if k != "query_id" {
			kparams = append(kparams, k)
		}
	}
	sort.Strings(kparams)
	for _, k := range kparams {
		bb.WriteString(k)
		v := params[k]
		switch v.tp {
		case 0:
			data := v.ival
			out := make([]byte, 4)
			out[0] = byte(data)
			out[1] = byte(data >> 8)
			out[2] = byte(data >> 16)
			out[3] = byte(data >> 24)
			bb.Write(out)
		case 1:
			bb.WriteString(v.sval)
		}
	}
	return bb.String()
}
