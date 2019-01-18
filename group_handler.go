package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/sets/treeset"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/valyala/fasthttp"
)

type Group struct {
	Name  string
	Count int

	subgroup1key   string
	subgroup1value string

	subgroup2key   string
	subgroup2value string

	subgroup3key   string
	subgroup3value string

	subgroup4key   string
	subgroup4value string

	subgroup5key   string
	subgroup5value string
}

func NewGroup(name string, count int) *Group {
	subgroups := strings.Split(name, "_")
	subGroupsLength := len(subgroups)
	var subgroup1key, subgroup2key, subgroup3key, subgroup4key, subgroup5key string
	var subgroup1value, subgroup2value, subgroup3value, subgroup4value, subgroup5value string
	if subGroupsLength > 0 {
		subgroupResults := strings.Split(subgroups[0], ":")
		if len(subgroupResults) > 1 && subgroupResults[1] != "" {
			subgroup1key = subgroupResults[0]
			subgroup1value = subgroupResults[1]
		}
	}
	if subGroupsLength > 1 {
		subgroupResults := strings.Split(subgroups[1], ":")
		if len(subgroupResults) > 1 && subgroupResults[1] != "" {
			subgroup2key = subgroupResults[0]
			subgroup2value = subgroupResults[1]
		}
	}
	if subGroupsLength > 2 {
		subgroupResults := strings.Split(subgroups[2], ":")
		if len(subgroupResults) > 1 && subgroupResults[1] != "" {
			subgroup3key = subgroupResults[0]
			subgroup3value = subgroupResults[1]
		}
	}
	if subGroupsLength > 3 {
		subgroupResults := strings.Split(subgroups[3], ":")
		if len(subgroupResults) > 1 && subgroupResults[1] != "" {
			subgroup4key = subgroupResults[0]
			subgroup4value = subgroupResults[1]
		}
	}
	if subGroupsLength > 4 {
		subgroupResults := strings.Split(subgroups[4], ":")
		if len(subgroupResults) > 1 && subgroupResults[1] != "" {
			subgroup5key = subgroupResults[0]
			subgroup5value = subgroupResults[1]
		}
	}

	return &Group{
		Name:           name,
		Count:          count,
		subgroup1key:   subgroup1key,
		subgroup1value: subgroup1value,
		subgroup2key:   subgroup2key,
		subgroup2value: subgroup2value,
		subgroup3key:   subgroup3key,
		subgroup3value: subgroup3value,
		subgroup4key:   subgroup4key,
		subgroup4value: subgroup4value,
		subgroup5key:   subgroup5key,
		subgroup5value: subgroup5value,
	}
}

type GroupList []*Group

func (p GroupList) Len() int      { return len(p) }
func (p GroupList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p GroupList) Less(i, j int) bool {
	result := 0

	if p[i].Count < p[j].Count {
		return true
	} else if p[i].Count > p[j].Count {
		return false
	} else {
		result = 0
	}

	if result == 0 {
		if p[i].subgroup1value < p[j].subgroup1value {
			return true
		} else if p[i].subgroup1value > p[j].subgroup1value {
			return false
		} else {
			result = 0
		}
	}

	if result == 0 {
		if p[i].subgroup2value < p[j].subgroup2value {
			return true
		} else if p[i].subgroup2value > p[j].subgroup2value {
			return false
		} else {
			result = 0
		}
	}

	if result == 0 {
		if p[i].subgroup3value < p[j].subgroup3value {
			return true
		} else if p[i].subgroup3value > p[j].subgroup3value {
			return false
		} else {
			result = 0
		}
	}

	if result == 0 {
		if p[i].subgroup4value < p[j].subgroup4value {
			return true
		} else if p[i].subgroup4value > p[j].subgroup4value {
			return false
		} else {
			result = 0
		}
	}

	return p[i].subgroup5value < p[j].subgroup5value
}

