package main

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fastjson"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

var (
	inverseUint32Comparator = func(a, b interface{}) int {
		return -utils.UInt32Comparator(a, b)
	}
	inverseFloat32Comparator = func(a, b interface{}) int {
		return -utils.Float32Comparator(a, b)
	}
	accountIndex   = treemap.NewWith(inverseUint32Comparator)
	countryIndex   = NewSafeTreemapIndex()
	cityIndex      = NewSafeIndex()
	birthYearIndex = NewSafeIndex()
	fnameIndex     = NewSafeIndex()
	snameIndex     = NewSafeIndex()
	sexIndex       = NewSafeIndex()
	interestsIndex = NewSafeIndex()
	likeeIndex     = NewSafeIndex() // who liked this user
	emailIndex     = NewSafeIndex()
	phoneIndex     = NewSafeIndex()

	inversedTreemapPool = &sync.Pool{
		New: func() interface{} {
			return treemap.NewWith(inverseUint32Comparator)
		},
	}
)

type Account struct {
	ID      int            `json:"id"`
	Email   string         `json:"email"`
	Fname   string         `json:"fname"`
	Sname   string         `json:"sname"`
	Phone   string         `json:"phone"`
	Sex     string         `json:"sex"`
	Birth   int            `json:"birth"`
	Country string         `json:"country"`
	City    string         `json:"city"`
	Joined  int            `json:"joined"`
	Status  string         `json:"status"`
	Premium map[string]int `json:"premium"`

	interestsMap map[string]struct{}
	emailDomain  string
	phoneCode    int
	birthYear    int
	joinedYear   int
	likes        map[int]LikesList

	sync.Mutex
}

func NewStringDictionary() *StringDictionary {
	return &StringDictionary{
		v: make(map[string]uint8, 0),
		k: make(map[uint8]string, 0),
	}
}

type StringDictionary struct {
	v map[string]uint8
	k map[uint8]string
	sync.Mutex
}

func (d *StringDictionary) Add(value string) uint8 {
	d.Lock()
	defer d.Unlock()

	if id, ok := d.v[value]; ok {
		return id
	}

	id := uint8(len(d.v)) + 1
	d.v[value] = id
	d.k[id] = value

	return id
}

func (d *StringDictionary) Get(id uint8) string {
	d.Lock()
	defer d.Unlock()

	if value, ok := d.k[id]; ok {
		return value
	}
	return ""
}

func (d *StringDictionary) GetId(value string) uint8 {
	d.Lock()
	defer d.Unlock()

	if id, ok := d.v[value]; ok {
		return id
	}
	return 0
}

var statusDict = &StringDictionary{
	v: map[string]uint8{
		"свободны":   1,
		"все сложно": 2,
		"заняты":     3,
	},
	k: map[uint8]string{
		1: "свободны",
		2: "все сложно",
		3: "заняты",
	},
}
var sexDict = &StringDictionary{
	v: map[string]uint8{
		"m": 1,
		"f": 2,
	},
	k: map[uint8]string{
		1: "m",
		2: "f",
	},
}
var interestsDict = NewStringDictionary()
var emailDomainsDict = NewStringDictionary()
var phoneCodesDict = NewStringDictionary()
var countriesDict = NewStringDictionary()
var citiesDict = NewStringDictionary()
var fnamesDict = NewStringDictionary()
var snamesDict = NewStringDictionary()

var likesMap map[*AccountUpdated]map[uint32]LikesList

//interestsMap map[byte]struct{}

//13000000 // MaxInt16
//** Interests 90
//*** Domain 13
//*** Fnames 108
//*** City 607
//*** Country 70
//*** Email prefixes 240
//*** Sname prefixes 64
//*** phone codes 100
type AccountUpdated struct {
	ID           uint32
	Status       byte
	Fname        uint8
	Sname        uint8
	Sex          byte
	Email        string
	Phone        string
	Birth        int32
	BirthYear    uint16
	Country      byte
	City         byte
	Joined       uint32
	JoinedYear   uint16
	EmailDomain  byte
	PhoneCode    byte
	Premium      map[string]uint32
	InterestsMap map[uint8]struct{}

	sync.Mutex
}

func (acc *Account) AppendLike(likeeId int, likeTs int) {
	acc.Lock()
	defer acc.Unlock()

	acc.likes[likeeId] = append(acc.likes[likeeId], likeTs)
}

