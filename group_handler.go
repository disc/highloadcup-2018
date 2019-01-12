package main

import (
	"sort"
	"strconv"
	"strings"

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

	var groupKeys = make(map[string]struct{})
	keysF := ctx.QueryArgs().Peek("keys")
	if len(keysF) > 0 {
		// TODO: use bytes?
		for _, v := range strings.Split(string(keysF), ",") {
			groupKeys[v] = struct{}{}
		}
	}

	var index *treemap.Map

	vnidxpool := namedIndexPool.Get()
	namedIndex := vnidxpool.(*NamedIndex)

	vmap := treemapPool.Get()
	suitableIndexes := vmap.(*treemap.Map)
	suitableIndexes.Put(accountMap.Size(), namedIndex.Update([]byte("default"), accountMap))

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

	conditionsMap := map[string]interface{}{}

	var foundGroups = make(map[string]int)

	if index != nil {
		it := index.Iterator()
		for it.Next() {
			account := it.Value().(*Account)

			// conditions
			// TODO: get more conditions from filter handler
			if value, ok := conditionsMap["joined"]; ok {
				if account.joinedYear != value.(int) {
					continue
				}
			}
			if value, ok := conditionsMap["birth"]; ok {
				if account.birthYear != value.(int) {
					continue
				}
			}

			var resultKey string
			for keyName := range groupKeys {
				switch keyName {
				case "sex":
					resultKey += "_" + keyName + ":" + account.Sex
				case "status":
					resultKey += "_" + keyName + ":" + account.Status
				case "country":
					if account.Country != "" {
						resultKey += "_" + keyName + ":" + account.Country
					}
				case "city":
					if account.City != "" {
						resultKey += "_" + keyName + ":" + account.City
					}
				}
			}
			if _, ok := groupKeys["interests"]; ok {
				for interest := range account.interestsMap {
					interestsKey := resultKey + "_interests:" + interest
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

		ctx.Success("application/json", prepareGroupResponseBytes(found[:limit]))
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
