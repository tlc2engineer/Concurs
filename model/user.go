package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

/*User - структура пользователя*/
type User struct {
	ID        uint32
	Email     string
	Domain    uint16
	FName     uint16
	SName     uint32
	Phone     string
	Code      uint16
	Sex       bool
	Birth     uint32
	Country   uint16
	City      uint16
	Joined    uint32
	Status    byte
	Interests []uint16
	Start     uint32
	Finish    uint32
}

//var convWg = &sync.WaitGroup{}

/*Conv - конвертация*/
func Conv(acc Account) User {

	//-------------------------------------
	country := DataCountry.GetOrAdd(acc.Country)
	countryMap.Add(uint32(country), uint32(acc.ID))
	//--------------------------------------
	city := DataCity.GetOrAdd(acc.City)
	cityMap.Add(uint32(city), uint32(acc.ID))

	//---------------------------------------
	fnameV := DataFname.GetOrAdd(acc.FName)
	fnameMap.Add(uint32(fnameV), uint32(acc.ID))
	//--------------------------------------
	snameV, ok := DataSname[acc.SName]
	if !ok {
		if acc.SName == "" {
			DataSname[acc.SName] = 0
			RSname[0] = ""
		} else {
			ln := uint32(len(DataSname) + 1)
			DataSname[acc.SName] = uint32(len(DataSname) + 1)
			RSname[ln] = acc.SName
			snameV = ln
		}
	}
	snameIndex.Add(snameV, uint32(acc.ID))
	//-------------------------------------------------
	interests := GetInterests(acc.Interests)
	for _, inter := range interests {
		intMap.Add(uint32(inter), uint32(acc.ID))
	}
	//---------------------------------------------------
	code := GetCode(acc.Phone)
	codeIndex.Add(uint32(code), uint32(acc.ID))
	//----------------------------------------------------
	if acc.Birth != 0 {
		date := time.Unix(int64(acc.Birth), 0).In(Loc)
		year := date.Year()
		bYearIndex.Add(uint32(year), uint32(acc.ID))
	}
	if acc.Joined != 0 {
		date := time.Unix(int64(acc.Joined), 0).In(Loc)
		year := date.Year()
		jYearIndex.Add(uint32(year), uint32(acc.ID))
	}
	//----------------------------------------------------
	domain := getDomain(acc.Email)
	domIndex.Add(uint32(domain), uint32(acc.ID))
	user := User{
		ID:        uint32(acc.ID),
		Email:     acc.Email,
		Domain:    domain,
		FName:     fnameV,
		SName:     snameV,
		Phone:     acc.Phone,
		Code:      code,
		Sex:       acc.Sex == "m",
		Birth:     uint32(acc.Birth),
		Country:   country,
		City:      city,
		Joined:    uint32(acc.Joined),
		Status:    DataStatus[acc.Status],
		Interests: interests,
		Start:     uint32(acc.Premium.Start),
		Finish:    uint32(acc.Premium.Finish),
	}
	// nuser := new(User) //UserPool.Get().(*User)
	// CopyUser(user, nuser)
	// mess := UserMess{
	// 	OldUser: nil,
	// 	NewUser: nuser,
	// }
	// UpdateChan <- &mess
	AddGIndex(user)
	AddRecIndex(user)

	//------Имена мужские и женские--------------
	addSexName(fnameV, acc.Sex == "m")
	//------Общие интересы-----------
	setCommonInt(interests)
	//-------------------------------
	return user
}

func getDomain(email string) uint16 {
	//fmt.Println("mail", email)
	return DataDomain.GetOrAdd(strings.Split(email, "@")[1])
}

func GetCode(phone string) uint16 {
	if phone == "" {
		return 0
	}
	// if !strings.Contains(phone, "(") || !strings.Contains(phone, ")") {
	// 	fmt.Println("phone", phone)
	// 	panic("!")
	// }
	part1 := strings.Split(phone, "(")[1]
	part2 := strings.Split(part1, ")")[0]
	res, _ := strconv.ParseInt(part2, 10, 0)
	return uint16(res)
}

/*GetInterests - Получить список интересов*/
func GetInterests(interests []string) []uint16 {
	out := make([]uint16, 0, len(interests))
	for i := range interests {
		in := DataInter.GetOrAdd(interests[i])
		out = append(out, in)
	}
	return out
}

