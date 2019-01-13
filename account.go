package main

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	inverseFloat32Comparator = func(a, b interface{}) int {
		return -utils.Float32Comparator(a, b)
	}
	accountIndex   = treemap.NewWith(inverseIntComparator)
	countryIndex   = NewSafeIndex()
	cityIndex      = NewSafeIndex()
	birthYearIndex = NewSafeIndex()
	fnameIndex     = NewSafeIndex()
	snameIndex     = NewSafeIndex()
	sexIndex       = NewSafeIndex()
	interestsIndex = NewSafeIndex()
	likeeIndex     = NewSafeIndex() // who liked this user
	emailIndex     = NewSafeIndex()
	phoneIndex     = NewSafeIndex()
)

type Account struct {
	ID        int             `json:"id"`
	Email     string          `json:"email"`
	Fname     string          `json:"fname"`
	Sname     string          `json:"sname"`
	Phone     string          `json:"phone"`
	Sex       string          `json:"sex"`
	Birth     int             `json:"birth"`
	Country   string          `json:"country"`
	City      string          `json:"city"`
	Joined    int             `json:"joined"`
	Status    string          `json:"status"`
	Interests []string        // temp data, cleared when user parsed
	Premium   map[string]int  `json:"premium"`
	TempLikes json.RawMessage `json:"likes"` // temp data, cleared when user parsed

	interestsMap map[string]struct{}
	emailDomain  string
	phoneCode    int
	birthYear    int
	joinedYear   int
	likes        map[int]LikesList
}

func (acc Account) hasActivePremium(now int64) bool {
	return acc.Premium["start"] <= int(now) && acc.Premium["finish"] > int(now)
}

