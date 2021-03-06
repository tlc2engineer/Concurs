package model

import (
	"math"
	"sort"
	"time"
)

var Loc, _ = time.LoadLocation("Europe/London")

var Now int64

/*Account - аккаунт*/
type Account struct {
	ID        int      `json:"id"`
	Email     string   `json:"email"`
	FName     string   `json:"fname"`
	SName     string   `json:"sname"`
	Phone     string   `json:"phone"`
	Sex       string   `json:"sex"`
	Birth     int64    `json:"birth"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	Joined    int64    `json:"joined"`
	Status    string   `json:"status"`
	Interests []string `json:"interests"`
	Premium   Premium  `json:"premium"`
	Likes     []Like   `json:"likes"`
	//mutex     *sync.RWMutex
}

/*FilterLike - фильтрация лайков*/
func (acc Account) FilterLike() []int64 {
	// acc.mutex.Lock()
	// defer acc.mutex.Unlock()
	mlike := make(map[int64]bool)
	for _, like := range acc.Likes {
		id := like.ID
		_, ok := mlike[id]
		if !ok {
			mlike[id] = true
		}
	}
	out := make([]int64, 0, len(mlike))
	for k := range mlike {
		out = append(out, k)
	}
	return out
}

/*IsPremium -  действует ли премиум-аккаунт*/
func (acc Account) IsPremium() bool {
	// acc.mutex.Lock()
	// defer acc.mutex.Unlock()
	premium := acc.Premium
	start := time.Unix(premium.Start, 0).In(Loc)
	finish := time.Unix(premium.Finish, 0).In(Loc)
	now := time.Unix(Now, 0).In(Loc)
	//acc.mutex.Unlock()
	return now.After(start) && now.Before(finish)
}

/*VStatus - цифровое представление статуса*/
func (acc Account) VStatus() int {
	// acc.mutex.Lock()
	// defer acc.mutex.Unlock()
	switch acc.Status {
	case "свободны":
		return 0
	case "всё сложно":
		return 1
	case "заняты":
		return 2
	}
	panic(acc.Status)
}

/*GetCommInt -  число общих интересов с другим аккаунтом*/
func (acc Account) GetCommInt(other Account) int {
	// acc.mutex.Lock()
	// defer acc.mutex.Unlock()
	ia := acc.Interests
	io := other.Interests
	comm := 0
	for _, ia1 := range ia {
		for _, io1 := range io {
			if ia1 == io1 {
				comm++
				break
			}
		}
	}
	return comm
}

/*GetBirth -  время рождения*/
func (acc Account) GetBirth() time.Time {
	// acc.mutex.Lock()
	// defer acc.mutex.Unlock()
	return time.Unix(acc.Birth, 0)
}

/*Suggest - общие предпочтения с другим аккаунтом*/
func (acc Account) Suggest(oth Account) float64 {
	ret := 0.0
	for _, like := range acc.Likes {
		for _, olike := range oth.Likes {
			if like.ID == olike.ID {
				ret += 1 / (math.Abs(like.Ts - olike.Ts))
			}
		}
	}
	return ret
}

/*NormLikes - нормализация*/
func NormLikes(inLikes []Like) []Like {
	// Сортировка
	sort.Slice(inLikes, func(i, j int) bool {
		return inLikes[i].ID < inLikes[j].ID
	})
	for i := range inLikes {
		inLikes[i].Num = 1
	}
	i := 0
	for i < len(inLikes)-1 {
		l1 := inLikes[i]
		l2 := inLikes[i+1]
		if l1.ID == l2.ID {
			// fmt.Println(i)
			// fmt.Println(inLikes)
			l1.Ts = (l1.Ts*float64(l1.Num) + l2.Ts*float64(l2.Num)) / (float64(l1.Num) + float64(l2.Num))
			l1.Num = l1.Num + l2.Num // суммирование лайков
			inLikes[i] = l1
			if i < len(inLikes)-2 {
				//fmt.Println("copy")
				copy(inLikes[i+1:], inLikes[i+2:])
			}
			inLikes = inLikes[:len(inLikes)-1] // уменьшение длины
			//fmt.Println(inLikes)
			continue
		}
		//fmt.Println(i, inLikes)
		i++
	}
	return inLikes
}

/*Premium - премиум*/
type Premium struct {
	Start  int64 `json:"start"`
	Finish int64 `json:"finish"`
}

/*NormAll -  нормализация лайков*/
func NormAll(acc []Account) []Account {
	for i := range acc {
		likes := acc[i].Likes
		acc[i].Likes = NormLikes(likes)
	}
	return acc
}
