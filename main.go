package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/Sirupsen/logrus"

	"github.com/valyala/fasthttp"
)

var (
	addr = ":80"

	now = time.Now().Unix()

	log = logrus.New()
)

func main() {
	log.Println("Started")

	parseDataDir("./data/")
	log.Println("Data has been parsed completely")

	if err := fasthttp.ListenAndServe(addr, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

/*
GET:
/accounts/filter/
/accounts/group/
/accounts/<id>/recommend/
/accounts/<id>/suggest/

POST:
/accounts/new/
/accounts/<id>/
/accounts/likes/
*/

func requestHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()

	isGetRequest := ctx.IsGet()

	pathLen := len(path)

	if isGetRequest {
		if pathLen > 14 && path[14] == 'p' {
			// group
			//FIXME
			ctx.Success("application/json", []byte("{\"groups\":[]}"))
			return
		}
		if pathLen > 15 && pathLen <= 17 && path[15] == 'r' {
			// filter
			filterHandler(ctx)
			return
		}
		if pathLen > 21 && path[21] == 't' {
			// suggest
			//FIXME
			ctx.Success("application/json", []byte("{\"accounts\":[]}"))
			return
		}
		if pathLen > 23 && path[23] == 'd' {
			// recommend
			//FIXME
			ctx.Success("application/json", []byte("{\"accounts\":[]}"))
			return
		}
		// 404
		ctx.Error("{}", 404)
		return
	}

}

func parseFile(filename string) {
	rawData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	if strings.LastIndex(filename, "accounts_") != -1 {
		parseAccountsMap(rawData)
	} else if strings.LastIndex(filename, "options.txt") != -1 {
		parseOptions(filename)
	}
}

func parseDataDir(dirPath string) {
	files, _ := ioutil.ReadDir(dirPath)
	for _, f := range files {
		parseFile(dirPath + f.Name())
	}
}

func parseAccountsMap(fileBytes []byte) {
	result := gjson.GetBytes(fileBytes, "accounts")
	for _, account := range result.Array() {
		UpdateAccount(account)
	}
}

func parseOptions(filename string) {
	if file, err := os.OpenFile(filename, os.O_RDONLY, 0644); err == nil {
		reader := bufio.NewReader(file)
		if line, _, err := reader.ReadLine(); err == nil {
			now, _ = strconv.ParseInt(string(line), 10, 32)
			log.Println("`Now` was updated from options.txt", now)
		}
	}
}
