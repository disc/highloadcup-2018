package main

import (
	"bytes"
	"sort"
	"strconv"

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
	fnameMap := map[string]int{}
	if len(fnameEqF) > 0 || len(fnameAnyF) > 0 || len(fnameNullF) > 0 {
		responseProperties = append(responseProperties, "fname")
		if len(fnameAnyF) > 0 {
			for _, fnameVal := range bytes.Split(fnameAnyF, []byte{','}) {
				fnameMap[string(fnameVal)] = 1
			}
		}
	}

	snameEqF := ctx.QueryArgs().Peek("sname_eq")
	snameStartsF := ctx.QueryArgs().Peek("sname_starts")
	snameNullF := ctx.QueryArgs().Peek("sname_null")
	if len(snameEqF) > 0 || len(snameStartsF) > 0 || len(snameNullF) > 0 {
		responseProperties = append(responseProperties, "sname")
	}

	var resultIds []int
	var results = make([]AccountAsMap, 0)

	hasFilters := 0
	// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
	_ = db.View(func(tx *buntdb.Tx) error {
		if len(sexEqF) > 0 {
			hasFilters = 1
			_ = tx.AscendEqual("sex", string(sexEqF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(emailDomainF) > 0 {
			hasFilters = 1
			_ = tx.AscendEqual("email_domain", string(emailDomainF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(emailLtF) > 0 {
			hasFilters = 1
			_ = tx.AscendLessThan("email", string(emailLtF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(emailGtF) > 0 {
			hasFilters = 1
			_ = tx.AscendGreaterOrEqual("email", string(emailGtF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(statusEqF) > 0 {
			hasFilters = 1
			_ = tx.AscendEqual("status", string(statusEqF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(statusNeqF) > 0 {
			hasFilters = 1
			_ = tx.Ascend("status", func(key, val string) bool {
				if val != string(statusEqF) {
					resultIds = append(resultIds, GetIdFromKey(key))
				}
				return true
			})
		}
		if len(fnameEqF) > 0 {
			hasFilters = 1
			_ = tx.AscendEqual("fname", string(fnameEqF), func(key, val string) bool {
				resultIds = append(resultIds, GetIdFromKey(key))
				return true
			})
		}
		if len(fnameAnyF) > 0 {
			hasFilters = 1
			_ = tx.Ascend("fname", func(key, val string) bool {
				if _, ok := fnameMap[val]; ok {
					resultIds = append(resultIds, GetIdFromKey(key))
				}
				return true
			})
		}
		if len(fnameAnyF) > 0 {
			hasFilters = 1
			_ = tx.Ascend("fname", func(key, val string) bool {
				if _, ok := fnameMap[val]; ok {
					resultIds = append(resultIds, GetIdFromKey(key))
				}
				return true
			})
		}
		if len(fnameNullF) > 0 {
			hasFilters = 1
			// null - выбрать всех, у кого указано имя (если 0) или не указано (если 1);
			sFnameNullF := string(fnameNullF)
			_ = tx.Ascend("fname", func(key, val string) bool {
				if sFnameNullF == "0" && len(val) > 0 {
					resultIds = append(resultIds, GetIdFromKey(key))
					return true
				}
				if sFnameNullF == "1" && len(val) == 0 {
					resultIds = append(resultIds, GetIdFromKey(key))
					return len(val) == 0
				}
				return true
			})
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

	// order desc
	// limit
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
