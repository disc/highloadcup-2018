package main

import (
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
	"github.com/tidwall/gjson"
)

type Account map[string]interface{}

type AccountResult *gjson.Result

func GetAccount(id int) AccountResult {
	return accountMap[id]
}

type HashMap map[string][]int
type TreeSet map[string]*treeset.Set

type AccountMap map[int]AccountResult

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	accountMap = make(AccountMap, 0)
	sexMap     = TreeSet{
		"m": treeset.NewWith(inverseIntComparator),
		"f": treeset.NewWith(inverseIntComparator),
	}
	statusMap = TreeSet{
		"свободны":   treeset.NewWith(inverseIntComparator),
		"заняты":     treeset.NewWith(inverseIntComparator),
		"всё сложно": treeset.NewWith(inverseIntComparator),
	}
	countryMap = make(HashMap, 0)
	cityMap    = make(HashMap, 0)
	fnameMap   = make(HashMap, 0)
	snameMap   = make(HashMap, 0)
)

/**
03 sex_eq
489 country_eq sex_eq
476 country_null sex_eq
416
289 interests_contains sex_eq
279 interests_any sex_eq
271 sex_eq status_neq
254 sex_eq status_eq
233 country_eq
201 city_eq sex_eq
197 city_any sex_eq
195 city_null sex_eq
192 country_null
192 likes_contains
189 interests_any sex_eq status_eq
186 interests_contains sex_eq status_neq
186 country_eq email_gt sex_eq
185 interests_any sex_eq status_neq
179 interests_contains sex_eq status_eq
179 country_null email_lt sex_eq
*/

func UpdateAccount(data gjson.Result) {
	record := data.Map()
	recordId := int(record["id"].Int())
	sex := record["sex"].String()
	country := record["country"].String()
	city := record["city"].String()
	status := record["status"].String()
	fname := record["fname"].String()
	sname := record["sname"].String()

	if sex != "" {
		sexMap[sex].Add(recordId)
	}
	if country != "" {
		countryMap[country] = append(countryMap[country], recordId)
	}
	if city != "" {
		cityMap[city] = append(cityMap[city], recordId)
	}
	if status != "" {
		statusMap[status].Add(recordId)
	}
	if fname != "" {
		fnameMap[fname] = append(fnameMap[fname], recordId)
	}
	if sname != "" {
		snameMap[sname] = append(snameMap[sname], recordId)
	}

	accountMap[recordId] = &data
}