/*GetCommInt -  число общих интересов с другим аккаунтом*/
func (acc User) GetCommInt(other User) int {
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

/*IsPremium -  действует ли премиум-аккаунт*/
func (acc User) IsPremium() bool {
	// start := time.Unix(int64(acc.Start), 0).In(Loc)
	// finish := time.Unix(int64(acc.Finish), 0).In(Loc)
	// now := time.Unix(Now, 0).In(Loc)
	//acc.mutex.Unlock()
	return acc.Start < uint32(Now) && acc.Finish > uint32(Now) //now.After(start) && now.Before(finish)
}

/*Suggest - общие предпочтения с другим аккаунтом*/
func (acc User) Suggest(oth *User) float64 {
	id := acc.ID
	oid := oth.ID
	likes := GetLikes(id) //likesMap[id]
	olikes := GetLikes(oid)
	ret := 0.0
	m := 0
	lenLikes := len(likes) / 8
	lenOlikes := len(olikes) / 8
	for i := 0; i < lenLikes; i++ {
		maddr := i * 8
		mid := uint32(likes[maddr]) | uint32(likes[maddr+1])<<8 | uint32(likes[maddr+2])<<16
		//like := LikeUnPack(likes[i*8 : i*8+8])
		for j := m; j < lenOlikes; j++ {
			addr := j * 8
			id := uint32(olikes[addr]) | uint32(olikes[addr+1])<<8 | uint32(olikes[addr+2])<<16
			if mid == id {
				mts := float64(binary.LittleEndian.Uint32(likes[maddr+3 : maddr+7]))
				ts := binary.LittleEndian.Uint32(olikes[addr+3 : addr+7])
				ret += 1 / (math.Abs(mts - float64(ts)))
				m = j + 1
				break
			}
			if id > mid {
				m = j
				break
			}
		}
	}
	return ret
}

/*GetSPVal - строковое значение по номеру*/
func GetSPVal(name string, val uint16) string {
	//fmt.Println(name)
	switch name {
	case "sex":
		if val == 0 {
			return "f"
		} else {
			return "m"
		}
	case "city":
		return DataCity.GetRev(val)
	case "country":
		return DataCountry.GetRev(val)
	case "status":
		for k, v := range DataStatus {
			if uint16(v) == val {
				return k
			}
		}
	case "interests":
		return DataInter.GetRev(val)
	}
	return ""
	//panic("Ошибка " + name)
}

func getRMap(m map[string]uint16, val uint16) string {
	for k, v := range m {
		if v == val {
			return k
		}
	}
	return ""
}

/*GetFname - имя пользователя строка*/
func (user User) GetFname() string {
	return DataFname.GetRev(user.FName)
}

func (user User) GetSname() string {
	return RSname[user.SName]
}

func (user *User) SetFname(fname string) {
	user.FName = DataFname.GetOrAdd(fname)
}

func (user *User) SetSname(sname string) {
	val, ok := DataSname[sname]
	if !ok {
		ln := uint32(len(DataSname) + 1)
		DataSname[sname] = ln
		RSname[ln] = sname
		user.SName = ln
	} else {
		user.SName = val
	}
}

// хак для перевода экранированных строк вида "\u1234\u5678" в нормальный юникод
func utf8Unescaped(b []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte('"')
	buf.Write(b)
	buf.WriteByte('"')
	var s string
	json.Unmarshal(buf.Bytes(), &s)
	return []byte(s)
}

/*getBYear - получение года рождения*/
func (acc User) getBYear() int {
	date := time.Unix(int64(acc.Birth), 0).In(Loc)
	return date.Year()

}

func (acc User) getJYear() int {
	date := time.Unix(int64(acc.Joined), 0).In(Loc)
	return date.Year()
}

/*CopyUser - копия User*/
func CopyUser(user User, nuser *User) {
	nuser.ID = user.ID
	nuser.Country = user.Country
	nuser.City = user.City
	nuser.FName = user.FName
	nuser.SName = user.SName
	nuser.Sex = user.Sex
	nuser.Joined = user.Joined
	nuser.Birth = user.Birth
	nuser.Email = user.Email
	nuser.Domain = user.Domain
	nuser.Status = user.Status
	nuser.Start = user.Start
	nuser.Finish = user.Finish
	ninter := make([]uint16, len(user.Interests))
	copy(ninter, user.Interests)
	nuser.Interests = ninter
}
