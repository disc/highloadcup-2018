package main

import (
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
	accountMap   = treemap.NewWith(inverseIntComparator)
	countryMap   = map[string]*treemap.Map{}
	cityMap      = map[string]*treemap.Map{}
	birthYearMap = map[int64]*treemap.Map{}
	fnameMap     = map[string]*treemap.Map{}
	snameMap     = map[string]*treemap.Map{}
)

type Account struct {
	record        map[string]gjson.Result
	interestsTree *trie.Trie
	emailBytes    []byte
	emailDomain   string
	phoneCode     int
	birthYear     int64
	premiumFinish int64
	likesMap      map[int64]int
}

func UpdateAccount(data gjson.Result) {
	record := data.Map()
	recordId := int(record["id"].Int())
	country := record["country"].String()
	city := record["city"].String()
	fname := record["fname"].String()
	sname := record["sname"].String()

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

	likesMap := make(map[int64]int, 0)
	if record["likes"].IsArray() && len(record["likes"].Array()) > 0 {
		record["likes"].ForEach(func(key, value gjson.Result) bool {
			like := value.Map()
			likesMap[like["id"].Int()] = 1

			return true
		})
	}

	account := &Account{
		record,
		interestsTree,
		[]byte(record["email"].String()),
		emailDomain,
		phoneCode,
		birthYear,
		premiumFinish,
		likesMap,
	}

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
