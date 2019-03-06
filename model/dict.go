package model

var DataCity = NewDict()
var DataCountry = NewDict()
var DataDomain = NewDict()
var DataInter = NewDict()
var DataStatus = map[string]byte{"свободны": 0, "всё сложно": 1, "заняты": 2}
var DataFname = NewDict()
var DataSname = make(map[string]uint32)
var RSname = make(map[uint32]string)

/*Dict - тип словарь*/
type Dict struct {
	main map[string]uint16
	rev  map[uint16]string
}

/*Get - возврат значения*/
func (dict Dict) Get(key string) (uint16, bool) {
	v, ok := dict.main[key]
	return v, ok
}

/*GetOrAdd - возврат или установка нового значения*/
func (dict Dict) GetOrAdd(key string) uint16 {
	v, ok := dict.main[key]
	if !ok {
		out := uint16(len(dict.main))
		dict.main[key] = out
		dict.rev[out] = key
		return out
	}
	return v
}

/*NewDict - новый словарь*/
func NewDict() Dict {
	m := make(map[string]uint16)
	m[""] = 0
	rev := make(map[uint16]string)
	return Dict{main: m, rev: rev}
}

/*GetRev - реверсное значение*/
func (dict Dict) GetRev(k uint16) string {
	return dict.rev[k]
}

/*GetMap - возвращаем главную карту для итераций*/
func (dict Dict) GetMap() map[string]uint16 {
	return dict.main
}
