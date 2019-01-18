package main

import (
	"math"
	"sort"
	"testing"

	"github.com/emirpasic/gods/lists/arraylist"
)

func init() {
	parseDataDir("./data/")
}

func _Test2Comparator(t *testing.T) {
	list := arraylist.New()
	list.Add(&CompatibilityResult{
		id:              1,
		hasPremiumNow:   false,
		status:          "свободны",
		commonInterests: 1,
		ageDiff:         0,
	})
	list.Add(&CompatibilityResult{
		id:              2,
		hasPremiumNow:   true,
		status:          "всё сложно",
		commonInterests: 1,
		ageDiff:         0,
	})
	list.Add(&CompatibilityResult{
		id:              3,
		hasPremiumNow:   false,
		status:          "свободны",
		commonInterests: 1,
		ageDiff:         0,
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
		hasPremiumNow:   true,
		commonInterests: 3,
		status:          "всё сложно",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              9,
		hasPremiumNow:   true,
		commonInterests: 3,
		status:          "заняты",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              10,
		hasPremiumNow:   false,
		commonInterests: 2,
		status:          "всё сложно",
		ageDiff:         2,
	})
	list.Add(&CompatibilityResult{
		id:              11,
		hasPremiumNow:   false,
		commonInterests: 3,
		status:          "всё сложно",
		ageDiff:         2,
	})
	////
	//list.Add(&CompatibilityResult{
	//	id:            1,
	//	hasPremiumNow: true,
	//	status:        "всё сложно",
	//})
	//list.Add(&CompatibilityResult{
	//	id:            2,
	//	hasPremiumNow: true,
	//	status:        "свободны",
	//})
	//list.Add(&CompatibilityResult{
	//	id:            3,
	//	hasPremiumNow: true,
	//	status:        "заняты",
	//})
	//list.Add(&CompatibilityResult{
	//	id:            4,
	//	hasPremiumNow: false,
	//	status:        "всё сложно",
	//})
	//list.Add(&CompatibilityResult{
	//	id:            5,
	//	hasPremiumNow: false,
	//	status:        "свободны",
	//})
	//list.Add(&CompatibilityResult{
	//	id:            6,
	//	hasPremiumNow: false,
	//	status:        "заняты",
	//})
	//list.Sort(compatibilityComparator)

	var tempSlice []*CompatibilityResult
	for _, v := range list.Values() {
		tempSlice = append(tempSlice, v.(*CompatibilityResult))
	}

	sort.Sort(compatibilitySort(tempSlice))

	log.Println("ID, prem, status, interests, ageDiff")
	for _, res := range tempSlice {
		log.Printf("|%3v|%6v|%12v (%v)|%3v|%3v", res.id, res.hasPremiumNow, res.status, res.status, res.commonInterests, res.ageDiff)
	}
}

type compTest struct {
	id          int
	expectedIDs []int
}

var compTestData = []compTest{
	{11335, []int{17254, 16886, 11154, 9546, 16188, 26286}},
	{4421, []int{27078, 27136}},
}

func TestComparator(t *testing.T) {
	for _, compTest := range compTestData {
		expectedData := compTest.expectedIDs

		var list []*CompatibilityResult
		requestedAccount, _ := accountIndex.Get(compTest.id)

		for _, acc := range accountIndex.Values() {
			account := acc.(*Account)

			if requestedAccount.(*Account).Sex == account.Sex {
				continue
			}

			intersectionsCount := intersectionsCount(requestedAccount.(*Account).interestsMap, account.interestsMap)
			if intersectionsCount == 0 {
				continue
			}

			list = append(list, &CompatibilityResult{
				id:              account.ID,
				hasPremiumNow:   account.hasActivePremium(now),
				status:          account.Status,
				commonInterests: intersectionsCount,
				ageDiff:         int(math.Abs(float64(requestedAccount.(*Account).Birth - account.Birth))),
				account:         account,
			})
		}

		var expectedList []*CompatibilityResult
		for _, accID := range expectedData {
			account, _ := accountIndex.Get(accID)

			expectedList = append(expectedList, &CompatibilityResult{
				id:              account.(*Account).ID,
				hasPremiumNow:   account.(*Account).hasActivePremium(now),
				status:          account.(*Account).Status,
				commonInterests: intersectionsCount(requestedAccount.(*Account).interestsMap, account.(*Account).interestsMap),
				ageDiff:         requestedAccount.(*Account).Birth - account.(*Account).Birth,
				account:         account.(*Account),
			})
		}

		sort.Sort(compatibilitySort(list))

		isError := false
		for idx, value := range expectedList {
			foundAccId := list[idx].id

			if foundAccId != expectedData[idx] {
				isError = true

				res := value
				log.Println("Expected:")
				log.Printf("|%6v(%v)|%6v|%12v|%3v|%15v", res.id, res.account.Sex, res.hasPremiumNow, res.status, res.commonInterests, res.ageDiff)
				t.Error("Incorrect position for #", idx, "Expected", value.id, "got", foundAccId)
			}
		}
		if isError {
			log.Println("Actual:")

			for _, res := range list[:10] {
				log.Printf("|%6v(%v)|%6v|%12v|%3v|%15v", res.id, res.account.Sex, res.hasPremiumNow, res.status, res.commonInterests, res.ageDiff)
			}
		}
	}
}
