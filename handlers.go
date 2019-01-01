package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin/json"
	"github.com/tidwall/buntdb"
	"github.com/valyala/fasthttp"
)

type FilterResponse struct {
	Accounts []AccountAsMap `json:"accounts"`
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
		"email",
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

	statusEqF := ctx.QueryArgs().Peek("email_eq")
	statusNeqF := ctx.QueryArgs().Peek("email_neq")
	if len(statusEqF) > 0 || len(statusNeqF) > 0 {
		responseProperties = append(responseProperties, "status")
	}

	fnameEqF := ctx.QueryArgs().Peek("fname_eq")
	fnameAnyF := ctx.QueryArgs().Peek("fname_any")
	fnameNullF := ctx.QueryArgs().Peek("fname_null")
	if len(fnameEqF) > 0 || len(fnameAnyF) > 0 || len(fnameNullF) > 0 {
		responseProperties = append(responseProperties, "fname")
	}

	snameEqF := ctx.QueryArgs().Peek("sname_eq")
	snameStartsF := ctx.QueryArgs().Peek("sname_starts")
	snameNullF := ctx.QueryArgs().Peek("sname_null")
	if len(snameEqF) > 0 || len(snameStartsF) > 0 || len(snameNullF) > 0 {
		responseProperties = append(responseProperties, "sname")
	}

	phoneCodeF := ctx.QueryArgs().Peek("phone_code")
	phoneNullF := ctx.QueryArgs().Peek("phone_null")
	if len(phoneCodeF) > 0 || len(phoneNullF) > 0 {
		responseProperties = append(responseProperties, "phone")
	}

	countryEqF := ctx.QueryArgs().Peek("country_eq")
	countryNullF := ctx.QueryArgs().Peek("country_null")
	if len(countryEqF) > 0 || len(countryNullF) > 0 {
		responseProperties = append(responseProperties, "country")
	}

	cityEqF := ctx.QueryArgs().Peek("city_eq")
	cityAnyF := ctx.QueryArgs().Peek("city_any")
	cityNullF := ctx.QueryArgs().Peek("city_null")
	if len(cityEqF) > 0 || len(cityAnyF) > 0 || len(cityNullF) > 0 {
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
	premiumNullF := ctx.QueryArgs().Peek("premium_now")
	if len(likesContainsF) > 0 {
		responseProperties = append(responseProperties, "premium")
	}

	var resultIds []int
	var results = make([]AccountAsMap, 0)

	hasFilters := 0
	// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
	_ = db.View(func(tx *buntdb.Tx) error {
		if len(sexEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("sex", string(sexEqF), tx)
		}
		if len(emailDomainF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("email_domain", string(emailDomainF), tx)
		}
		if len(emailLtF) > 0 {
			hasFilters = 1
			resultIds = ltFilter("email", string(emailLtF), tx)
		}
		if len(emailGtF) > 0 {
			hasFilters = 1
			resultIds = gtFilter("email", string(emailGtF), tx)
		}
		if len(statusEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("status", string(statusEqF), tx)
		}
		if len(statusNeqF) > 0 {
			hasFilters = 1
			_ = tx.Ascend("status", func(key, val string) bool {
				// TODO: Rewrite
				if val != string(statusEqF) {
					resultIds = append(resultIds, GetIdFromKey(key))
				}
				return true
			})
		}
		if len(fnameEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("fname", string(fnameEqF), tx)
		}
		if len(fnameAnyF) > 0 {
			hasFilters = 1
			anyFilter("fname", string(fnameAnyF), tx)
		}
		if len(fnameNullF) > 0 {
			hasFilters = 1
			// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
			resultIds = nullFilter("fname", string(fnameNullF), tx)
		}
		if len(snameEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("sname", string(snameEqF), tx)
		}
		if len(snameStartsF) > 0 {
			hasFilters = 1
			sSnameStartsF := string(snameStartsF)
			_ = tx.Ascend("sname", func(key, val string) bool {
				if strings.HasPrefix(val, sSnameStartsF) {
					resultIds = append(resultIds, GetIdFromKey(key))
				}

				return true
			})
		}
		if len(snameNullF) > 0 {
			hasFilters = 1
			resultIds = nullFilter("sname", string(snameNullF), tx)
		}
		if len(phoneCodeF) > 0 {
			hasFilters = 1
			//TODO: create phone_code index
			resultIds = eqFilter("phone_code", string(phoneCodeF), tx)
		}
		if len(phoneNullF) > 0 {
			hasFilters = 1
			resultIds = nullFilter("phone", string(phoneNullF), tx)
		}
		if len(countryEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("country", string(countryEqF), tx)
		}
		if len(countryNullF) > 0 {
			hasFilters = 1
			resultIds = nullFilter("country", string(countryNullF), tx)
		}
		if len(cityEqF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("city", string(cityEqF), tx)
		}
		if len(cityAnyF) > 0 {
			hasFilters = 1
			resultIds = anyFilter("city", string(cityAnyF), tx)
		}
		if len(cityNullF) > 0 {
			hasFilters = 1
			resultIds = nullFilter("city", string(cityNullF), tx)
		}
		if len(birthLtF) > 0 {
			hasFilters = 1
			resultIds = ltFilter("birth", string(birthLtF), tx)
		}
		if len(birthGtF) > 0 {
			hasFilters = 1
			resultIds = gtFilter("birth", string(birthGtF), tx)
		}
		if len(birthYearF) > 0 {
			hasFilters = 1
			resultIds = eqFilter("birth_year", string(statusEqF), tx)
		}
		if len(interestsContainsF) > 0 {
			hasFilters = 1
			resultIds = containsFilter("interests", string(interestsContainsF), tx)
		}
		if len(interestsAnyF) > 0 {
			hasFilters = 1
			resultIds = anyFilter("interests", string(interestsAnyF), tx)
		}
		if len(likesContainsF) > 0 {
			hasFilters = 1
			resultIds = containsFilter("likes", string(likesContainsF), tx)
		}
		if len(premiumNowF) > 0 {
			hasFilters = 1
			resultIds = ltFilter("premium_to", string(premiumNowF), tx)
		}
		if len(premiumNullF) > 0 {
			hasFilters = 1
			resultIds = nullFilter("premium", string(premiumNullF), tx)
		}
		if hasFilters == 0 {
			_ = tx.Descend("id", func(key, val string) bool {
				// todo: use val?
				resultIds = append(resultIds, GetIdFromKey(key))
				return len(resultIds) < limit
			})
		}

		// TODO: intersect results after each filter

		return nil
	})

	// todo: apply unique for slice

	// order by ID desc
	// apply limit
	sort.Sort(sort.Reverse(sort.IntSlice(resultIds)))
	if len(resultIds) > 0 && len(resultIds) > limit {
		resultIds = resultIds[0:limit]
	}

	for _, id := range resultIds {
		results = append(results, GetAccount(id, responseProperties))
	}

	response := &FilterResponse{
		Accounts: results,
	}

	jsonData, _ := json.Marshal(response)

	ctx.Success("application/json", jsonData)
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

func ltFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	_ = tx.AscendLessThan(fKey, fVal, func(key, val string) bool {
		resultIds = append(resultIds, GetIdFromKey(key))
		return true
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

func anyFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)
	fMap := make(map[string]int, 0)
	if len(fVal) > 0 {
		for _, fnameVal := range strings.Split(fVal, ",") {
			fMap[string(fnameVal)] = 1
		}
	}
	_ = tx.Ascend(fKey, func(key, val string) bool {
		if _, ok := fMap[val]; ok {
			resultIds = append(resultIds, GetIdFromKey(key))
		}
		return true
	})

	return resultIds
}

func nullFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
	resultIds := make([]int, 0)
	_ = tx.Ascend(fKey, func(key, val string) bool {
		if fVal == "0" && len(val) > 0 {
			resultIds = append(resultIds, GetIdFromKey(key))
			return true
		}
		if fVal == "1" && len(val) == 0 {
			resultIds = append(resultIds, GetIdFromKey(key))
			return len(val) == 0
		}
		return true
	})

	return resultIds
}

func containsFilter(fKey string, fVal string, tx *buntdb.Tx) []int {
	resultIds := make([]int, 0)

	origin := strings.Split(fVal, ",")

	_ = tx.Ascend(fKey, func(key, val string) bool {
		result := sliceIntersection(origin, strings.Split(val, ","))
		if len(result) > 0 {
			resultIds = append(resultIds, GetIdFromKey(key))
		}
		return true
	})

	return resultIds
}

func sliceIntersection(a, b []string) (c []string) {
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
