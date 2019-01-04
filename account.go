package main

import (
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/tidwall/gjson"
)

type AccountResponse map[string]interface{}

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	accountMap        = treemap.NewWith(inverseIntComparator)
	interestsIndexMap = map[string]*arraylist.List{}
)

type Account struct {
	record map[string]gjson.Result
}

func UpdateAccount(data gjson.Result) {
	record := data.Map()
	recordId := int(record["id"].Int())

	account := &Account{
		record,
	}

	record["interests"].ForEach(func(key, value gjson.Result) bool {
		val := value.String()
		if _, ok := interestsIndexMap[val]; !ok {
			interestsIndexMap[val] = arraylist.New()
		}

		interestsIndexMap[val].Add(val)

		return true
	})

	//todo: try set
	accountMap.Put(recordId, account)
}

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
