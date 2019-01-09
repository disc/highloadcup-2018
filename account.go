package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

type AccountResponse map[string]interface{}

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	inverseFloat64Comparator = func(a, b interface{}) int {
		return -utils.Float64Comparator(a, b)
	}
	accountMap         = treemap.NewWith(inverseIntComparator)
	countryMap         = map[string]*treemap.Map{}
	cityMap            = map[string]*treemap.Map{}
	birthYearMap       = map[int]*treemap.Map{}
	fnameMap           = map[string]*treemap.Map{}
	snameMap           = map[string]*treemap.Map{}
	similarityMap      = map[int]*treemap.Map{}
	globalInterestsMap = map[string]*treemap.Map{}
	globalLikesMap     = map[*Account][]*Account{}
	likersIndex        = map[int]*treemap.Map{} // likes by user
	likeeIndex         = map[int]*treemap.Map{} // who liked this user
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

	interestsMap  map[string]struct{} // try map[string]struct{}{}
	emailDomain   string
	phoneCode     int
	birthYear     int
	premiumFinish int64
	likes         map[int][]int
}

func createAccount(acc Account) {
	// TODO: unset likes, interests

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
		if len(components) > 0 {
			acc.emailDomain = components[1]
		}
	}

	if acc.Phone != "" {
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
			acc.phoneCode = phoneCode
		}
	}

	if acc.Birth != 0 {
		tm := time.Unix(int64(acc.Birth), 0)
		acc.birthYear = tm.Year()
	}
	if finish, ok := acc.Premium["finish"]; ok {
		acc.premiumFinish = int64(finish)
	}

	if len(acc.TempLikes) > 0 {
		acc.likes = make(map[int][]int, 0)
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

	//if len(acc.TempLikes) > 0 {
	//	likesMap := make(map[int][]int, 0)
	//	for _, like := range acc.TempLikes {
	//		likesMap[like["id"]] = append(likesMap[like["id"]], like["ts"])
	//
	//		//globalLikesMap[accId] = append(globalLikesMap[accId], &acc)
	//	}
	//	acc.TempLikes = nil
	//
	//	uniqLikeMap := map[int]int{}
	//	for id, timestamps := range likesMap {
	//		var ts int
	//		if len(timestamps) > 1 {
	//			var total = 0
	//			for _, value := range timestamps {
	//				total += value
	//			}
	//			ts = total / int(len(timestamps))
	//		} else {
	//			ts = timestamps[0]
	//		}
	//		uniqLikeMap[id] = ts
	//	}
	//	//acc.likes = uniqLikeMap
	//}

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

	accountMap.Put(acc.ID, &acc)
}

func calculateSimilarityIndex() {
	accountMap.Each(func(key interface{}, value interface{}) {
		calculateSimilarityForUser(value.(*Account))
	})
	// calculate for ine user
	//value, _ := accountMap.Get(24156)
	//calculateSimilarityForUser(value.(*Account))
}

func calculateSimilarityForUser(account *Account) {
	//user1Likes := account.likes
	//for likeId, ts1 := range user1Likes {
	//	for _, acc2 := range globalLikesMap[likeId] {
	//		ts2 := acc2.likes[likeId]
	//		var similarity float64
	//		if ts1 == ts2 {
	//			similarity += 1
	//		} else {
	//			similarity += 1 / math.Abs(float64(ts1-ts2))
	//		}
	//		if similarity > 0 {
	//			user1Id := account.ID
	//			if _, ok := similarityMap[user1Id]; !ok {
	//				similarityMap[user1Id] = treemap.NewWith(inverseFloat64Comparator)
	//			}
	//			similarityMap[user1Id].Put(similarity, acc2)
	//		}
	//	}
	//}
}