func (acc Account) hasActivePremium(now int64) bool {
	return acc.Premium["start"] <= int(now) && acc.Premium["finish"] > int(now)
}

func (acc *Account) Update(data []byte) {
	p := pp.Get()

	jsonValue, _ := p.ParseBytes(data)

	if jsonValue.Exists("interests") {
		// delete old value from indexes
		for v := range acc.interestsMap {
			interestsIndex.Get(v).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.interestsMap = make(map[string]struct{})
		for _, v := range jsonValue.GetArray("interests") {
			interest := string(v.GetStringBytes())
			acc.interestsMap[interest] = struct{}{}

			if !interestsIndex.Exists(interest) {
				interestsIndex.Update(interest, treemap.NewWith(inverseUint32Comparator))
			}
			interestsIndex.Get(interest).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("email") {
		// delete old value from indexes
		emailIndex.Delete(acc.Email)
		acc.emailDomain = ""

		// set new value
		acc.Email = string(jsonValue.GetStringBytes("email"))
		if len(acc.Email) > 0 {
			components := strings.Split(acc.Email, "@")
			if len(components) > 1 {
				acc.emailDomain = components[1]
			}
			emailIndex.Update(acc.Email, struct{}{})
		}
	}

	if jsonValue.Exists("status") {
		// set new value
		acc.Status = string(jsonValue.GetStringBytes("status"))
	}

	if jsonValue.Exists("phone") {
		// delete old value from indexes
		phoneIndex.Delete(acc.Phone)
		acc.phoneCode = 0

		// set new value
		acc.Phone = string(jsonValue.GetStringBytes("phone"))
		if len(acc.Phone) > 0 {
			phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
			if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
				acc.phoneCode = phoneCode
			}
			phoneIndex.Update(acc.Phone, struct{}{})
		}
	}

	if jsonValue.Exists("birth") {
		// delete old value from indexes
		if birthYearIndex.Exists(acc.birthYear) {
			birthYearIndex.Get(acc.birthYear).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Birth = jsonValue.GetInt("birth")
		acc.birthYear = time.Unix(int64(acc.Birth), 0).In(time.UTC).Year()

		if acc.birthYear > 0 {
			if !birthYearIndex.Exists(acc.birthYear) {
				birthYearIndex.Update(acc.birthYear, treemap.NewWith(inverseUint32Comparator))
			}
			birthYearIndex.Get(acc.birthYear).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("joined") {
		// set new value
		acc.Joined = jsonValue.GetInt("joined")
		acc.joinedYear = time.Unix(int64(acc.Joined), 0).In(time.UTC).Year()
	}

	//todo: likes?

	if jsonValue.Exists("country") {
		// delete old value from indexes
		if countryIndex.Exists(acc.Country) {
			countryIndex.Get(acc.Country).Remove(acc.ID)
		}

		// set new value
		acc.Country = string(jsonValue.GetStringBytes("country"))
		if len(acc.Country) > 0 {
			if !countryIndex.Exists(acc.Country) {
				countryIndex.Update(acc.Country, treemap.NewWith(inverseUint32Comparator))
			}
			countryIndex.Get(acc.Country).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("city") {
		// delete old value from indexes
		if cityIndex.Exists(acc.City) {
			cityIndex.Get(acc.City).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.City = string(jsonValue.GetStringBytes("city"))
		if len(acc.City) > 0 {
			if !cityIndex.Exists(acc.City) {
				cityIndex.Update(acc.City, treemap.NewWith(inverseUint32Comparator))
			}
			cityIndex.Get(acc.City).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("fname") {
		// delete old value from indexes
		if fnameIndex.Exists(acc.Fname) {
			fnameIndex.Get(acc.Fname).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Fname = string(jsonValue.GetStringBytes("fname"))
		if len(acc.Fname) > 0 {
			if !fnameIndex.Exists(acc.Fname) {
				fnameIndex.Update(acc.Fname, treemap.NewWith(inverseUint32Comparator))
			}
			fnameIndex.Get(acc.Fname).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("sname") {
		// delete old value from indexes
		if snameIndex.Exists(acc.Sname) {
			snameIndex.Get(acc.Sname).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Sname = string(jsonValue.GetStringBytes("sname"))
		if len(acc.Sname) > 0 {
			if !snameIndex.Exists(acc.Sname) {
				snameIndex.Update(acc.Sname, treemap.NewWith(inverseUint32Comparator))
			}
			snameIndex.Get(acc.Sname).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if jsonValue.Exists("sex") {
		// delete old value from indexes
		if sexIndex.Exists(acc.Sex) {
			sexIndex.Get(acc.Sex).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Sex = string(jsonValue.GetStringBytes("sex"))
		if !sexIndex.Exists(acc.Sex) {
			sexIndex.Update(acc.Sex, treemap.NewWith(inverseUint32Comparator))
		}
		sexIndex.Get(acc.Sex).(*treemap.Map).Put(acc.ID, acc)
	}

	if jsonValue.Exists("premium") {
		premiumObj := jsonValue.GetObject("premium")
		if premiumObj != nil && premiumObj.Len() > 0 {
			acc.Premium = make(map[string]int, 0)

			acc.Premium["start"] = premiumObj.Get("start").GetInt()
			acc.Premium["finish"] = premiumObj.Get("finish").GetInt()
		}
	}
}

// Update user
func (acc *Account) UpdateOld(changedData map[string]interface{}) {
	if newValue, ok := changedData["interests"]; ok {
		// delete old value from indexes
		for v := range acc.interestsMap {
			interestsIndex.Get(v).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.interestsMap = make(map[string]struct{})
		for _, v := range newValue.([]interface{}) {
			interest := v.(string)
			acc.interestsMap[interest] = struct{}{}
			if !interestsIndex.Exists(interest) {
				interestsIndex.Update(interest, treemap.NewWith(inverseUint32Comparator))
			}
			interestsIndex.Get(interest).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if newValue, ok := changedData["email"]; ok {
		// delete old value from indexes
		emailIndex.Delete(acc.Email)
		acc.emailDomain = ""

		// set new value
		acc.Email = newValue.(string)
		components := strings.Split(acc.Email, "@")
		if len(components) > 1 {
			acc.emailDomain = components[1]
		}
		emailIndex.Update(acc.Email, struct{}{})
	}

	if newValue, ok := changedData["status"]; ok {
		// set new value
		acc.Status = newValue.(string)
	}

	if newValue, ok := changedData["phone"]; ok {
		// delete old value from indexes
		phoneIndex.Delete(acc.Phone)
		acc.phoneCode = 0

		// set new value
		acc.Phone = newValue.(string)
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
			acc.phoneCode = phoneCode
		}
		phoneIndex.Update(acc.Phone, struct{}{})
	}

	if newValue, ok := changedData["birth"]; ok {
		// delete old value from indexes
		if birthYearIndex.Exists(acc.birthYear) {
			birthYearIndex.Get(acc.birthYear).(*treemap.Map).Remove(acc.ID)
		}

		acc.Birth = newValue.(int)
		// set new value
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(int64(acc.Birth), 0)
		acc.birthYear = tm.In(loc).Year()

		if acc.birthYear > 0 {
			if !birthYearIndex.Exists(acc.birthYear) {
				birthYearIndex.Update(acc.birthYear, treemap.NewWith(inverseUint32Comparator))
			}
			birthYearIndex.Get(acc.birthYear).(*treemap.Map).Put(acc.ID, acc)
		}
	}

	if newValue, ok := changedData["joined"]; ok {
		// set new value
		acc.Joined = newValue.(int)
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(int64(acc.Joined), 0)
		acc.joinedYear = tm.In(loc).Year()
	}

	//FIXME:
	//if newValue, ok := changedData["likes"]; ok {
	//}

	if newValue, ok := changedData["country"]; ok {
		// delete old value from indexes
		if countryIndex.Exists(acc.Country) {
			countryIndex.Get(acc.Country).Remove(acc.ID)
		}

		// set new value
		acc.Country = newValue.(string)
		if !countryIndex.Exists(acc.Country) {
			countryIndex.Update(acc.Country, treemap.NewWith(inverseUint32Comparator))
		}
		countryIndex.Get(acc.Country).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["city"]; ok {
		// delete old value from indexes
		if cityIndex.Exists(acc.City) {
			cityIndex.Get(acc.City).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.City = newValue.(string)
		if !cityIndex.Exists(acc.City) {
			cityIndex.Update(acc.City, treemap.NewWith(inverseUint32Comparator))
		}
		cityIndex.Get(acc.City).(*treemap.Map).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["fname"]; ok {
		// delete old value from indexes
		if fnameIndex.Exists(acc.Fname) {
			fnameIndex.Get(acc.Fname).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Fname = newValue.(string)
		if !fnameIndex.Exists(acc.Fname) {
			fnameIndex.Update(acc.Fname, treemap.NewWith(inverseUint32Comparator))
		}
		fnameIndex.Get(acc.Fname).(*treemap.Map).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["sname"]; ok {
		// delete old value from indexes
		if snameIndex.Exists(acc.Sname) {
			snameIndex.Get(acc.Sname).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Sname = newValue.(string)
		if !snameIndex.Exists(acc.Sname) {
			snameIndex.Update(acc.Sname, treemap.NewWith(inverseUint32Comparator))
		}
		snameIndex.Get(acc.Sname).(*treemap.Map).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["sex"]; ok {
		// delete old value from indexes
		if sexIndex.Exists(acc.Sex) {
			sexIndex.Get(acc.Sex).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Sex = newValue.(string)
		if !sexIndex.Exists(acc.Sex) {
			sexIndex.Update(acc.Sex, treemap.NewWith(inverseUint32Comparator))
		}
		sexIndex.Get(acc.Sex).(*treemap.Map).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["premium"]; ok {
		acc.Premium = make(map[string]int, 0)
		// set new value
		data := newValue.(map[string]interface{})
		acc.Premium["start"] = int(data["start"].(float64))
		acc.Premium["finish"] = int(data["finish"].(float64))
	}
}

type LikesList []int

func (list LikesList) getTimestamp() int {
	var ts int

	if len(list) > 1 {
		var total = 0
		for _, value := range list {
			total += value
		}
		ts = total / int(len(list))
	} else {
		ts = list[0]
	}

	return ts
}

func NewAccountFromByte(data []byte) {
	p := pp.Get()

	jsonData, _ := p.ParseBytes(data)

	NewAccountFromJson(jsonData)

	pp.Put(p)
}

func NewAccountFromJson(jsonValue *fastjson.Value) {
	acc := &AccountUpdated{}
	acc.ID = uint32(jsonValue.GetUint("id"))

	acc.InterestsMap = make(map[uint8]struct{}, 0)
	for _, v := range jsonValue.GetArray("interests") {
		interest := string(v.GetStringBytes())
		interestId := interestsDict.Add(interest)
		acc.InterestsMap[interestId] = struct{}{}

		if !interestsIndex.Exists(interestId) {
			interestsIndex.Update(interestId, treemap.NewWith(inverseUint32Comparator))
		}
		interestsIndex.Get(interestId).(*treemap.Map).Put(acc.ID, acc)
	}

	acc.Email = string(jsonValue.GetStringBytes("email"))
	if len(acc.Email) > 0 {
		components := strings.Split(acc.Email, "@")
		if len(components) > 1 {
			acc.EmailDomain = emailDomainsDict.Add(components[1])
		}
		emailIndex.Update(acc.Email, struct{}{})
	}

	acc.Status = statusDict.GetId(string(jsonValue.GetStringBytes("status")))

	acc.Phone = string(jsonValue.GetStringBytes("phone"))
	if len(acc.Phone) > 0 {
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		acc.PhoneCode = phoneCodesDict.Add(phoneCodeStr)
		phoneIndex.Update(acc.Phone, struct{}{})
	}

	acc.Birth = int32(jsonValue.GetInt("birth"))
	acc.BirthYear = uint16(time.Unix(int64(acc.Birth), 0).In(time.UTC).Year())

	if acc.BirthYear > 0 {
		if !birthYearIndex.Exists(acc.BirthYear) {
			birthYearIndex.Update(acc.BirthYear, treemap.NewWith(inverseUint32Comparator))
		}
		birthYearIndex.Get(acc.BirthYear).(*treemap.Map).Put(acc.ID, acc)
	}

	acc.Joined = uint32(jsonValue.GetUint("joined"))
	acc.JoinedYear = uint16(time.Unix(int64(acc.Joined), 0).In(time.UTC).Year())

	//acc.likes = make(map[int]LikesList, 0)
	//for _, v := range jsonValue.GetArray("likes") {
	//	likeId := v.GetInt("id")
	//	ts := v.GetInt("ts")
	//
	//	// TODO: ignore 0 in like id / ts
	//	if likeId == 0 || ts == 0 {
	//		continue
	//	}
	//
	//	acc.AppendLike(likeId, ts)
	//
	//	if !likeeIndex.Exists(likeId) {
	//		likeeIndex.Update(likeId, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	likeeIndex.Get(likeId).(*treemap.Map).Put(acc.ID, acc)
	//}

	acc.Country = countriesDict.Add(string(jsonValue.GetStringBytes("country")))
	if acc.Country > 0 {
		if !countryIndex.Exists(acc.Country) {
			countryIndex.Update(acc.Country, treemap.NewWith(inverseUint32Comparator))
		}
		countryIndex.Get(acc.Country).Put(acc.ID, acc)
	}

	acc.City = citiesDict.Add(string(jsonValue.GetStringBytes("city")))
	if acc.City > 0 {
		if !cityIndex.Exists(acc.City) {
			cityIndex.Update(acc.City, treemap.NewWith(inverseUint32Comparator))
		}
		cityIndex.Get(acc.City).(*treemap.Map).Put(acc.ID, acc)
	}

	acc.Fname = fnamesDict.Add(string(jsonValue.GetStringBytes("fname")))
	if acc.Fname > 0 {
		if !fnameIndex.Exists(acc.Fname) {
			fnameIndex.Update(acc.Fname, treemap.NewWith(inverseUint32Comparator))
		}
		fnameIndex.Get(acc.Fname).(*treemap.Map).Put(acc.ID, acc)
	}

	acc.Sname = snamesDict.Add(string(jsonValue.GetStringBytes("sname")))
	if acc.Sname > 0 {
		if !snameIndex.Exists(acc.Sname) {
			snameIndex.Update(acc.Sname, treemap.NewWith(inverseUint32Comparator))
		}
		snameIndex.Get(acc.Sname).(*treemap.Map).Put(acc.ID, acc)
	}

	acc.Sex = sexDict.GetId(string(jsonValue.GetStringBytes("sex")))
	if !sexIndex.Exists(acc.Sex) {
		sexIndex.Update(acc.Sex, treemap.NewWith(inverseUint32Comparator))
	}
	sexIndex.Get(acc.Sex).(*treemap.Map).Put(acc.ID, acc)

	premiumObj := jsonValue.GetObject("premium")
	if premiumObj != nil && premiumObj.Len() > 0 {
		acc.Premium = make(map[string]uint32, 0)

		acc.Premium["start"] = uint32(premiumObj.Get("start").GetUint())
		acc.Premium["finish"] = uint32(premiumObj.Get("finish").GetUint())
	}

	accountIndex.Put(acc.ID, acc) // 394
	//accountMapIndex[acc.ID] = acc // 341
}

func NewAccount(acc Account) {
	//if len(acc.Interests) > 0 {
	//	acc.interestsMap = make(map[string]struct{})
	//	for _, interest := range acc.Interests {
	//		acc.interestsMap[interest] = struct{}{}
	//		if !interestsIndex.Exists(interest) {
	//			interestsIndex.Update(interest, treemap.NewWith(inverseUint32Comparator))
	//		}
	//		interestsIndex.Get(interest).(*treemap.Map).Put(acc.ID, &acc)
	//	}
	//	acc.Interests = nil
	//}
	//
	//if acc.Email != "" {
	//	components := strings.Split(acc.Email, "@")
	//	if len(components) > 1 {
	//		acc.emailDomain = components[1]
	//	}
	//	emailIndex.Update(acc.Email, 1)
	//}
	//
	//if acc.Phone != "" {
	//	phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
	//	if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
	//		acc.phoneCode = phoneCode
	//	}
	//	phoneIndex.Update(acc.Phone, struct{}{})
	//}
	//
	//if acc.Birth != 0 {
	//	loc, _ := time.LoadLocation("UTC")
	//	tm := time.Unix(int64(acc.Birth), 0)
	//	acc.birthYear = tm.In(loc).Year()
	//}
	//
	//if acc.Joined != 0 {
	//	loc, _ := time.LoadLocation("UTC")
	//	tm := time.Unix(int64(acc.Joined), 0)
	//	acc.joinedYear = tm.In(loc).Year()
	//}
	//
	//if len(acc.TempLikes) > 0 {
	//	acc.likes = make(map[int]LikesList, 0)
	//	//gjson.ParseBytes(acc.TempLikes).ForEach(func(key, value gjson.Result) bool {
	//	//	like := value.Map()
	//	//	likeId := int(like["id"].Int())
	//	//
	//	//	acc.AppendLike(likeId, int(like["ts"].Int()))
	//	//
	//	//	if !likeeIndex.Exists(likeId) {
	//	//		likeeIndex.Update(likeId, treemap.NewWith(inverseUint32Comparator))
	//	//	}
	//	//	likeeIndex.Get(likeId).(*treemap.Map).Put(acc.ID, &acc)
	//	//	return true
	//	//})
	//	acc.TempLikes = nil
	//}

	if acc.Country != "" {
		vmap := inversedTreemapPool.Get()
		if !countryIndex.Exists(acc.Country) {
			countryIndex.Update(acc.Country, vmap.(*treemap.Map))
		}
		countryIndex.Get(acc.Country).Put(acc.ID, &acc)
		inversedTreemapPool.Put(vmap)
	}
	//if acc.City != "" {
	//	if !cityIndex.Exists(acc.City) {
	//		cityIndex.Update(acc.City, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	cityIndex.Get(acc.City).(*treemap.Map).Put(acc.ID, &acc)
	//}
	//if acc.birthYear > 0 {
	//	if !birthYearIndex.Exists(acc.birthYear) {
	//		birthYearIndex.Update(acc.birthYear, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	birthYearIndex.Get(acc.birthYear).(*treemap.Map).Put(acc.ID, &acc)
	//}
	//if acc.Fname != "" {
	//	if !fnameIndex.Exists(acc.Fname) {
	//		fnameIndex.Update(acc.Fname, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	fnameIndex.Get(acc.Fname).(*treemap.Map).Put(acc.ID, &acc)
	//}
	//if acc.Sname != "" {
	//	if !snameIndex.Exists(acc.Sname) {
	//		snameIndex.Update(acc.Sname, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	snameIndex.Get(acc.Sname).(*treemap.Map).Put(acc.ID, &acc)
	//}
	//if acc.Sex != "" {
	//	if !sexIndex.Exists(acc.Sex) {
	//		sexIndex.Update(acc.Sex, treemap.NewWith(inverseUint32Comparator))
	//	}
	//	sexIndex.Get(acc.Sex).(*treemap.Map).Put(acc.ID, &acc)
	//}

	accountIndex.Put(acc.ID, &acc)
}

func calculateSimilarityForUser(account *Account) *treemap.Map {
	user1Likes := account.likes
	if len(user1Likes) == 0 {
		return nil
	}
	userSimilarityMap := treemap.NewWith(inverseFloat32Comparator)
	var similarMap = map[*Account]float32{}

	for likeId, tsList := range user1Likes {
		ts1 := tsList.getTimestamp()
		it := likeeIndex.Get(likeId).(*treemap.Map).Iterator()

		for it.Next() {
			similarAcc := it.Value().(*Account)
			ts2 := similarAcc.likes[likeId].getTimestamp()

			if ts1 == ts2 {
				similarMap[similarAcc] += 1
			} else {
				similarMap[similarAcc] += float32(1 / math.Abs(float64(ts1-ts2)))
			}
		}
	}

	for similarAcc, similarity := range similarMap {
		userSimilarityMap.Put(similarity, similarAcc)
	}

	return userSimilarityMap
}

func updateLikes(data []byte) {
	p := pp.Get()

	jsonData, _ := p.ParseBytes(data)

	likes := jsonData.GetArray("likes")

	for _, v := range likes {
		likerId := v.GetInt("liker")
		likeeId := v.GetInt("likee")
		ts := v.GetInt("ts")

		if likerId == 0 || likeeId == 0 || ts == 0 {
			continue
		}

		liker, _ := accountIndex.Get(likerId)

		likerAcc := liker.(*Account)
		if likerAcc.likes == nil {
			likerAcc.likes = make(map[int]LikesList, 0)
		}

		likerAcc.AppendLike(likeeId, ts)

		if !likeeIndex.Exists(likeeId) {
			likeeIndex.Update(likeeId, treemap.NewWith(inverseUint32Comparator))
		}
		likeeIndex.Get(likeeId).(*treemap.Map).Put(likerAcc.ID, likerAcc)
	}

	pp.Put(p)
}
