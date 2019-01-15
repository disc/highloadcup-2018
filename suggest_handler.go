package main

import (
	"strconv"

	"github.com/emirpasic/gods/lists/arraylist"

	"github.com/emirpasic/gods/maps/treemap"

	"github.com/valyala/fasthttp"
)

func suggestHandler(ctx *fasthttp.RequestCtx, accountId int) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"country": 1, "city": 1,
	}

	var requestedAccount *Account
	if account, ok := accountIndex.Get(accountId); ok {
		requestedAccount = account.(*Account)
	} else {
		ctx.Error("{}", 404)
		return
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
	if limit, err = strconv.Atoi(string(ctx.QueryArgs().Peek("limit"))); err != nil || limit <= 0 {
		ctx.Error("{}", 400)
		return
	}
	// Limit is required
	var countryEqF []byte
	var cityEqF []byte
	if ctx.QueryArgs().Has("country") {
		countryEqF = ctx.QueryArgs().Peek("country")
		if len(countryEqF) == 0 {
			ctx.Error("{}", 400)
			return
		}
	}
	if ctx.QueryArgs().Has("city") {
		cityEqF = ctx.QueryArgs().Peek("city")
		if len(cityEqF) == 0 {
			ctx.Error("{}", 400)
			return
		}
	}

	index := calculateSimilarityForUser(requestedAccount)
	if index == nil || index.Size() == 0 {
		emptyResponse(ctx)
		return
	}

	foundAccounts := arraylist.New()

	filters := make(map[string]interface{})

	var countryEqFilter string
	if len(countryEqF) > 0 {
		countryEqFilter = string(countryEqF)
		filters["country"] = 1
	}
	var cityEqFilter string
	if len(cityEqF) > 0 {
		cityEqFilter = string(cityEqF)
		filters["city"] = 1
	}

	filtersCount := len(filters)

	it := index.Iterator()
	for it.Next() {
		if foundAccounts.Size() >= limit {
			break
		}
		passedFilters := 0
		account := it.Value().(*Account)

		if countryEqFilter != "" {
			if account.Country == countryEqFilter {
				passedFilters += 1
			} else {
				continue
			}
		}

		if cityEqFilter != "" {
			if account.City == cityEqFilter {
				passedFilters += 1
			} else {
				continue
			}
		}

		if passedFilters == filtersCount {
			suggestsByOneUser := treemap.NewWith(inverseUint32Comparator)
			for likeId := range account.likes {
				// ignore exists like
				if _, exists := requestedAccount.likes[likeId]; exists {
					continue
				}
				if suggestedLike, ok := accountIndex.Get(likeId); ok {
					suggestedAccount := suggestedLike.(*Account)
					if suggestedAccount.Sex != requestedAccount.Sex {
						// sort by like id from one user
						suggestsByOneUser.Put(suggestedAccount.ID, suggestedAccount)
					}
				}
			}
			if suggestsByOneUser.Size() > 0 {
				foundAccounts.Add(suggestsByOneUser.Values()...)
			}
		}
	}

	if foundAccounts.Size() > 0 {
		var found []*Account
		it := foundAccounts.Iterator()
		for it.Next() && len(found) < limit {
			found = append(found, it.Value().(*Account))
		}

		ctx.Success("application/json", prepareResponseBytes(found, []string{
			"id", "email", "status", "fname", "sname",
		}))
		return
	}

	emptyResponse(ctx)
	return
}
