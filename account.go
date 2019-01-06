package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/derekparker/trie"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/tidwall/gjson"
)

type AccountResponse map[string]interface{}

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	inverseFloat64Comparator = func(a, b interface{}) int {
		return -utils.Float64Comparator(a, b)
	}
	accountMap     = treemap.NewWith(inverseIntComparator)
	countryMap     = map[string]*treemap.Map{}
	cityMap        = map[string]*treemap.Map{}
	birthYearMap   = map[int64]*treemap.Map{}
	fnameMap       = map[string]*treemap.Map{}
	snameMap       = map[string]*treemap.Map{}
	similarityMap  = map[int]*treemap.Map{}
	globalLikesMap = map[int][]*Account{}
)

type Account struct {
	id            int
	record        map[string]gjson.Result
	interestsTree *trie.Trie
	emailBytes    []byte
	emailDomain   string
	phoneCode     int
	birthYear     int64
	premiumFinish int64
	sex           string //FIXME: use byte or rune
	likes         map[int]int
}

func UpdateAccount(data gjson.Result) {
	// TODO: unset likes, interests
	record := data.Map()
	recordId := int(record["id"].Int())
	country := record["country"].String()
	city := record["city"].String()
	fname := record["fname"].String()
	sname := record["sname"].String()
	sex := record["sex"].String()

	interestsTree := trie.New()

	record["interests"].ForEach(func(key, value gjson.Result) bool {
		interestsTree.Add(value.String(), 1)

		return true
	})

	var emailDomain string
	if record["email"].Exists() {
		components := strings.Split(record["email"].String(), "@")
		emailDomain = components[1]
	}

	var phoneCode int
	if record["phone"].Exists() {
		phoneCodeStr := strings.SplitN(strings.SplitN(record["phone"].String(), "(", 2)[1], ")", 2)[0]
		phoneCode, _ = strconv.Atoi(phoneCodeStr)
	}

	var birthYear int64
	if record["birth"].Exists() {
		tm := time.Unix(record["birth"].Int(), 0)
		birthYear = int64(tm.Year())
	}
	var premiumFinish int64
	if record["premium"].IsObject() {
		premiumFinish = record["premium"].Map()["finish"].Int()
	}

	account := &Account{
		recordId,
		record,
		interestsTree,
		[]byte(record["email"].String()),
		emailDomain,
		phoneCode,
		birthYear,
		premiumFinish,
		sex,
		nil,
	}

	likesMap := map[int][]int{}
	if record["likes"].IsArray() && len(record["likes"].Array()) > 0 {
		record["likes"].ForEach(func(key, value gjson.Result) bool {
			like := value.Map()
			accId := int(like["id"].Int())
			likesMap[accId] = append(likesMap[accId], int(like["ts"].Int()))

			globalLikesMap[accId] = append(globalLikesMap[accId], account)

			return true
		})
	}

	uniqLikeMap := map[int]int{}
	for id, likes := range likesMap {
		var ts int
		if len(likes) > 1 {
			var total = 0
			for _, value := range likes {
				total += value
			}
			ts = total / int(len(likes))
		} else {
			ts = likes[0]
		}
		uniqLikeMap[id] = ts
	}

	account.likes = uniqLikeMap

	if country != "" {
		if _, ok := countryMap[country]; !ok {
			countryMap[country] = treemap.NewWith(inverseIntComparator)
		}
		countryMap[country].Put(recordId, account)
	}
	if city != "" {
		if _, ok := cityMap[city]; !ok {
			cityMap[city] = treemap.NewWith(inverseIntComparator)
		}
		cityMap[city].Put(recordId, account)
	}
	if birthYear > 0 {
		if _, ok := birthYearMap[birthYear]; !ok {
			birthYearMap[birthYear] = treemap.NewWith(inverseIntComparator)
		}
		birthYearMap[birthYear].Put(recordId, account)
	}
	if fname != "" {
		if _, ok := fnameMap[fname]; !ok {
			fnameMap[fname] = treemap.NewWith(inverseIntComparator)
		}
		fnameMap[fname].Put(recordId, account)
	}
	if sname != "" {
		if _, ok := snameMap[sname]; !ok {
			snameMap[sname] = treemap.NewWith(inverseIntComparator)
		}
		snameMap[sname].Put(recordId, account)
	}

	//todo: try set
	accountMap.Put(recordId, account)
}

func calculateSimilarityIndex() {
	accountMap.Each(func(key interface{}, value interface{}) {
		calculateSimilarityForUser(value.(*Account))
	})
	//value, _ := accountMap.Get(6327)
	//calculateSimilarityForUser(value.(*Account))
}

func calculateSimilarityForUser(account *Account) {
	user1Likes := account.likes
	for likeId, ts1 := range user1Likes {
		for _, acc2 := range globalLikesMap[likeId] {
			ts2 := acc2.likes[likeId]
			var similarity float64
			if ts1 == ts2 {
				similarity += 1
			} else {
				similarity += 1 / math.Abs(float64(ts1-ts2))
			}
			if similarity > 0 {
				user1Id := account.id
				if _, ok := similarityMap[user1Id]; !ok {
					similarityMap[user1Id] = treemap.NewWith(inverseFloat64Comparator)
				}
				similarityMap[user1Id].Put(similarity, acc2)
			}
		}
	}
}
