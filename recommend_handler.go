package main

import (
	"math"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"

	"github.com/valyala/fasthttp"
)

func recommendHandler(ctx *fasthttp.RequestCtx, accountId int) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"country": 1, "city": 1,
	}

	var requestedAccount *Account
	if account, ok := accountMap.Get(accountId); ok {
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

	var index *treemap.Map

	vnidxpool := namedIndexPool.Get()
	namedIndex := vnidxpool.(*NamedIndex)

	vmap := treemapPool.Get()
	suitableIndexes := vmap.(*treemap.Map)
	suitableIndexes.Put(accountMap.Size(), namedIndex.Update([]byte("default"), accountMap))

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

	foundAccounts := treemap.NewWith(inverseFloat32Comparator)

	filters := make(map[string]interface{})

	filters["compatibility"] = 1

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

	if suitableIndexes.Size() > 0 {
		if _, shortestIndex := suitableIndexes.Min(); &shortestIndex != nil {
			res := shortestIndex.(*NamedIndex)
			index = res.index
		}
	}

	namedIndexPool.Put(vnidxpool)
	treemapPool.Put(vmap)

	if index == nil || index.Size() == 0 {
		emptyResponse(ctx)
		return
	}

	filtersCount := len(filters)

	it := index.Iterator()
	for it.Next() {
		if foundAccounts.Size() >= limit {
			break
		}
		passedFilters := 0
		account := it.Value().(*Account)

		if requestedAccount.Sex == account.Sex {
			continue
		}

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

		compatibility := calculateCompatibility(requestedAccount, account)
		if compatibility > 0 {
			passedFilters += 1
		} else {
			continue
		}

		if passedFilters == filtersCount {
			foundAccounts.Put(compatibility, account)
		}
	}

	if foundAccounts.Size() > 0 {
		var found []*Account
		it := foundAccounts.Iterator()
		for it.Next() && len(found) < limit {
			found = append(found, it.Value().(*Account))
		}

		ctx.Success("application/json", prepareResponseBytes(found, []string{
			"id", "email", "status", "fname", "sname", "birth", "premium",
		}))
		return
	}

	emptyResponse(ctx)
	return
}

func calculateCompatibility(me *Account, somebody *Account) float32 {
	interestsIntersections := intersectionsCount(somebody.interestsMap, me.interestsMap)

	if interestsIntersections == 0 {
		return 0
	}

	var quantity = float32(1.0)
	switch somebody.Status {
	case "свободны":
		quantity *= 1
	case "все сложно":
		quantity *= 0.4
	case "заняты":
		quantity *= 0.1
	}

	// update this
	// max 1
	quantity *= 1 / float32(len(me.interestsMap)/interestsIntersections)

	if me.Birth == somebody.Birth {
		quantity *= 1
	} else {
		tsDiff := float32(math.Abs(float64(me.Birth - somebody.Birth)))

		// age
		quantity *= 1 / tsDiff
	}

	if somebody.hasActivePremium(now) {
		quantity *= 100
	}

	return quantity
}
