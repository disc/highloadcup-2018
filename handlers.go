package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/tidwall/buntdb"
	"github.com/valyala/fasthttp"
)

type FilterResponse struct {
	Accounts []Account `json:"accounts,string"`
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

	snameEqF := ctx.QueryArgs().Peek("sname_eq")
	snameStartsF := ctx.QueryArgs().Peek("sname_starts")
	snameNullF := ctx.QueryArgs().Peek("sname_null")
	if len(snameEqF) > 0 || len(snameStartsF) > 0 {
		responseProperties = append(responseProperties, "sname")
	}

	phoneCodeF := ctx.QueryArgs().Peek("phone_code")
	phoneNullF := ctx.QueryArgs().Peek("phone_null")
	if len(phoneCodeF) > 0 {
		responseProperties = append(responseProperties, "phone")
	}

	countryEqF := ctx.QueryArgs().Peek("country_eq")
	countryNullF := ctx.QueryArgs().Peek("country_null")
	if len(countryEqF) > 0 {
		responseProperties = append(responseProperties, "country")
	}

	cityEqF := ctx.QueryArgs().Peek("city_eq")
	cityAnyF := ctx.QueryArgs().Peek("city_any")
	cityNullF := ctx.QueryArgs().Peek("city_null")
	if len(cityEqF) > 0 || len(cityAnyF) > 0 {
		responseProperties = append(responseProperties, "city")
	}

	birthLtF := ctx.QueryArgs().Peek("birth_lt")
	birthGtF := ctx.QueryArgs().Peek("birth_gt")
	birthYearF := ctx.QueryArgs().Peek("birth_year")
	if len(birthLtF) > 0 || len(birthGtF) > 0 || len(birthYearF) > 0 {
		responseProperties = append(responseProperties, "birth")
	}

	interestsContainsF := ctx.QueryArgs().Peek("interests_contains")
	interestsAnyF := ctx.QueryArgs().Peek("interests_any")

	likesContainsF := ctx.QueryArgs().Peek("likes_contains")

	premiumNowF := ctx.QueryArgs().Peek("premium_now")
	premiumNullF := ctx.QueryArgs().Peek("premium_null")
	if len(premiumNowF) > 0 {
		responseProperties = append(responseProperties, "premium")
	}

	var resultIds []int

	hasFilters := 0
	_ = db.View(func(tx *buntdb.Tx) error {
		if len(sexEqF) > 0 {
			hasFilters = 1
			value := `{"sex":"` + string(sexEqF) + `"}`
			resultIds = processResults(
				eqFilter("sex", value, tx),
				resultIds,
			)
		}
		if len(emailDomainF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				eqFilter("email_domain", string(emailDomainF), tx),
				resultIds,
			)
		}
		if len(emailLtF) > 0 {
			hasFilters = 1
			value := `{"email":"` + string(emailLtF) + `"}`
			resultIds = processResults(
				ltFilter("email", value, tx),
				resultIds,
			)
		}
		if len(emailGtF) > 0 {
			hasFilters = 1
			value := `{"email":"` + string(emailGtF) + `"}`
			resultIds = processResults(
				gtFilter("email", value, tx),
				resultIds,
			)
		}
		if len(statusEqF) > 0 {
			hasFilters = 1
			value := `{"status":"` + string(statusEqF) + `"}`
			resultIds = processResults(
				eqFilter("status", value, tx),
				resultIds,
			)
		}
		if len(statusNeqF) > 0 {
			hasFilters = 1
			value := string(statusEqF)
			resultIds = processResults(
				neqFilter("status", value, tx),
				resultIds,
			)
		}
		if len(fnameEqF) > 0 {
			hasFilters = 1
			value := `{"fname":"` + string(fnameEqF) + `"}`
			resultIds = processResults(
				eqFilter("fname", value, tx),
				resultIds,
			)
		}
		if len(fnameAnyF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				anyFilter("fname", string(fnameAnyF), tx),
				resultIds,
			)
		}
		if len(fnameNullF) > 0 {
			hasFilters = 1
			value := string(fnameNullF)
			if value == "0" {
				responseProperties = append(responseProperties, "fname")
			}
			resultIds = processResults(
				nullFilter("fname", value, tx),
				resultIds,
			)
		}
		if len(snameEqF) > 0 {
			hasFilters = 1
			value := `{"sname":"` + string(snameEqF) + `"}`
			resultIds = processResults(
				eqFilter("sname", value, tx),
				resultIds,
			)
		}
		if len(snameStartsF) > 0 {
			hasFilters = 1
			sSnameStartsF := string(snameStartsF)
			var foundResults []int
			_ = tx.Ascend("sname", func(key, val string) bool {
				if strings.HasPrefix(val, sSnameStartsF) {
					foundResults = append(resultIds, GetIdFromKey(key))
				}

				return true
			})

			resultIds = processResults(
				foundResults,
				resultIds,
			)
		}
		if len(snameNullF) > 0 {
			hasFilters = 1
			value := string(snameNullF)
			if value == "0" {
				responseProperties = append(responseProperties, "sname")
			}
			resultIds = processResults(
				nullFilter("sname", value, tx),
				resultIds,
			)
		}
		if len(phoneCodeF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				eqFilter("phone_code", string(phoneCodeF), tx),
				resultIds,
			)
		}
		if len(phoneNullF) > 0 {
			hasFilters = 1
			value := string(phoneNullF)
			if value == "0" {
				responseProperties = append(responseProperties, "phone")
			}
			resultIds = processResults(
				nullFilter("phone", value, tx),
				resultIds,
			)
		}
		if len(countryEqF) > 0 {
			hasFilters = 1
			value := `{"country":"` + string(countryEqF) + `"}`
			resultIds = processResults(
				eqFilter("country", value, tx),
				resultIds,
			)
		}
		if len(countryNullF) > 0 {
			hasFilters = 1
			value := string(countryNullF)

			if value == "0" {
				responseProperties = append(responseProperties, "country")
			}
			resultIds = processResults(
				nullFilter("country", value, tx),
				resultIds,
			)
		}
		if len(cityEqF) > 0 {
			hasFilters = 1
			value := `{"city":"` + string(cityEqF) + `"}`
			resultIds = processResults(
				eqFilter("city", value, tx),
				resultIds,
			)
		}
		if len(cityAnyF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				anyFilter("city", string(cityAnyF), tx),
				resultIds,
			)
		}
		if len(cityNullF) > 0 {
			hasFilters = 1
			value := string(cityNullF)
			if value == "0" {
				responseProperties = append(responseProperties, "city")
			}
			resultIds = processResults(
				nullFilter("city", value, tx),
				resultIds,
			)
		}
		if len(birthLtF) > 0 {
			hasFilters = 1
			value := `{"birth":"` + string(birthLtF) + `"}`
			resultIds = processResults(
				ltFilter("birth", value, tx),
				resultIds,
			)
		}
		if len(birthGtF) > 0 {
			hasFilters = 1
			value := `{"birth":"` + string(birthGtF) + `"}`
			resultIds = processResults(
				gtFilter("birth", value, tx),
				resultIds,
			)
		}
		if len(birthYearF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				eqFilter("birth_year", string(birthYearF), tx),
				resultIds,
			)
		}
		if len(interestsContainsF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				containsFilter("interests", string(interestsContainsF), tx),
				resultIds,
			)
		}
		if len(interestsAnyF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				anyFilter("interests", string(interestsAnyF), tx),
				resultIds,
			)
		}
		if len(likesContainsF) > 0 {
			hasFilters = 1
			resultIds = processResults(
				containsFilter("likes", string(likesContainsF), tx),
				resultIds,
			)
		}
		if string(premiumNowF) == "1" {
			hasFilters = 1
			resultIds = processResults(
				ltFilter("premium_to", fmt.Sprintf("%v", now), tx),
				resultIds,
			)
		}
		if len(premiumNullF) > 0 {
			hasFilters = 1
			value := string(premiumNullF)
			if value == "0" {
				responseProperties = append(responseProperties, "premium")
			}
			resultIds = processResults(
				nullFilter("premium", value, tx),
				resultIds,
			)
		}
		if hasFilters == 0 {
			_ = tx.Descend("id", func(key, val string) bool {
				// todo: use val?
				resultIds = append(resultIds, GetIdFromKey(key))
				return len(resultIds) < limit
			})
		}

		return nil
	})

	// todo: apply unique for slice

	// order by ID desc
	// apply limit
	sort.Sort(sort.Reverse(sort.IntSlice(resultIds)))
	if len(resultIds) > 0 && limit > 0 && len(resultIds) > limit {
		resultIds = resultIds[0:limit]
	}

	var results = make([]Account, 0)
	for _, id := range resultIds {
		parsed := gjson.ParseBytes(GetAccount(int64(id)))
		resultMap := parsed.Map()

		result := make(Account, 0)
		for _, key := range responseProperties {
			result[key] = resultMap[key].Value()
		}
		results = append(results, result)
	}

	response := &FilterResponse{
		Accounts: results,
	}
	jsonData, _ := json.Marshal(response)

	ctx.Success("application/json; charset=utf-8", jsonData)
	return
}

func eqFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	_ = tx.AscendEqual(fKey, fVal, func(key, val string) bool {
		resultIds = append(resultIds, GetIdFromKey(key))
		return true
	})

	return resultIds
}

func neqFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	_ = tx.Ascend(fKey, func(key, val string) bool {
		value := gjson.Parse(val).Get(fKey)

		if value.String() != fVal {
			resultIds = append(resultIds, GetIdFromKey(key))
		}

		return true
	})

	return resultIds
}

func ltFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	_ = tx.AscendLessThan(fKey, fVal, func(key, val string) bool {
		resultIds = append(resultIds, GetIdFromKey(key))
		return true //TODO: get not more than limit
	})

	return resultIds
}

func gtFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	_ = tx.AscendGreaterOrEqual(fKey, fVal, func(key, val string) bool {
		resultIds = append(resultIds, GetIdFromKey(key))
		return true
	})

	return resultIds
}

func nullFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
	resultIds := make([]int, 0)

	if fVal == "0" {
		_ = tx.Descend(fKey, func(key, val string) bool {
			value := gjson.Parse(val).Get(fKey)

			isNotEmpty := value.Exists() || value.String() != ""

			if isNotEmpty {
				resultIds = append(resultIds, GetIdFromKey(key))
			}

			return isNotEmpty
		})
	}

	if fVal == "1" {
		_ = tx.Ascend(fKey, func(key, val string) bool {
			value := gjson.Parse(val).Get(fKey)

			isEmpty := !value.Exists() || value.String() == ""

			if isEmpty {
				resultIds = append(resultIds, GetIdFromKey(key))
			}

			return isEmpty
		})
	}

	return resultIds
}

func anyFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)

	_ = strings.Split(fVal, ",")

	//if valid {
	//	resultIds = append(resultIds, GetIdFromKey(key))
	//}

	return resultIds
}

func containsFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
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