func groupHandler(ctx *fasthttp.RequestCtx) {
	allowedParams := map[string]int{
		"query_id": 1, "limit": 1,
		"order": 1, "keys": 1,
		"joined": 1,
	}
	_ = allowedParams

	allowedKeys := map[string]struct{}{
		"sex":       {},
		"status":    {},
		"interests": {},
		"country":   {},
		"city":      {},
	}

	var limit int
	var err error
	if limit, err = strconv.Atoi(string(ctx.QueryArgs().Peek("limit"))); err != nil || limit <= 0 {
		ctx.Error("{}", 400)
		return
	}

	var order int
	if order, err = strconv.Atoi(string(ctx.QueryArgs().Peek("order"))); err != nil || (order != -1 && order != 1) {
		ctx.Error("{}", 400)
		return
	}

	vnidxpool := namedIndexPool.Get()
	namedIndex := vnidxpool.(*NamedIndex)

	vmap := treemapPool.Get()
	suitableIndexes := vmap.(*treemap.Map)

	suitableIndexes.Put(accountIndex.Size(), namedIndex.Update([]byte("default"), accountIndex))

	groupKeys := treeset.NewWithStringComparator()
	keysF := ctx.QueryArgs().Peek("keys")
	hasInterestsKey := false
	if len(keysF) > 0 {
		valid := true
		for _, v := range strings.Split(string(keysF), ",") {
			if _, ok := allowedKeys[v]; ok {
				groupKeys.Add(v)
			} else {
				valid = false
				break
			}
		}
		if !valid {
			ctx.Error("{}", 400)
			return
		}
		hasInterestsKey = groupKeys.Contains("interests")
	}

	sexF := ctx.QueryArgs().Peek("sex")
	var sexFilter string
	if len(sexF) > 0 { //TODO: Add validation
		sexFilter = string(sexF)
	}

	emailF := ctx.QueryArgs().Peek("email")
	var emailFilter string
	if len(emailF) > 0 { //TODO: Add validation
		emailFilter = string(emailF)
	}

	statusF := ctx.QueryArgs().Peek("status")
	var statusFilter string
	if len(statusF) > 0 { //TODO: Add validation
		statusFilter = string(statusF)
	}

	fnameF := ctx.QueryArgs().Peek("fname")
	var fnameFilter string
	if len(fnameF) > 0 { //TODO: Add validation
		fnameFilter = string(fnameF)
	}

	snameF := ctx.QueryArgs().Peek("sname")
	var snameFilter string
	if len(snameF) > 0 { //TODO: Add validation
		snameFilter = string(snameF)
	}

	phoneF := ctx.QueryArgs().Peek("phone")
	var phoneFilter string
	if len(phoneF) > 0 { //TODO: Add validation
		phoneFilter = string(phoneF)
	}

	countryF := ctx.QueryArgs().Peek("country")
	var countryFilter string
	if len(countryF) > 0 { //TODO: Add validation
		countryFilter = string(countryF)
	}

	cityF := ctx.QueryArgs().Peek("city")
	var cityFilter string
	if len(cityF) > 0 { //TODO: Add validation
		cityFilter = string(cityF)
	}

	birthF := ctx.QueryArgs().Peek("birth")
	var birthFilter uint16
	if len(birthF) > 0 { //TODO: Add validation
		tmpUint, _ := strconv.ParseUint(string(birthF), 10, 16)
		birthFilter = uint16(tmpUint)
		if birthYearIndex.Exists(birthFilter) {
			currIndex := birthYearIndex.Get(birthFilter).(*treemap.Map)
			suitableIndexes.Put(
				currIndex.Size(),
				namedIndex.Update([]byte("birth_year"), currIndex),
			)
		} else {
			emptyGroupResponse(ctx)
			return
		}
	}

	joinedF := ctx.QueryArgs().Peek("joined")
	var joinedFilter uint16
	if len(joinedF) > 0 { //TODO: Add validation
		tempUint, _ := strconv.ParseUint(string(joinedF), 10, 16)
		joinedFilter = uint16(tempUint)
	}

	interestsF := ctx.QueryArgs().Peek("interests")
	var interestsFilter uint8
	if len(interestsF) > 0 { //TODO: Add validation
		interestsFilter = interestsDict.GetId(string(interestsF))
	}

	likesF := ctx.QueryArgs().Peek("likes")
	var likesFilter uint32
	if len(likesF) > 0 { //TODO: Add validation
		tempUint, _ := strconv.ParseUint(string(likesF), 10, 32)
		likesFilter = uint32(tempUint)
	}
	//todo: add premium filter support?

	var index *treemap.Map
	//TODO: Select index by filter

	var selectedIndexName []byte
	if suitableIndexes.Size() > 0 {
		if _, shortestIndex := suitableIndexes.Min(); &shortestIndex != nil {
			res := shortestIndex.(*NamedIndex)
			selectedIndexName = res.name
			index = res.index
		}
	}

	_ = selectedIndexName

	namedIndexPool.Put(vnidxpool)
	treemapPool.Put(vmap)

	var foundGroups = make(map[string]int)

	if index != nil {
		it := index.Iterator()
		for it.Next() {
			account := it.Value().(*AccountUpdated)

			// conditions
			if len(sexFilter) > 0 { //TODO: move len from loop
				if account.Sex != sexDict.GetId(sexFilter) {
					continue
				}
			}

			if len(emailFilter) > 0 {
				if account.Email != emailsDict.GetId(emailFilter) {
					continue
				}
			}

			if len(statusFilter) > 0 {
				if account.Status != statusDict.GetId(statusFilter) {
					continue
				}
			}

			if len(fnameFilter) > 0 {
				if account.Fname != fnamesDict.GetId(fnameFilter) {
					continue
				}
			}

			if len(snameFilter) > 0 {
				if account.Sname != snamesDict.GetId(snameFilter) {
					continue
				}
			}

			if len(phoneFilter) > 0 {
				if account.Phone != phonesDict.GetId(phoneFilter) {
					continue
				}
			}

			if len(countryFilter) > 0 {
				if account.Country != countriesDict.GetId(countryFilter) {
					continue
				}
			}

			if len(cityFilter) > 0 {
				if account.City != citiesDict.GetId(cityFilter) {
					continue
				}
			}

			if birthFilter != 0 {
				if account.BirthYear != birthFilter {
					continue
				}
			}

			if joinedFilter != 0 {
				if account.JoinedYear != joinedFilter {
					continue
				}
			}

			if interestsFilter != 0 {
				if _, ok := account.InterestsMap[interestsFilter]; !ok {
					continue
				}
			}

			if likesFilter != 0 {
				if _, ok := likesMap.getLikesFor(account.ID)[likesFilter]; !ok {
					continue
				}
			}

			//TODO: Premium?

			// key grouping
			var resultKey string
			groupKeys.Each(func(index int, value interface{}) {
				keyName := value.(string)
				switch keyName {
				case "sex":
					resultKey += "_" + keyName + ":" + string(account.Sex)
				case "status":
					resultKey += "_" + keyName + ":" + string(account.Status)
				case "country":
					resultKey += "_" + keyName + ":" + string(account.Country)
				case "city":
					resultKey += "_" + keyName + ":" + string(account.City)
				}
			})

			if hasInterestsKey {
				for interest := range account.InterestsMap {
					interestsKey := resultKey + "_interests:" + string(interest)
					foundGroups[interestsKey] += 1
				}
			} else if len(resultKey) > 0 {
				foundGroups[resultKey[1:]] += 1
			}
		}
	}

	if len(foundGroups) > 0 {
		var found = GroupList{}

		for k, v := range foundGroups {
			found = append(found, NewGroup(k, v))
		}

		// use reverse
		if order == -1 {
			sort.Sort(sort.Reverse(found))
		} else {
			sort.Sort(found)
		}

		if limit > 0 && len(found) > limit {
			found = found[:limit]
		}

		ctx.Success("application/json", prepareGroupResponseBytes(found))
		return
	}

	emptyGroupResponse(ctx)
	return
}

