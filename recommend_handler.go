package main

import (
	"math"
	"strconv"

	"github.com/emirpasic/gods/lists/arraylist"

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

	foundAccounts := arraylist.New()

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
		//if foundAccounts.Size() >= limit {
		//	break
		//}
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

		interestsIntersections := intersectionsCount(requestedAccount.interestsMap, account.interestsMap)
		if interestsIntersections > 0 {
			passedFilters += 1
		} else {
			continue
		}

		// WHERE commonInterests>0 ORDER BY premium_now, status, commonInterests, ageDiffSeconds

		if passedFilters == filtersCount {
			foundAccounts.Add(&CompatibilityResult{
				id:              account.ID,
				hasPremiumNow:   account.hasActivePremium(now),
				status:          account.Status,
				commonInterests: interestsIntersections,
				ageDiff:         math.Abs(float64(requestedAccount.Birth - account.Birth)),
				account:         account,
			})
		}
	}

	if foundAccounts.Size() > 0 {
		foundAccounts.Sort(inverseCompatibilityComparator)
		var found []*Account
		it := foundAccounts.Iterator()
		for it.Next() && len(found) < limit {
			found = append(found, it.Value().(*CompatibilityResult).account)
		}

		ctx.Success("application/json", prepareResponseBytes(found, []string{
			"id", "email", "status", "fname", "sname", "birth", "premium",
		}))
		return
	}

	emptyResponse(ctx)
	return
}

type CompatibilityResult struct {
	id              int
	hasPremiumNow   bool
	status          string
	commonInterests int
	ageDiff         float64
	account         *Account
}

var inverseCompatibilityComparator = func(a, b interface{}) int {
	return -compatibilityComparator(a, b)
}

// Custom comparator (sort by IDs)
// Should return a number:
//    negative , if a < b
//    zero     , if a == b
//    positive , if a > b
func compatibilityComparator(a, b interface{}) int {

	// WHERE commonInterests>0 ORDER BY premium_now desc, status enum, commonInterests desc, ageDiffSeconds asc

	// Type assertion, program will panic if this is not respected
	acc1 := a.(*CompatibilityResult)
	acc2 := b.(*CompatibilityResult)

	switch {
	case acc1.hasPremiumNow == acc2.hasPremiumNow && acc1.hasPremiumNow && acc2.hasPremiumNow:
		return 0
	case acc1.hasPremiumNow != acc2.hasPremiumNow:
		//if acc1.hasPremiumNow && acc2.hasPremiumNow {
		//	return 0
		//}
		if acc1.hasPremiumNow {
			return -1
		} else {
			return 1
		}
	case acc1.status != acc2.status:
		if acc1.status == "свободны" {
			return -1
		}
		if acc2.status == "свободны" {
			return 1
		}
		if acc1.status == "всё сложно" {
			return -1
		}
		if acc2.status == "всё сложно" {
			return 1
		}
		if acc1.status == "заняты" {
			return -1
		}
		if acc2.status == "заняты" {
			return 1
		}
		return 0
	case acc1.commonInterests != acc2.commonInterests:
		if acc1.commonInterests > acc2.commonInterests {
			return -1
		} else {
			return 1
		}
	case acc1.ageDiff != acc2.ageDiff:
		if acc1.ageDiff < acc2.ageDiff {
			return -1
		} else {
			return 1
		}
	default:
		if acc1.id < acc2.id {
			return -1
		} else {
			return 1
		}
	}
}
