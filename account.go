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
	accountMap         = treemap.NewWith(inverseIntComparator)
	countryMap         = map[string]*treemap.Map{}
	cityMap            = map[string]*treemap.Map{}
	birthYearMap       = map[int]*treemap.Map{}
	fnameMap           = map[string]*treemap.Map{}
	snameMap           = map[string]*treemap.Map{}
	sexIndex           = map[string]*treemap.Map{}
	globalInterestsMap = map[string]*treemap.Map{}
	likeeIndex         = map[int]*treemap.Map{} // who liked this user
	emailIndex         = &SafeIndex{v: make(map[interface{}]interface{})}
	phoneIndex         = &SafeIndex{v: make(map[interface{}]interface{})}
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
	if newValue, ok := changedData["interests"]; ok {
		// delete old value from indexes
		for _, v := range acc.Interests {
			globalInterestsMap[v].Remove(acc.ID)
		}

		// set new value
		acc.interestsMap = make(map[string]struct{})
		for _, v := range newValue.([]interface{}) {
			interest := v.(string)
			acc.interestsMap[interest] = struct{}{}
			if _, ok := globalInterestsMap[interest]; !ok {
				globalInterestsMap[interest] = treemap.NewWith(inverseIntComparator)
			}
			globalInterestsMap[interest].Put(acc.ID, &acc)
		}
		acc.Interests = nil
	}

	if newValue, ok := changedData["email"]; ok {
		// delete old value from indexes
		emailIndex.Delete(acc.Email)

		// set new value
		acc.Email = newValue.(string)
		components := strings.Split(acc.Email, "@")
		if len(components) > 1 {
			acc.emailDomain = components[1]
		}
		emailIndex.Update(acc.Email, struct{}{})
	}

	if newValue, ok := changedData["phone"]; ok {
		// delete old value from indexes
		phoneIndex.Delete(acc.Phone)

		// set new value
		acc.Phone = newValue.(string)
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
			acc.phoneCode = phoneCode
		}
		phoneIndex.Update(acc.Phone, struct{}{})
	}

	if newValue, ok := changedData["birth"]; ok {
		// set new value
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(newValue.(int64), 0)
		acc.birthYear = tm.In(loc).Year()
	}

	if newValue, ok := changedData["birth"]; ok {
		// set new value
		loc, _ := time.LoadLocation("UTC")
		tm := time.Unix(newValue.(int64), 0)
		acc.joinedYear = tm.In(loc).Year()
	}

	//
	//if len(acc.TempLikes) > 0 {
	//	acc.likes = make(map[int]LikesList, 0)
	//	gjson.ParseBytes(acc.TempLikes).ForEach(func(key, value gjson.Result) bool {
	//		like := value.Map()
	//		likeId := int(like["id"].Int())
	//
	//		acc.likes[likeId] = append(acc.likes[likeId], int(like["ts"].Int()))
	//
	//		if _, ok := likeeIndex[likeId]; !ok {
	//			likeeIndex[likeId] = treemap.NewWith(inverseIntComparator)
	//		}
	//		likeeIndex[likeId].Put(acc.ID, &acc)
	//		return true
	//	})
	//	acc.TempLikes = nil
	//}
	//
	//if acc.Country != "" {
	//	if _, ok := countryMap[acc.Country]; !ok {
	//		countryMap[acc.Country] = treemap.NewWith(inverseIntComparator)
	//	}
	//	countryMap[acc.Country].Put(acc.ID, &acc)
	//}
	//if acc.City != "" {
	//	if _, ok := cityMap[acc.City]; !ok {
	//		cityMap[acc.City] = treemap.NewWith(inverseIntComparator)
	//	}
	//	cityMap[acc.City].Put(acc.ID, &acc)
	//}
	//if acc.birthYear > 0 {
	//	if _, ok := birthYearMap[acc.birthYear]; !ok {
	//		birthYearMap[acc.birthYear] = treemap.NewWith(inverseIntComparator)
	//	}
	//	birthYearMap[acc.birthYear].Put(acc.ID, &acc)
	//}
	//if acc.Fname != "" {
	//	if _, ok := fnameMap[acc.Fname]; !ok {
	//		fnameMap[acc.Fname] = treemap.NewWith(inverseIntComparator)
	//	}
	//	fnameMap[acc.Fname].Put(acc.ID, &acc)
	//}
	//if acc.Sname != "" {
	//	if _, ok := snameMap[acc.Sname]; !ok {
	//		snameMap[acc.Sname] = treemap.NewWith(inverseIntComparator)
	//	}
	//	snameMap[acc.Sname].Put(acc.ID, &acc)
	//}
	//if acc.Sex != "" {
	//	if _, ok := sexIndex[acc.Sex]; !ok {
	//		sexIndex[acc.Sex] = treemap.NewWith(inverseIntComparator)
	//	}
	//	sexIndex[acc.Sex].Put(acc.ID, &acc)
	//}
	//
	//accountMap.Put(acc.ID, &acc)
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
			if _, ok := globalInterestsMap[interest]; !ok {
				globalInterestsMap[interest] = treemap.NewWith(inverseIntComparator)
			}
			globalInterestsMap[interest].Put(acc.ID, &acc)
		}
		acc.Interests = nil
	}

	if acc.Email != "" {
		components := strings.Split(acc.Email, "@")
		if len(components) > 1 {
			acc.emailDomain = components[1]
		}
		emailIndex.Update(acc.Email, struct{}{})
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

			if _, ok := likeeIndex[likeId]; !ok {
				likeeIndex[likeId] = treemap.NewWith(inverseIntComparator)
			}
			likeeIndex[likeId].Put(acc.ID, &acc)
			return true
		})
		acc.TempLikes = nil
	}

	if acc.Country != "" {
		if _, ok := countryMap[acc.Country]; !ok {
			countryMap[acc.Country] = treemap.NewWith(inverseIntComparator)
		}
		countryMap[acc.Country].Put(acc.ID, &acc)
	}
	if acc.City != "" {
		if _, ok := cityMap[acc.City]; !ok {
			cityMap[acc.City] = treemap.NewWith(inverseIntComparator)
		}
		cityMap[acc.City].Put(acc.ID, &acc)
	}
	if acc.birthYear > 0 {
		if _, ok := birthYearMap[acc.birthYear]; !ok {
			birthYearMap[acc.birthYear] = treemap.NewWith(inverseIntComparator)
		}
		birthYearMap[acc.birthYear].Put(acc.ID, &acc)
	}
	if acc.Fname != "" {
		if _, ok := fnameMap[acc.Fname]; !ok {
			fnameMap[acc.Fname] = treemap.NewWith(inverseIntComparator)
		}
		fnameMap[acc.Fname].Put(acc.ID, &acc)
	}
	if acc.Sname != "" {
		if _, ok := snameMap[acc.Sname]; !ok {
			snameMap[acc.Sname] = treemap.NewWith(inverseIntComparator)
		}
		snameMap[acc.Sname].Put(acc.ID, &acc)
	}
	if acc.Sex != "" {
		if _, ok := sexIndex[acc.Sex]; !ok {
			sexIndex[acc.Sex] = treemap.NewWith(inverseIntComparator)
		}
		sexIndex[acc.Sex].Put(acc.ID, &acc)
	}

	accountMap.Put(acc.ID, &acc)
}

func calculateSimilarityForUser(account *Account) *treemap.Map {
	user1Likes := account.likes
	if len(user1Likes) == 0 {
		return nil
	}
	userSimilarityMap := treemap.NewWith(inverseFloat32Comparator)
	for likeId, tsList := range user1Likes {
		ts1 := tsList.getTimestamp()
		it := likeeIndex[likeId].Iterator()
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