func emptyGroupResponse(ctx *fasthttp.RequestCtx) {
	ctx.Success("application/json", []byte(`{"groups":[]}`))
}

func prepareGroupResponseBytes(found []*Group) []byte {
	vbuf := bytesPool.Get()
	bytesBuffer := vbuf.([]byte)

	bytesBuffer = append(bytesBuffer, `{"groups":[`...)

	foundLen := len(found)

	for groupIdx, group := range found {
		lastAcc := groupIdx == foundLen-1
		_ = lastAcc

		bytesBuffer = append(bytesBuffer, `{`...)
		bytesBuffer = append(bytesBuffer, `"count":`...)
		bytesBuffer = fasthttp.AppendUint(bytesBuffer, group.Count)

		if group.subgroup1key != "" {
			bytesBuffer = append(bytesBuffer, `,"`+group.subgroup1key+`":"`+group.subgroup1value+`"`...)
		}
		if group.subgroup2key != "" {
			bytesBuffer = append(bytesBuffer, `,"`+group.subgroup2key+`":"`+group.subgroup2value+`"`...)
		}
		if group.subgroup3key != "" {
			bytesBuffer = append(bytesBuffer, `,"`+group.subgroup3key+`":"`+group.subgroup3value+`"`...)
		}
		if group.subgroup4key != "" {
			bytesBuffer = append(bytesBuffer, `,"`+group.subgroup4key+`":"`+group.subgroup4value+`"`...)
		}
		if group.subgroup5key != "" {
			bytesBuffer = append(bytesBuffer, `,"`+group.subgroup5key+`":"`+group.subgroup5value+`"`...)
		}

		bytesBuffer = append(bytesBuffer, `}`...)

		if !lastAcc {
			bytesBuffer = append(bytesBuffer, `,`...)
		}
	}

	bytesBuffer = append(bytesBuffer, `]}`...)

	bytesPool.Put(vbuf)

	return bytesBuffer
}
