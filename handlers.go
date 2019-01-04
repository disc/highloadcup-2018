package main

import (
	"encoding/json"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/emirpasic/gods/sets/treeset"

	"github.com/valyala/fasthttp"
)

type FilterResponse struct {
	Accounts []AccountResponse `json:"accounts,string"`
}

func filterHandler(ctx *fasthttp.RequestCtx) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"sex_eq":       1,
		"email_domain": 1, "email_lt": 1, "email_gt": 1,
		"status_eq": 1, "status_neq": 1,
		"fname_eq": 1, "fname_any": 1, "fname_null": 1,
		"sname_eq": 1, "sname_starts": 1, "sname_null": 1,
		"phone_code": 1, "phone_null": 1,
		"country_eq": 1, "country_null": 1,
		"city_eq": 1, "city_any": 1, "city_null": 1,
		"birth_year": 1, "birth_lt": 1, "birth_gt": 1,
		"interests_contains": 1, "interests_any": 1,
		"likes_contains": 1,
		"premium_now":    1, "premium_null": 1,
	}

	validQueryArgs := true
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		if _, ok := allowedParams[string(key)]; !ok {
			validQueryArgs = false
			return
		}
	})
	if !validQueryArgs {
		ctx.Error("{}", 400)
		return
	}

	// ignore interests, likes
	responseProperties := []string{
		"id", "email",
	}

	limit := 0
	if ctx.QueryArgs().Has("limit") {
		limit, _ = strconv.Atoi(string(ctx.QueryArgs().Peek("limit")))
	}
	sexEqF := ctx.QueryArgs().Peek("sex_eq")
	if len(sexEqF) > 0 {
		responseProperties = append(responseProperties, "sex")
	}

	emailDomainF := ctx.QueryArgs().Peek("email_domain")
	emailLtF := ctx.QueryArgs().Peek("email_lt")
	emailGtF := ctx.QueryArgs().Peek("email_gt")
	if len(emailLtF) > 0 || len(emailGtF) > 0 || len(emailDomainF) > 0 {
		responseProperties = append(responseProperties, "email")
	}

	statusEqF := ctx.QueryArgs().Peek("status_eq")
	statusNeqF := ctx.QueryArgs().Peek("status_neq")
	if len(statusEqF) > 0 || len(statusNeqF) > 0 {
		responseProperties = append(responseProperties, "status")
	}

	fnameEqF := ctx.QueryArgs().Peek("fname_eq")
	fnameAnyF := ctx.QueryArgs().Peek("fname_any")
	fnameNullF := ctx.QueryArgs().Peek("fname_null")
	if len(fnameEqF) > 0 || len(fnameAnyF) > 0 {
		responseProperties = append(responseProperties, "fname")
	}
	//
	//snameEqF := ctx.QueryArgs().Peek("sname_eq")
	//snameStartsF := ctx.QueryArgs().Peek("sname_starts")
	//snameNullF := ctx.QueryArgs().Peek("sname_null")
	//if len(snameEqF) > 0 || len(snameStartsF) > 0 {
	//	responseProperties = append(responseProperties, "sname")
	//}
	//
	//phoneCodeF := ctx.QueryArgs().Peek("phone_code")
	//phoneNullF := ctx.QueryArgs().Peek("phone_null")
	//if len(phoneCodeF) > 0 {
	//	responseProperties = append(responseProperties, "phone")
	//}
	//
	//countryEqF := ctx.QueryArgs().Peek("country_eq")
	//countryNullF := ctx.QueryArgs().Peek("country_null")
	//if len(countryEqF) > 0 {
	//	responseProperties = append(responseProperties, "country")
	//}
	//
	//cityEqF := ctx.QueryArgs().Peek("city_eq")
	//cityAnyF := ctx.QueryArgs().Peek("city_any")
	//cityNullF := ctx.QueryArgs().Peek("city_null")
	//if len(cityEqF) > 0 || len(cityAnyF) > 0 {
	//	responseProperties = append(responseProperties, "city")
	//}
	//
	//birthLtF := ctx.QueryArgs().Peek("birth_lt")
	//birthGtF := ctx.QueryArgs().Peek("birth_gt")
	//birthYearF := ctx.QueryArgs().Peek("birth_year")
	//if len(birthLtF) > 0 || len(birthGtF) > 0 || len(birthYearF) > 0 {
	//	responseProperties = append(responseProperties, "birth")
	//}
	//
	interestsContainsF := ctx.QueryArgs().Peek("interests_contains")
	interestsAnyF := ctx.QueryArgs().Peek("interests_any")

	//
	//likesContainsF := ctx.QueryArgs().Peek("likes_contains")
	//
	//premiumNowF := ctx.QueryArgs().Peek("premium_now")
	//premiumNullF := ctx.QueryArgs().Peek("premium_null")
	//if len(premiumNowF) > 0 {
	//	responseProperties = append(responseProperties, "premium")
	//}

	var resultIds []map[string]gjson.Result

	//tempResults := make([]*treeset.Set, 0)

	filters := make(map[string]interface{})

	if len(sexEqF) > 0 {
		filters["sex_eq"] = string(sexEqF)
	}
	if len(emailLtF) > 0 {
		filters["email_lt"] = string(emailLtF)
	}
	if len(statusEqF) > 0 {
		filters["status_eq"] = string(statusEqF)
	}
	if len(statusNeqF) > 0 {
		filters["status_neq"] = string(statusNeqF)
	}
	if len(fnameNullF) > 0 {
		if string(fnameNullF) == "0" {
			filters["fname_not_null"] = 1
		} else {
			filters["fname_null"] = 1
		}

	}
	var interestsFilter []interface{}
	if len(interestsAnyF) > 0 {
		words := strings.Split(string(interestsAnyF), ",")
		if len(words) > 0 {
			interestsFilter = make([]interface{}, len(words))
			filters["interests_any"] = words
			for i, v := range words {
				interestsFilter[i] = v
			}
		}
	}
	if len(interestsContainsF) > 0 {
		words := strings.Split(string(interestsContainsF), ",")
		if len(words) > 0 {
			interestsFilter = make([]interface{}, len(words))
			filters["interests_contains"] = words
			for i, v := range words {
				interestsFilter[i] = v
			}
		}
	}
	filtersCount := len(filters)
	if filtersCount == 0 {
		//resultIds = append(resultIds, GetIdFromKey(key))
		//return len(resultIds) < limit
	}

	// full scan search
	it := accountMap.Iterator()
	for it.Next() {
		if len(resultIds) >= limit {
			break
		}
		passedFilters := 0
		account := *it.Value().(*Account)
		value := account.record
		if sexEqFilter, ok := filters["sex_eq"]; ok && value["sex"].Value() == sexEqFilter {
			passedFilters += 1
		}
		if statusEqFilter, ok := filters["status_eq"]; ok && value["status"].Value() == statusEqFilter {
			passedFilters += 1
		}
		if statusNeqFilter, ok := filters["status_neq"]; ok && value["status"].Value() != statusNeqFilter {
			passedFilters += 1
		}
		if _, ok := filters["fname_null"]; ok && value["fname"].String() == "" {
			passedFilters += 1
		}
		if _, ok := filters["fname_not_null"]; ok && value["fname"].String() != "" {
			passedFilters += 1
		}
		if &interestsFilter != nil {
			//start := time.Now()
			if account.interestsList.Size() == 0 {
				continue
			}
			if account.interestsList.Contains(interestsFilter...) {
				passedFilters += 1
			}
			//log.Printf("contains took %s", time.Since(start))
		}
		if passedFilters == filtersCount {
			resultIds = append(resultIds, value)
		}
	}

	// index search
	//if emailLtFilter, ok := filters["email_lt"].(string); ok && value["email"].String()[0:len(emailLtFilter)] < emailLtFilter {
	//	passedFilters += 1
	//}

	// order by ID desc
	// apply limit

	jsonData := []byte(`{"accounts":[]}`)
	if len(resultIds) > 0 {
		jsonData, _ = json.Marshal(prepareResponse(resultIds, responseProperties))
	}

	// TODO: Use sjson for updates
	ctx.Success("application/json", jsonData)
	return
}

