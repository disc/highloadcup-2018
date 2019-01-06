package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"hash/crc32"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/handlers"
	_ "github.com/mkevac/debugcharts"

	"github.com/Sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	addr = ":80"

	now = time.Now().Unix()

	log = logrus.New()

	isDebugMode = os.Getenv("DEBUG")

	crc32q = crc32.MakeTable(0xD5828281)
)

func main() {
	if isDebugMode != "" {
		log.Println("Debug-mode enabled")
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()

		go func() {
			log.Fatal(http.ListenAndServe(":9090", handlers.CompressHandler(http.DefaultServeMux)))
		}()
	}

	log.Println("Started")

	parseDataDir("./data/")

	log.Println("Started calculateSimilarityIndex")
	//calculateSimilarityIndex()
	log.Println("Finished calculateSimilarityIndex, len is", len(similarityMap))

	log.Println("Data has been parsed completely")

	runtime.GC()
	log.Println("GC has been finished")

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
/accounts/uniqLikes/
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
		if pathLen >= 20 && pathLen <= 30 && path[pathLen-2] == 't' {
			// suggest
			suggestHandler(ctx, parseAccountId(path))
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

func parseAccountId(path []byte) int {
	from := bytes.IndexByte(path[1:], '/')
	to := bytes.IndexByte(path[from+2:], '/')

	if to == -1 {
		to = len(path)
	} else {
		to += from + 2
	}

	entityId, _ := strconv.Atoi(string(path[from+2 : to]))

	return entityId
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
	type jsonKey struct {
		Accounts []Account
	}

	var accounts jsonKey
	json.Unmarshal(fileBytes, &accounts)

	for _, account := range accounts.Accounts {
		createAccount(account)
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
