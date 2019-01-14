package main

import (
	"testing"
	"unsafe"

	"github.com/valyala/fasthttp"
)

var accounts = []*Account{
	{ID: 1, Email: "a1@b.com", Status: "f", Premium: map[string]int{"start": 1, "finish": 2}, Birth: 123},
	{ID: 2, Email: "a2@b.com", Status: "m", Premium: map[string]int{"start": 1, "finish": 2}, Birth: 456},
	{ID: 3, Email: "a3@b.com", Status: "f", Premium: map[string]int{"start": 1, "finish": 2}, Birth: 789},
	{ID: 4, Email: "a4@b.com", Status: "m", Premium: map[string]int{"start": 1, "finish": 2}, Birth: 246},
	{ID: 5, Email: "a5@b.com", Status: "f", Premium: map[string]int{"start": 1, "finish": 2}, Birth: 357},
}

var keys = []string{"id", "email", "status", "premium", "birth"}

func BenchmarkPrepareResponseBytes(b *testing.B) {
	var result []byte

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		result = prepareResponseBytes(accounts, keys)
	}
	_ = result
}

func BenchmarkFilterHandler(b *testing.B) {
	var ctx fasthttp.RequestCtx
	args := ctx.QueryArgs()

	args.Add("sex_eq", "f")
	args.Add("interests_contains", "YouTube,Бургеры")
	args.Add("limit", "50")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		filterHandler(&ctx)
	}
}

func BenchmarkContains(b *testing.B) {
	haystack := map[string]struct{}{
		"YouTube": {},
		"Пицца":   {},
		"Music":   {},
		"Sports":  {},
		"Пиво":    {},
		"Mac":     {},
		"Бургеры": {},
	}

	needle := map[string]struct{}{
		"YouTube": {},
		"Бургеры": {},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		filterContains(needle, haystack)
	}
}

func TestSizeofs(t *testing.T) {
	account := &AccountUpdated{}
	t.Log("*account", unsafe.Sizeof(account))
	t.Log("account.ID", unsafe.Sizeof(account.ID))
	t.Log("account.Email", unsafe.Sizeof(account.Email))
	t.Log("account.Fname", unsafe.Sizeof(account.Fname))
	t.Log("account.Sname", unsafe.Sizeof(account.Sname))
	t.Log("account.Phone", unsafe.Sizeof(account.Phone))
	t.Log("account.Sex", unsafe.Sizeof(account.Sex))
	t.Log("account.Birth", unsafe.Sizeof(account.Birth))
	t.Log("account.Country", unsafe.Sizeof(account.Country))
	t.Log("account.City", unsafe.Sizeof(account.City))
	t.Log("account.Joined", unsafe.Sizeof(account.Joined))
	t.Log("account.Status", unsafe.Sizeof(account.Status))
	t.Log("account.Premium", unsafe.Sizeof(account.Premium))
	t.Log("account.interestsMap", unsafe.Sizeof(account.interestsMap))
	t.Log("account.emailDomain", unsafe.Sizeof(account.emailDomain))
	t.Log("account.phoneCode", unsafe.Sizeof(account.phoneCode))
	t.Log("account.birthYear", unsafe.Sizeof(account.birthYear))
	t.Log("account.joinedYear", unsafe.Sizeof(account.joinedYear))
	t.Log("account.likes", unsafe.Sizeof(account.likes))
	t.Log("Account{}", unsafe.Sizeof(AccountUpdated{}))
}
