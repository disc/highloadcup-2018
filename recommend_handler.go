package main

import (
	"bytes"
	"math"
	"sort"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"

	"github.com/valyala/fasthttp"
)

func recommendHandler(ctx *fasthttp.RequestCtx, accountId uint32) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"country": 1, "city": 1,
	}

	var requestedAccount *AccountUpdated
	if account, ok := accountIndex.Get(accountId); ok {
		requestedAccount = account.(*AccountUpdated)
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
	switch requestedAccount.Sex {
	case sexDict.GetId("m"):
		femaleIndex := sexIndex.Get(sexDict.GetId("f")).(*treemap.Map)
		suitableIndexes.Put(femaleIndex.Size(), namedIndex.Update([]byte("sex_f"), femaleIndex))
	case sexDict.GetId("f"):
		maleIndex := sexIndex.Get(sexDict.GetId("m")).(*treemap.Map)
		suitableIndexes.Put(maleIndex.Size(), namedIndex.Update([]byte("sex_m"), maleIndex))
	}

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

	var foundAccounts []*CompatibilityResult

	filters := make(map[string]interface{})

	filters["compatibility"] = 1

	var countryEqFilter string
	if len(countryEqF) > 0 {
		countryEqFilter = string(countryEqF)
		filters["country"] = 1
		if countryIndex.Exists(countryEqFilter) {
			currIndex := countryIndex.Get(countryEqFilter)
			suitableIndexes.Put(
				currIndex.Size(),
				namedIndex.Update([]byte("country"), currIndex),
			)
		}
	}
	var cityEqFilter string
	if len(cityEqF) > 0 {
		cityEqFilter = string(cityEqF)
		filters["city"] = 1
		if cityIndex.Exists(cityEqFilter) {
			currIndex := cityIndex.Get(cityEqFilter).(*treemap.Map)
			suitableIndexes.Put(
				currIndex.Size(),
				namedIndex.Update([]byte("city"), currIndex),
			)
		}
	}

	var selectedIndexName []byte
	if suitableIndexes.Size() > 0 {
		if _, shortestIndex := suitableIndexes.Min(); &shortestIndex != nil {
			res := shortestIndex.(*NamedIndex)
			index = res.index
			selectedIndexName = res.name
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
		passedFilters := 0
		account := it.Value().(*AccountUpdated)

		if requestedAccount.Sex == account.Sex {
			continue
		}

		if countryEqFilter != "" {
			if bytes.Equal(selectedIndexName, []byte("country")) || account.Country == countriesDict.GetId(countryEqFilter) {
				passedFilters += 1
			} else {
				continue
			}
		}

		if cityEqFilter != "" {
			if bytes.Equal(selectedIndexName, []byte("city ")) || account.City == citiesDict.GetId(cityEqFilter) {
				passedFilters += 1
			} else {
				continue
			}
		}

		interestsIntersections := intersectionsCount(requestedAccount.InterestsMap, account.InterestsMap)
		if interestsIntersections > 0 {
			passedFilters += 1
		} else {
			continue
		}

		// WHERE commonInterests>0 ORDER BY premium_now, status, commonInterests, ageDiffSeconds

		if passedFilters == filtersCount {
			foundAccounts = append(foundAccounts, &CompatibilityResult{
				id:              account.ID,
				hasPremiumNow:   account.hasActivePremium(now),
				status:          account.Status,
				commonInterests: interestsIntersections,
				ageDiff:         int(math.Abs(float64(requestedAccount.Birth - account.Birth))),
				account:         account,
			})
		}
	}

	if len(foundAccounts) > 0 {
		sort.Sort(compatibilitySort(foundAccounts))

		var found []*AccountUpdated
		for _, v := range foundAccounts {
			if len(found) >= limit {
				break
			}
			found = append(found, v.account)
		}

		ctx.Success("application/json", prepareResponseBytesUpdated(found, []string{
			"id", "email", "status", "fname", "sname", "birth", "premium",
		}))
		return
	}

	emptyResponse(ctx)
	return
}

type CompatibilityResult struct {
	id              uint32
	hasPremiumNow   bool
	status          byte
	commonInterests int
	ageDiff         int
	account         *AccountUpdated
}

func (cr CompatibilityResult) getPremium() int {
	if cr.hasPremiumNow {
		return 0
	} else {
		return 1
	}
}

type compatibilitySort []*CompatibilityResult

func (s compatibilitySort) Len() int {
	return len(s)
}
func (s compatibilitySort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s compatibilitySort) Less(i, j int) bool {
	acc1 := s[i]
	acc2 := s[j]
	result := 0

	if acc1.getPremium() < acc2.getPremium() {
		return true
	} else if acc1.getPremium() > acc2.getPremium() {
		return false
	} else {
		result = 0
	}

	if result == 0 {
		if acc1.status < acc2.status {
			return true
		} else if acc1.status > acc2.status {
			return false
		} else {
			result = 0
		}
	}
	if result == 0 {
		if acc1.commonInterests > acc2.commonInterests {
			return true
		} else if acc1.commonInterests < acc2.commonInterests {
			return false
		} else {
			result = 0
		}
	}
	if result == 0 {
		if acc1.ageDiff < acc2.ageDiff {
			return true
		} else if acc1.ageDiff > acc2.ageDiff {
			return false
		} else {
			result = 0
		}
	}

	return acc1.id < acc2.id
}
