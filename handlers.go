package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/maps/treemap"

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

	limit, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("limit")))
	// Limit is required
	if limit <= 0 {
		emptyFilterResponse(ctx)
		return
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
	snameEqF := ctx.QueryArgs().Peek("sname_eq")
	snameStartsF := ctx.QueryArgs().Peek("sname_starts")
	snameNullF := ctx.QueryArgs().Peek("sname_null")
	if len(snameEqF) > 0 || len(snameStartsF) > 0 {
		responseProperties = append(responseProperties, "sname")
	}
	//
	phoneCodeF := ctx.QueryArgs().Peek("phone_code")
	phoneNullF := ctx.QueryArgs().Peek("phone_null")
	if len(phoneCodeF) > 0 {
		responseProperties = append(responseProperties, "phone")
	}
	//
	countryEqF := ctx.QueryArgs().Peek("country_eq")
	countryNullF := ctx.QueryArgs().Peek("country_null")
	if len(countryEqF) > 0 {
		responseProperties = append(responseProperties, "country")
	}
	//
	cityEqF := ctx.QueryArgs().Peek("city_eq")
	cityAnyF := ctx.QueryArgs().Peek("city_any")
	cityNullF := ctx.QueryArgs().Peek("city_null")
	if len(cityEqF) > 0 || len(cityAnyF) > 0 {
		responseProperties = append(responseProperties, "city")
	}
	//
	birthLtF := ctx.QueryArgs().Peek("birth_lt")
	birthGtF := ctx.QueryArgs().Peek("birth_gt")
	birthYearF := ctx.QueryArgs().Peek("birth_year")
	if len(birthLtF) > 0 || len(birthGtF) > 0 || len(birthYearF) > 0 {
		responseProperties = append(responseProperties, "birth")
	}
	//
	interestsContainsF := ctx.QueryArgs().Peek("interests_contains")
	interestsAnyF := ctx.QueryArgs().Peek("interests_any")

	//
	//likesContainsF := ctx.QueryArgs().Peek("likes_contains")
	//
	premiumNowF := ctx.QueryArgs().Peek("premium_now")
	premiumNullF := ctx.QueryArgs().Peek("premium_null")
	if len(premiumNowF) > 0 {
		responseProperties = append(responseProperties, "premium")
	}

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
	var birthYearFilter int64
	if len(birthYearF) > 0 {
		birthYearFilter, _ = strconv.ParseInt(string(birthYearF), 10, 32)
		filters["birth_year"] = 1
	}
	var birthLtFilter int64
	if len(birthLtF) > 0 {
		birthLtFilter, _ = strconv.ParseInt(string(birthLtF), 10, 32)
		filters["birth_lt"] = 1
	}
	var birthGtFilter int64
	if len(birthGtF) > 0 {
		birthGtFilter, _ = strconv.ParseInt(string(birthGtF), 10, 32)
		filters["birth_gt"] = 1
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
	var fnameEqFilter string
	var fnameAnyFilter = make(map[string]int, 0)
	if len(fnameEqF) > 0 {
		fnameEqFilter = string(fnameEqF)
		filters["fname_eq"] = 1
	}
	if len(fnameNullF) > 0 {
		if string(fnameNullF) == "0" {
			fnameNotNullFilter = true
			filters["fname_not_null"] = 1
		} else {
			fnameNullFilter = true
			filters["fname_null"] = 1
		}

	}
	if len(fnameAnyF) > 0 {
		words := strings.Split(string(fnameAnyF), ",")
		for _, word := range words {
			fnameAnyFilter[word] = 1
		}
		filters["fname_any"] = 1
	}
	var snameNullFilter bool
	var snameNotNullFilter bool
	var snameEqFilter string
	var snameStartsFilter string
	if len(snameEqF) > 0 {
		snameEqFilter = string(snameEqF)
		filters["sname_eq"] = 1
	}
	if len(snameStartsF) > 0 {
		snameStartsFilter = string(snameStartsF)
		filters["sname_starts"] = 1
	}
	if len(snameNullF) > 0 {
		if string(snameNullF) == "0" {
			snameNotNullFilter = true
			filters["sname_not_null"] = 1
		} else {
			snameNullFilter = true
			filters["sname_null"] = 1
		}

	}
	var phoneNullFilter bool
	var phoneNotNullFilter bool
	var phoneCodeFilter int
	if len(phoneNullF) > 0 {
		if string(phoneNullF) == "0" {
			phoneNotNullFilter = true
			filters["phone_not_null"] = 1
		} else {
			phoneNullFilter = true
			filters["phone_null"] = 1
		}

	}
	if len(phoneCodeF) > 0 {
		phoneCodeFilter, _ = strconv.Atoi(string(phoneCodeF))
		filters["phone_code"] = 1
	}

	var countryEqFilter string
	var countryNullFilter bool
	var countryNotNullFilter bool
	if len(countryEqF) > 0 {
		countryEqFilter = string(countryEqF)
		filters["country_eq"] = 1
	}
	if len(countryNullF) > 0 {
		if string(countryNullF) == "0" {
			countryNotNullFilter = true
			filters["country_not_null"] = 1
		} else {
			countryNullFilter = true
			filters["country_null"] = 1
		}

	}
	var cityEqFilter string
	var cityAnyFilter = make(map[string]int, 0)
	var cityNullFilter bool
	var cityNotNullFilter bool
	if len(cityEqF) > 0 {
		cityEqFilter = string(cityEqF)
		filters["city_eq"] = 1
	}
	if len(cityAnyF) > 0 {
		words := strings.Split(string(cityAnyF), ",")
		for _, word := range words {
			cityAnyFilter[word] = 1
		}
		filters["city_any"] = 1
	}
	if len(cityNullF) > 0 {
		if string(cityNullF) == "0" {
			cityNotNullFilter = true
			filters["city_not_null"] = 1
		} else {
			cityNullFilter = true
			filters["city_null"] = 1
		}

	}
	var premiumNullFilter bool
	var premiumNotNullFilter bool
	var premiumNowFilter bool
	if len(premiumNullF) > 0 {
		if string(premiumNullF) == "0" {
			premiumNotNullFilter = true
			filters["premium_not_null"] = 1
		} else {
			premiumNullFilter = true
			filters["premium_null"] = 1
		}

	}
	if bytes.Equal(premiumNowF, []byte("1")) {
		premiumNowFilter = true
		filters["premium_now"] = 1
	}
	var interestsAnyFilter []string
	var interestsContainsFilter []string
	if len(interestsAnyF) > 0 {
		words := strings.Split(string(interestsAnyF), ",")
		if len(words) > 0 {
			interestsAnyFilter = words
			filters["interests_any"] = 1
		}
	}
	if len(interestsContainsF) > 0 {
		words := strings.Split(string(interestsContainsF), ",")
		if len(words) > 0 {
			interestsContainsFilter = words
			filters["interests_contains"] = 1
		}
	}

	var index *treemap.Map

	type namedIndex struct {
		name  string
		index *treemap.Map
	}

	suitableIndexes := treemap.NewWithIntComparator()
	suitableIndexes.Put(accountMap.Size(), namedIndex{"default", accountMap})

	if countryEqFilter != "" {
		if countryMap[countryEqFilter] != nil && countryMap[countryEqFilter].Size() > 0 {
			suitableIndexes.Put(
				countryMap[countryEqFilter].Size(),
				namedIndex{"country", countryMap[countryEqFilter]},
			)
		} else {
			// todo: return empty json
			emptyFilterResponse(ctx)
			return
		}
	}

	if cityEqFilter != "" {
		if cityMap[cityEqFilter] != nil && cityMap[cityEqFilter].Size() > 0 {
			suitableIndexes.Put(
				cityMap[cityEqFilter].Size(),
				namedIndex{"city", cityMap[cityEqFilter]},
			)
		} else {
			// todo: return empty json
			emptyFilterResponse(ctx)
			return
		}
	}

	if birthYearFilter > 0 {
		if birthYearMap[birthYearFilter] != nil && birthYearMap[birthYearFilter].Size() > 0 {
			suitableIndexes.Put(
				birthYearMap[birthYearFilter].Size(),
				namedIndex{"birth_year", birthYearMap[birthYearFilter]},
			)
		} else {
			// todo: return empty json
			emptyFilterResponse(ctx)
			return
		}
	}

	if snameEqFilter != "" {
		if snameMap[snameEqFilter] != nil && snameMap[snameEqFilter].Size() > 0 {
			suitableIndexes.Put(
				snameMap[snameEqFilter].Size(),
				namedIndex{"sname", snameMap[snameEqFilter]},
			)
		} else {
			// todo: return empty json
			emptyFilterResponse(ctx)
			return
		}
	}

	if fnameEqFilter != "" {
		if fnameMap[fnameEqFilter] != nil && fnameMap[fnameEqFilter].Size() > 0 {
			suitableIndexes.Put(
				fnameMap[fnameEqFilter].Size(),
				namedIndex{"fname", fnameMap[fnameEqFilter]},
			)
		} else {
			// todo: return empty json
			emptyFilterResponse(ctx)
			return
		}
	}

	var selectedIndexName string
	if suitableIndexes.Size() > 0 {
		if _, shortestIndex := suitableIndexes.Min(); &shortestIndex != nil {
			res := shortestIndex.(namedIndex)
			selectedIndexName = res.name
			index = res.index
		}
	}

	filtersCount := len(filters)

	if index != nil {
		it := index.Iterator()
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
			if len(statusEqFilter) > 0 {
				if value["status"].Value() == statusEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if len(statusNeqFilter) > 0 {
				if value["status"].Value() != statusNeqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if fnameEqFilter != "" {
				// use const for index name
				if selectedIndexName == "fname" || value["fname"].String() == fnameEqFilter {
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
			if len(fnameAnyFilter) > 0 {
				fname := account.record["fname"].String()
				if len(fname) == 0 {
					continue
				}
				if _, ok := fnameAnyFilter[fname]; ok {
					passedFilters += 1
				} else {
					continue
				}
			}
			if snameEqFilter != "" {
				// use const for index name
				if selectedIndexName == "sname" || value["sname"].String() == snameEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			} else if snameStartsFilter != "" {
				// slow
				// use const for index name
				//FIXME: slow solution
				if strings.HasPrefix(value["sname"].String(), snameStartsFilter) {
					passedFilters += 1
				} else {
					continue
				}
			}
			if snameNullFilter {
				if value["sname"].String() == "" {
					passedFilters += 1
				} else {
					continue
				}
			} else if snameNotNullFilter {
				if value["sname"].String() != "" {
					passedFilters += 1
				} else {
					continue
				}
			}
			if phoneNullFilter {
				if value["phone"].String() == "" {
					passedFilters += 1
				} else {
					continue
				}
			} else if phoneNotNullFilter {
				if value["phone"].String() != "" {
					passedFilters += 1
				} else {
					continue
				}
			}
			if phoneCodeFilter > 0 {
				if account.phoneCode == phoneCodeFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if countryEqFilter != "" {
				// use const for index name
				if selectedIndexName == "country" || value["country"].String() == countryEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			//FIXME: group null/not-null filters
			if countryNullFilter {
				if value["country"].String() == "" {
					passedFilters += 1
				} else {
					continue
				}
			} else if countryNotNullFilter {
				if value["country"].String() != "" {
					passedFilters += 1
				} else {
					continue
				}
			}
			if cityEqFilter != "" {
				// use const for index name
				if selectedIndexName == "city" || value["city"].String() == cityEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if cityNullFilter {
				if value["city"].String() == "" {
					passedFilters += 1
				} else {
					continue
				}
			} else if cityNotNullFilter {
				if value["city"].String() != "" {
					passedFilters += 1
				} else {
					continue
				}
			}
			if len(cityAnyFilter) > 0 {
				// FIXME: slow solution
				accountCity := account.record["city"].String()
				if len(accountCity) == 0 {
					continue
				}
				if _, ok := cityAnyFilter[accountCity]; ok {
					passedFilters += 1
				} else {
					continue
				}
			}
			if len(interestsAnyFilter) > 0 {
				for _, v := range interestsAnyFilter {
					if account.interestsTree.HasKeysWithPrefix(v) {
						passedFilters += 1
						break
					}
				}
			} else if len(interestsContainsFilter) > 0 {
				// FIXME: slow solution
				suitable := true
				for _, v := range interestsContainsFilter {
					if !account.interestsTree.HasKeysWithPrefix(v) {
						suitable = false
						break
					}
				}
				if suitable {
					passedFilters += 1
				} else {
					continue
				}
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
			if len(emailDomainFilter) > 0 {
				if account.emailDomain == emailDomainFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if birthYearFilter > 0 {
				// use const for index name
				if selectedIndexName == "birth_year" || account.birthYear == birthYearFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if birthLtFilter > 0 {
				if value["birth"].Int() < birthLtFilter {
					passedFilters += 1
				} else {
					continue
				}
			} else if birthGtFilter > 0 {
				if value["birth"].Int() > birthGtFilter {
					passedFilters += 1
				} else {
					continue
				}
			}
			if premiumNullFilter {
				if !value["premium"].IsObject() {
					passedFilters += 1
				} else {
					continue
				}
			} else if premiumNotNullFilter {
				if value["premium"].IsObject() {
					passedFilters += 1
				} else {
					continue
				}
			}
			if premiumNowFilter {
				if account.premiumFinish >= int64(now) {
					passedFilters += 1
				} else {
					continue
				}
			}
			if passedFilters == filtersCount {
				foundAccounts = append(foundAccounts, &account)
			}
		}
	}

	jsonData := []byte(`{"accounts":[]}`)
	if len(foundAccounts) > 0 {
		jsonData, _ = json.Marshal(prepareResponse(foundAccounts, responseProperties))
	}

	// TODO: Use sjson for updates
	ctx.Success("application/json", jsonData)
	return
}

func emptyFilterResponse(ctx *fasthttp.RequestCtx) {
	ctx.Success("application/json", []byte(`{"accounts":[]}`))
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
