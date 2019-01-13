package main

import (
	"testing"
)

func BenchmarkParseFile(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		parseFile("./data/accounts_1.json")
	}
}
