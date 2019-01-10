package main

import (
	"math"
	"testing"

	"github.com/emirpasic/gods/lists/arraylist"
)

func Test2Comparator(t *testing.T) {
	list := arraylist.New()
	list.Add(&CompatibilityResult{
		id:              1,
		hasPremiumNow:   false,
		status:          "свободны",
		commonInterests: 1,
	})
	list.Add(&CompatibilityResult{
		id:              2,
		hasPremiumNow:   true,
		status:          "все сложно",
		commonInterests: 1,
	})
	list.Add(&CompatibilityResult{
		id:              3,
		hasPremiumNow:   false,
		status:          "свободны",
		commonInterests: 1,
	})
	list.Add(&CompatibilityResult{
		id:              4,
		hasPremiumNow:   true,
		status:          "свободны",
		commonInterests: 1,
		ageDiff:         0,
	})
	list.Add(&CompatibilityResult{
		id:              5,
		hasPremiumNow:   true,
		commonInterests: 2,
		status:          "свободны",
		ageDiff:         3,
	})
	list.Add(&CompatibilityResult{
		id:              6,
		hasPremiumNow:   true,
		commonInterests: 2,
		status:          "свободны",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              7,
		hasPremiumNow:   true,
		commonInterests: 3,
		status:          "свободны",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              8,
		hasPremiumNow:   false,
		commonInterests: 2,
		status:          "все сложно",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              9,
		hasPremiumNow:   false,
		commonInterests: 3,
		status:          "все сложно",
		ageDiff:         2,
	})

	list.Sort(compatibilityComparator)

	for _, v := range list.Values() {
		log.Println(v.(*CompatibilityResult).id)
	}
}

func TestComparator(t *testing.T) {
	expectedData := []int{17254, 16886, 11154, 9546, 16188, 26286}

	parseDataDir("./data/")

	list := arraylist.New()
	requestedAccount, _ := accountMap.Get(11335)

	for _, acc := range accountMap.Values() {
		account := acc.(*Account)
		intersectionsCount := intersectionsCount(requestedAccount.(*Account).interestsMap, account.interestsMap)
		if intersectionsCount == 0 {
			continue
		}

		list.Add(&CompatibilityResult{
			id:              account.ID,
			hasPremiumNow:   account.hasActivePremium(now),
			status:          account.Status,
			commonInterests: intersectionsCount,
			ageDiff:         math.Abs(float64(requestedAccount.(*Account).Birth - account.Birth)),
			account:         account,
		})
	}

	list.Sort(compatibilityComparator)

	results := list.Values()

	for idx, value := range expectedData {
		foundAccId := results[idx].(*CompatibilityResult).id
		if foundAccId != expectedData[idx] {
			t.Error("Incorrect position for #", idx, "Expected", value, "got", foundAccId)
		}
	}
}
