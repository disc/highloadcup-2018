package main

import (
	"encoding/json"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"

	"github.com/valyala/fasthttp"
)

func suggestHandler(ctx *fasthttp.RequestCtx, accountId int) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"country": 1, "city": 1,
	}

	// check if account exists
	index, found := similarityMap[accountId]
	if !found {
		ctx.Error("{}", 404)
		return
	}
	var requestedAccount *Account
	if account, ok := accountMap.Get(accountId); ok {
		requestedAccount = account.(*Account)
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

	var limit int
	var err error
	if limit, err = strconv.Atoi(string(ctx.QueryArgs().Peek("limit"))); err != nil {
		ctx.Error("{}", 400)
		return
	}
	if limit <= 0 {
		emptyFilterResponse(ctx)
		return
	}
	// Limit is required

	countryEqF := ctx.QueryArgs().Peek("country")
	cityEqF := ctx.QueryArgs().Peek("city_eq")

	foundAccounts := treemap.NewWith(inverseIntComparator)

	filters := make(map[string]interface{})

	var countryEqFilter string
	if len(countryEqF) > 0 {
		countryEqFilter = string(countryEqF)
		filters["country"] = 1
	}
	var cityEqFilter string
	if len(cityEqF) > 0 {
		cityEqFilter = string(cityEqF)
		filters["city_eq"] = 1
	}

	filtersCount := len(filters)

	if index != nil {
		it := index.Iterator()
		for it.Next() {
			if foundAccounts.Size() >= limit {
				break
			}
			passedFilters := 0
			account := *it.Value().(*Account)
			value := account.record

			if countryEqFilter != "" {
				if value["country"].String() == countryEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}

			if cityEqFilter != "" {
				if value["city"].String() == cityEqFilter {
					passedFilters += 1
				} else {
					continue
				}
			}

			if passedFilters == filtersCount {
				for likeId := range account.likes {
					if _, exists := requestedAccount.likes[likeId]; exists {
						continue
					}
					//if foundAccounts.Size() >= limit {
					//	break
					//}
					if suggestedLike, ok := accountMap.Get(likeId); ok {
						suggestedAccount := suggestedLike.(*Account)
						if suggestedAccount.sex != requestedAccount.sex {
							foundAccounts.Put(suggestedAccount.id, suggestedAccount)
						}
					}
				}
			}
		}
	}

	jsonData := []byte(`{"accounts":[]}`)
	if foundAccounts.Size() > 0 {
		jsonData, _ = json.Marshal(prepareSuggestResponse(foundAccounts, limit))
	}

	// TODO: Use sjson for updates
	ctx.Success("application/json", jsonData)
	return
}

func prepareSuggestResponse(found *treemap.Map, limit int) *FilterResponse {
	// ignore interests, likes
	responseProperties := []string{
		"id", "email", "status", "fname", "sname",
	}
	var results = make([]AccountResponse, 0)
	it := found.Iterator()
	for it.Next() && len(results) < limit {
		account := it.Value().(*Account)
		result := AccountResponse{}
		for _, key := range responseProperties {
			if account.record[key].Exists() {
				result[key] = account.record[key].Value()
			}
		}
		results = append(results, result)
	}

	return &FilterResponse{
		Accounts: results,
	}
}
