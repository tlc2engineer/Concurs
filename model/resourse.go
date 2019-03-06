package model

import (
	"fmt"
)

/*Res32Producer -  поставщик []uint32*/
type Res32Producer struct {
	buff []uint32
	m    int
}

/*New32res - новый поставщик ресурсов*/
func New32res(l int) *Res32Producer {
	buff := make([]uint32, l)
	return &Res32Producer{buff, 0}
}

/*Reset - сброс данных*/
func (prod *Res32Producer) Reset() {
	prod.m = 0
}

/*Len - длина занятых данных*/
func (prod *Res32Producer) Len() int {
	return prod.m
}

/*Get - получение данных*/
func (prod *Res32Producer) Get(l int) ([]uint32, error) {
	if (prod.m + l) > len(prod.buff) {

		return nil, fmt.Errorf("Нет ресурсов")
	}
	ret := prod.buff[prod.m : prod.m+l]
	prod.m = prod.m + l
	return ret, nil
}

/*Available -сколько свободно*/
func (prod *Res32Producer) Available() int {
	return len(prod.buff) - prod.m
}