func intersectFoundResults(tempResults []*treeset.Set, limit int) []int {
	resultIds := make([]int, 0)
	// find smalest set
	var smalest *treeset.Set
	for _, set := range tempResults {
		if smalest == nil {
			smalest = set
			continue
		}
		if set.Size() < smalest.Size() {
			smalest = set
		}
	}

	if smalest != nil {
		it := smalest.Iterator()
		for it.Next() {
			if len(resultIds) >= limit {
				break
			}
			value := it.Value()
			ok := true
			for _, tempSet := range tempResults {
				if *tempSet == *smalest {
					continue
				}
				if !tempSet.Contains(value) {
					ok = false
					break
				}
			}
			if ok {
				resultIds = append(resultIds, value.(int))
			}
		}
	}

	return resultIds
}

func diffFoundResults(ignoreSet *treeset.Set, tempResults []*treeset.Set) *treeset.Set {
	resultIds := make([]int, 0)
	for _, tempSet := range tempResults {
		it := tempSet.Iterator()
		for it.Next() {
			value := it.Value()

			if ignoreSet.Contains(value) {
				continue
			}

			resultIds = append(resultIds, value.(int))
		}
	}

	resultSet := treeset.NewWith(inverseIntComparator)

	return resultSet
}

func prepareResponse(found []map[string]gjson.Result, responseProperties []string) *FilterResponse {
	var results = make([]AccountResponse, 0)
	for _, account := range found {
		result := make(AccountResponse, 0)
		for _, key := range responseProperties {
			result[key] = account[key].Value()
		}
		results = append(results, result)
	}

	return &FilterResponse{
		Accounts: results,
	}
}

