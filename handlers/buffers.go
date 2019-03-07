package handlers

import (
	"Concurs/model"
	"Concurs/rgbtree"
	"bytes"
	"sync"
)

var bbuf = sync.Pool{
	New: func() interface{} {
		var bTs = make([]byte, 10000)
		return bytes.NewBuffer(bTs)
	},
}

var ubuff = sync.Pool{
	New: func() interface{} {
		return make([]*model.User, 0, 10000)
	},
}

var mapBuff = sync.Pool{
	New: func() interface{} {
		return make(map[uint64]int, 10000)
	},
}

var resBuff = sync.Pool{
	New: func() interface{} {
		return make([]res, 0, 10000)
	},
}

var buffTmps = sync.Pool{
	New: func() interface{} {
		return make([]tmpS, 0, 10000)
	},
}

var groupMap = sync.Pool{
	New: func() interface{} {
		nodes := make([]rgbtree.Node, 5000)
		return rgbtree.NewUTree(nodes)
	},
}

var uintB = sync.Pool{
	New: func() interface{} {

		return make([]uint64, 20000)
	},
}
var intB = sync.Pool{
	New: func() interface{} {

		return make([]int, 20000)
	},
}

var uint16Buff = sync.Pool{
	New: func() interface{} {

		return make([]uint16, 10000)
	},
}

var likesTempB = sync.Pool{
	New: func() interface{} {
		return make([]Temp, 200)
	},
}