// Update user
func (acc *Account) Update(changedData map[string]interface{}) {
	// 11 (w/o likes)
	if newValue, ok := changedData["interests"]; ok {
		// delete old value from indexes
		for _, v := range acc.Interests {
			interestsIndex.Get(v).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.interestsMap = make(map[string]struct{})
		for _, v := range newValue.([]interface{}) {
			interest := v.(string)
			acc.interestsMap[interest] = struct{}{}
			if !interestsIndex.Exists(interest) {
				interestsIndex.Update(interest, treemap.NewWith(inverseIntComparator))
			}
			interestsIndex.Get(interest).(*treemap.Map).Put(acc.ID, acc)
		}
		acc.Interests = nil
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
				birthYearIndex.Update(acc.birthYear, treemap.NewWith(inverseIntComparator))
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
			countryIndex.Get(acc.Country).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.Country = newValue.(string)
		if !countryIndex.Exists(acc.Country) {
			countryIndex.Update(acc.Country, treemap.NewWith(inverseIntComparator))
		}
		countryIndex.Get(acc.Country).(*treemap.Map).Put(acc.ID, acc)
	}

	if newValue, ok := changedData["city"]; ok {
		// delete old value from indexes
		if cityIndex.Exists(acc.City) {
			cityIndex.Get(acc.City).(*treemap.Map).Remove(acc.ID)
		}

		// set new value
		acc.City = newValue.(string)
		if !cityIndex.Exists(acc.City) {
			cityIndex.Update(acc.City, treemap.NewWith(inverseIntComparator))
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
			fnameIndex.Update(acc.Fname, treemap.NewWith(inverseIntComparator))
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
			snameIndex.Update(acc.Sname, treemap.NewWith(inverseIntComparator))
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
			sexIndex.Update(acc.Sex, treemap.NewWith(inverseIntComparator))
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

func NewAccount(acc Account) {
	if len(acc.Interests) > 0 {
		acc.interestsMap = make(map[string]struct{})
		for _, interest := range acc.Interests {
			acc.interestsMap[interest] = struct{}{}
			if !interestsIndex.Exists(interest) {
				interestsIndex.Update(interest, treemap.NewWith(inverseIntComparator))
			}
			interestsIndex.Get(interest).(*treemap.Map).Put(acc.ID, &acc)
		}
		acc.Interests = nil
	}

	if acc.Email != "" {
		components := strings.Split(acc.Email, "@")
		if len(components) > 1 {
			acc.emailDomain = components[1]
		}
		emailIndex.Update(acc.Email, 1)
	}

	if acc.Phone != "" {
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
			acc.phoneCode = phoneCode
		}
		phoneIndex.Update(acc.Phone, struct{}{})
	}

	if acc.Birth != 0 {
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(int64(acc.Birth), 0)
		acc.birthYear = tm.In(loc).Year()
	}

	if acc.Joined != 0 {
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(int64(acc.Joined), 0)
		acc.joinedYear = tm.In(loc).Year()
	}

	if len(acc.TempLikes) > 0 {
		acc.likes = make(map[int]LikesList, 0)
		gjson.ParseBytes(acc.TempLikes).ForEach(func(key, value gjson.Result) bool {
			like := value.Map()
			likeId := int(like["id"].Int())

			acc.likes[likeId] = append(acc.likes[likeId], int(like["ts"].Int()))

			if !likeeIndex.Exists(likeId) {
				likeeIndex.Update(likeId, treemap.NewWith(inverseIntComparator))
			}
			likeeIndex.Get(likeId).(*treemap.Map).Put(acc.ID, &acc)
			return true
		})
		acc.TempLikes = nil
	}

	if acc.Country != "" {
		if !countryIndex.Exists(acc.Country) {
			countryIndex.Update(acc.Country, treemap.NewWith(inverseIntComparator))
		}
		countryIndex.Get(acc.Country).(*treemap.Map).Put(acc.ID, &acc)
	}
	if acc.City != "" {
		if !cityIndex.Exists(acc.City) {
			cityIndex.Update(acc.City, treemap.NewWith(inverseIntComparator))
		}
		cityIndex.Get(acc.City).(*treemap.Map).Put(acc.ID, &acc)
	}
	if acc.birthYear > 0 {
		if !birthYearIndex.Exists(acc.birthYear) {
			birthYearIndex.Update(acc.birthYear, treemap.NewWith(inverseIntComparator))
		}
		birthYearIndex.Get(acc.birthYear).(*treemap.Map).Put(acc.ID, &acc)
	}
	if acc.Fname != "" {
		if !fnameIndex.Exists(acc.Fname) {
			fnameIndex.Update(acc.Fname, treemap.NewWith(inverseIntComparator))
		}
		fnameIndex.Get(acc.Fname).(*treemap.Map).Put(acc.ID, &acc)
	}
	if acc.Sname != "" {
		if !snameIndex.Exists(acc.Sname) {
			snameIndex.Update(acc.Sname, treemap.NewWith(inverseIntComparator))
		}
		snameIndex.Get(acc.Sname).(*treemap.Map).Put(acc.ID, &acc)
	}
	if acc.Sex != "" {
		if !sexIndex.Exists(acc.Sex) {
			sexIndex.Update(acc.Sex, treemap.NewWith(inverseIntComparator))
		}
		sexIndex.Get(acc.Sex).(*treemap.Map).Put(acc.ID, &acc)
	}

	accountIndex.Put(acc.ID, &acc)
}

func calculateSimilarityForUser(account *Account) *treemap.Map {
	user1Likes := account.likes
	if len(user1Likes) == 0 {
		return nil
	}
	userSimilarityMap := treemap.NewWith(inverseFloat32Comparator)
	for likeId, tsList := range user1Likes {
		ts1 := tsList.getTimestamp()
		it := likeeIndex.Get(likeId).(*treemap.Map).Iterator()
		for it.Next() {
			acc2 := it.Value().(*Account)
			ts2 := acc2.likes[likeId].getTimestamp()
			var similarity float32
			if ts1 == ts2 {
				similarity += 1
			} else {
				similarity += float32(1 / math.Abs(float64(ts1-ts2)))
			}
			if similarity > 0 {
				userSimilarityMap.Put(similarity, acc2)
			}
		}
	}

	return userSimilarityMap
}

func updateLikes(data json.RawMessage) {
	gjson.ParseBytes(data).ForEach(func(key, value gjson.Result) bool {
		value.ForEach(func(key, value gjson.Result) bool {
			like := value.Map()
			likerId := int(like["liker"].Int())
			likeeId := int(like["likee"].Int())

			liker, found := accountIndex.Get(likerId)
			if !found {
				a := "error"
				_ = a

				return true
			}

			if liker.(*Account).likes == nil {
				liker.(*Account).likes = make(map[int]LikesList, 0)
			}

			liker.(*Account).likes[likeeId] = append(liker.(*Account).likes[likeeId], int(like["ts"].Int()))

			if !likeeIndex.Exists(likeeId) {
				likeeIndex.Update(likeeId, treemap.NewWith(inverseIntComparator))
			}
			likeeIndex.Get(likeeId).(*treemap.Map).Put(liker.(*Account).ID, liker.(*Account))

			return true
		})

		return true
	})
}

//if len(acc.TempLikes) > 0 {
//acc.likes = make(map[int]LikesList, 0)
//gjson.ParseBytes(acc.TempLikes).ForEach(func(key, value gjson.Result) bool {
//like := value.Map()
//likeId := int(like["id"].Int())
//
//acc.likes[likeId] = append(acc.likes[likeId], int(like["ts"].Int()))
//
//if !likeeIndex.Exists(likeId) {
//likeeIndex.Update(likeId, treemap.NewWith(inverseIntComparator))
//}
//likeeIndex.Get(likeId).(*treemap.Map).Put(acc.ID, &acc)
//return true
//})
//acc.TempLikes = nil
//}
