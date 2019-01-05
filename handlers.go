package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

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
	if len(interestsContainsF) > 0 || len(interestsAnyF) > 0 {
		//responseProperties = append(responseProperties, "interests")
	}

	//
	//likesContainsF := ctx.QueryArgs().Peek("likes_contains")
	//
	//premiumNowF := ctx.QueryArgs().Peek("premium_now")
	//premiumNullF := ctx.QueryArgs().Peek("premium_null")
	//if len(premiumNowF) > 0 {
	//	responseProperties = append(responseProperties, "premium")
	//}

	var foundAccounts []*Account

	filters := make(map[string]interface{})

	var sexEqFilter string
	if len(sexEqF) > 0 {
		sexEqFilter = string(sexEqF)
		filters["sex_eq"] = 1
	}
	var emailDomainFilter string
	if len(emailDomainF) > 0 {
		emailDomainFilter = string(emailDomainF)
		filters["email_domain"] = 1
	}
	var emailLtFilter []byte
	if len(emailLtF) > 0 {
		emailLtFilter = emailLtF
		filters["email_lt"] = 1
	}
	var emailGtFilter []byte
	if len(emailGtF) > 0 {
		emailGtFilter = emailGtF
		filters["email_gt"] = 1
	}
	var statusEqFilter string
	if len(statusEqF) > 0 {
		statusEqFilter = string(statusEqF)
		filters["status_eq"] = 1
	}
	var statusNeqFilter string
	if len(statusNeqF) > 0 {
		statusNeqFilter = string(statusNeqF)
		filters["status_neq"] = 1
	}
	var fnameNullFilter bool
	var fnameNotNullFilter bool
	if len(fnameNullF) > 0 {
		if string(fnameNullF) == "0" {
			fnameNotNullFilter = true
			filters["fname_not_null"] = 1
		} else {
			fnameNullFilter = true
			filters["fname_null"] = 1
		}

	}
	var interestsAnyFilter []string
	//var interestsContainsFilter []string
	if len(interestsAnyF) > 0 {
		words := strings.Split(string(interestsAnyF), ",")
		if len(words) > 0 {
			filters["interests_any"] = 1
			interestsAnyFilter = words
		}
	}
	//if len(interestsContainsF) > 0 {
	//	words := strings.Split(string(interestsContainsF), ",")
	//	if len(words) > 0 {
	//		filters["interests_contains"] = 1
	//		interestsContainsFilter = words
	//	}
	//}
	filtersCount := len(filters)
	if filtersCount == 0 {
		//foundAccounts = append(foundAccounts, GetIdFromKey(key))
		//return len(foundAccounts) < limit
	}

	// full scan search
	it := accountMap.Iterator()
	for it.Next() {
		if len(foundAccounts) >= limit {
			break
		}
		passedFilters := 0
		account := *it.Value().(*Account)
		value := account.record
		if sexEqFilter != "" {
			if value["sex"].Value() == sexEqFilter {
				passedFilters += 1
			} else {
				continue
			}
		}
		if statusEqFilter != "" {
			if value["status"].Value() == statusEqFilter {
				passedFilters += 1
			} else {
				continue
			}
		}
		if statusNeqFilter != "" {
			if value["status"].Value() != statusNeqFilter {
				passedFilters += 1
			} else {
				continue
			}
		}
		if fnameNullFilter {
			if value["fname"].String() == "" {
				passedFilters += 1
			} else {
				continue
			}
		} else if fnameNotNullFilter {
			if value["fname"].String() != "" {
				passedFilters += 1
			} else {
				continue
			}
		}
		if len(interestsAnyFilter) > 0 {
			//start := time.Now()
			for _, v := range interestsAnyFilter {
				if account.interestsTree.HasKeysWithPrefix(v) {
					passedFilters += 1
					break
				}
			}
			//log.Printf("contains took %s", time.Since(start))
		}
		if len(emailLtFilter) > 0 {
			if bytes.Compare(account.emailBytes, emailLtFilter) < 0 {
				passedFilters += 1
			} else {
				continue
			}
		} else if len(emailGtFilter) > 0 {
			if bytes.Compare(account.emailBytes, emailGtFilter) > 0 {
				passedFilters += 1
			} else {
				continue
			}
		}
		if emailDomainFilter != "" {
			if account.emailDomain == emailDomainFilter {
				passedFilters += 1
			} else {
				continue
			}
		}
		if passedFilters == filtersCount {
			foundAccounts = append(foundAccounts, &account)
		}
	}

	// index search
	//if emailLtFilter, ok := filters["email_lt"].(string); ok && value["email"].String()[0:len(emailLtFilter)] < emailLtFilter {
	//	passedFilters += 1
	//}

	// order by ID desc
	// apply limit

	jsonData := []byte(`{"accounts":[]}`)
	if len(foundAccounts) > 0 {
		jsonData, _ = json.Marshal(prepareResponse(foundAccounts, responseProperties))
	}

	// TODO: Use sjson for updates
	ctx.Success("application/json", jsonData)
	return
}

func prepareResponse(found []*Account, responseProperties []string) *FilterResponse {
	var results = make([]AccountResponse, 0)
	for _, account := range found {
		result := AccountResponse{}
		for _, key := range responseProperties {
			result[key] = account.record[key].Value()
		}
		results = append(results, result)
	}

	return &FilterResponse{
		Accounts: results,
	}
}
