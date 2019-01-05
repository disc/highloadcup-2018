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
	accountMap = treemap.NewWith(inverseIntComparator)
)

type Account struct {
	record        map[string]gjson.Result
	interestsTree *trie.Trie
	emailBytes    []byte
	emailDomain   string
	phoneCode     int
	birthYear     int
}

func UpdateAccount(data gjson.Result) {
	record := data.Map()
	recordId := int(record["id"].Int())

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

	var birthYear int
	if record["birth"].Exists() {
		tm := time.Unix(int64(record["birth"].Float()), 0)
		birthYear = tm.Year()
	}

	account := &Account{
		record,
		interestsTree,
		[]byte(record["email"].String()),
		emailDomain,
		phoneCode,
		birthYear,
	}

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
