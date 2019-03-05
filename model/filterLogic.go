package model

type indexLogic interface {
	nextIt() int               // следующая итерация
	get() int                  // текущее значение
	nextCheck(val uint32) bool // следующая проверка
	len() int                  // длина
}

type simpleIndexLogic struct {
	m    int // маркер
	data []uint32
}

/*newSLogic - новый итератор*/
func newSLogic(data []uint32) *simpleIndexLogic {
	return &simpleIndexLogic{len(data) - 1, data}
}
func (sil *simpleIndexLogic) len() int {
	return len(sil.data)
}

func (sil *simpleIndexLogic) get() int {
	if sil.m < 0 {
		return -1
	}
	return int(sil.data[sil.m])
}

func (sil *simpleIndexLogic) nextIt() int {
	if sil.m < 0 {
		return -1
	}
	out := sil.get()
	sil.m = sil.m - 1
	return out
}

func (sil *simpleIndexLogic) nextCheck(val uint32) bool {
	if sil.m < 0 {
		return false
	}
	var i int
	for i = sil.m; i >= 0; i-- {
		if sil.data[i] == val {
			sil.m = i - 1
			return true
		}
		if sil.data[i] < val {
			sil.m = i
			return false
		}

	}
	sil.m = -1 // все значения прочитаны
	return false
}

type cmplIndLog struct {
	data []*simpleIndexLogic
}

func newCmplLog(data []*simpleIndexLogic) *cmplIndLog {
	return &cmplIndLog{data}
}

/*len - суммарная длина индексов*/
func (cil *cmplIndLog) len() int {
	var sum int
	for i := 0; i < len(cil.data); i++ {
		sum += cil.data[i].len()
	}
	return sum
}

/*get - получить значение*/
func (cil *cmplIndLog) get() int {
	var max = -1
	for i := 0; i < len(cil.data); i++ {
		val := cil.data[i].get()
		if val > max {
			max = val
		}
	}
	return max
}

func (cil *cmplIndLog) nextCheck(val uint32) bool {
	for i := 0; i < len(cil.data); i++ {
		if cil.data[i].nextCheck(val) {
			return true
		}
	}
	return false
}

func (cil *cmplIndLog) nextIt() int {
	max := cil.get()
	for i := 0; i < len(cil.data); i++ {
		if cil.data[i].get() == max {
			cil.data[i].nextIt()
		}
	}
	return max
}
