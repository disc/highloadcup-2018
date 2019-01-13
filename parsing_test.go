package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func BenchmarkParseFile(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	rawData, err := ioutil.ReadFile("./data/accounts_1.json")
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	for n := 0; n < b.N; n++ {
		parseAccountsMap(rawData)
	}
}