func eqFilter(sourceMap map[interface{}][]int, value interface{}) []int {
	resultIds := make([]int, 0)

	resultIds = sourceMap[value]

	return resultIds
}

func neqFilter(source map[string]*treeset.Set, value string) *treeset.Set {
	start := time.Now()
	sets := make([]*treeset.Set, 0)
	for key, set := range source {
		if key == value {
			continue
		}
		sets = append(sets, set)
	}
	log.Printf("range sources took %s", time.Since(start))
	resultSet := diffFoundResults(source[value], sets)
	log.Printf("nonEQ filter took %s", time.Since(start))

	return resultSet
}

func ltFilter(fKey string, fVal string) []int {
	resultIds := make([]int, 0)
	//_ = tx.AscendLessThan(fKey, fVal, func(key, val string) bool {
	//	resultIds = append(resultIds, GetIdFromKey(key))
	//	return true //TODO: get not more than limit
	//})

	return resultIds
}

func gtFilter(fKey string, fVal string) []int {
	resultIds := make([]int, 0)
	//_ = tx.AscendGreaterOrEqual(fKey, fVal, func(key, val string) bool {
	//	resultIds = append(resultIds, GetIdFromKey(key))
	//	return true
	//})

	return resultIds
}

func nullFilter(fKey string, fVal string) []int {
	// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
	resultIds := make([]int, 0)

	if fVal == "0" {
		//_ = tx.Descend(fKey, func(key, val string) bool {
		//	value := gjson.Parse(val).Get(fKey)
		//
		//	isNotEmpty := value.Exists() || value.String() != ""
		//
		//	if isNotEmpty {
		//		resultIds = append(resultIds, GetIdFromKey(key))
		//	}
		//
		//	return isNotEmpty
		//})
	}

	if fVal == "1" {
		//_ = tx.Ascend(fKey, func(key, val string) bool {
		//	value := gjson.Parse(val).Get(fKey)
		//
		//	isEmpty := !value.Exists() || value.String() == ""
		//
		//	if isEmpty {
		//		resultIds = append(resultIds, GetIdFromKey(key))
		//	}
		//
		//	return isEmpty
		//})
	}

	return resultIds
}

func anyFilter(fKey string, fVal string) []int {
	resultIds := make([]int, 0)

	_ = strings.Split(fVal, ",")

	//if valid {
	//	resultIds = append(resultIds, GetIdFromKey(key))
	//}

	return resultIds
}

func containsFilter(fKey string, fVal string) []int {
	resultIds := make([]int, 0)

	_ = strings.Split(fVal, ",")

	//if valid {
	//	resultIds = append(resultIds, GetIdFromKey(key))
	//}

	return resultIds
}

func processResults(results []int, original []int) []int {
	if len(original) > 0 {
		original = intSliceIntersection(results, original)
	} else {
		original = results
	}

	return original
}

func stringSliceIntersection(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

func intSliceIntersection(a, b []int) (c []int) {
	m := make(map[int]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

type QInterface interface {
	sort.Interface
	// Partition returns slice[:i] and slice[i+1:]
	// These should references the original memory
	// since this does an in-place sort
	Partition(i int) (left QInterface, right QInterface)
}

func Qsort(a QInterface, prng *rand.Rand) QInterface {
	if a.Len() < 2 {
		return a
	}

	left, right := 0, a.Len()-1

	// Pick a pivot
	pivotIndex := prng.Int() % a.Len()
	// Move the pivot to the right
	a.Swap(pivotIndex, right)

	// Pile elements smaller than the pivot on the left
	for i := 0; i < a.Len(); i++ {
		if a.Less(i, right) {

			a.Swap(i, left)
			left++
		}
	}

	// Place the pivot after the last smaller element
	a.Swap(left, right)

	// Go down the rabbit hole
	leftSide, rightSide := a.Partition(left)
	Qsort(leftSide, prng)
	Qsort(rightSide, prng)

	return a
}

type QIntSlice []int

func (is QIntSlice) Less(i, j int) bool {
	return is[i] < is[j]
}

func (is QIntSlice) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

func (is QIntSlice) Len() int {
	return len(is)
}

func (is QIntSlice) Partition(i int) (left QInterface, right QInterface) {
	return QIntSlice(is[:i]), QIntSlice(is[i+1:])
}
