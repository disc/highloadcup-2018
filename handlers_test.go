package main

import "testing"

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
