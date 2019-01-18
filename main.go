package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fastjson"

	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/handlers"
	_ "github.com/mkevac/debugcharts"

	"github.com/Sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	addr = ":80"

	now = uint32(time.Now().Unix())

	log = logrus.New()

	isDebugMode = os.Getenv("DEBUG")

	//dataDir = "/Users/disc/Downloads/elim_accounts_261218/data/data/"
	dataDir = "./data/"

	pp fastjson.ParserPool
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

	parseDataDir(dataDir)

	log.Println("Data has been parsed completely")

	runtime.GC()
	log.Println("GC has been finished")
	log.Println("Stats:")
	log.Println("Accounts count:", accountIndex.Size())
	log.Println("Countries index length:", len(countryIndex.v))
	log.Println("Cities index length:", len(cityIndex.v))
	log.Println("Interests index length:", len(interestsIndex.v))
	log.Println("BirthYear index index length:", len(birthYearIndex.v))
	log.Println("Fname  index length:", len(fnameIndex.v))
	log.Println("Sname index index length:", len(snameIndex.v))

	if err := fasthttp.ListenAndServe(addr, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()

	isGetRequest := ctx.IsGet()
	isPostRequest := ctx.IsPost()

	pathLen := len(path)

	if isGetRequest {
		// /accounts/group/
		if pathLen == 16 && path[14] == 'p' {
			groupHandler(ctx)
			return
		}
		// /accounts/filter/
		if pathLen == 17 && path[15] == 'r' {
			filterHandler(ctx)
			return
		}
		// /accounts/<id>/suggest/
		if pathLen >= 20 && pathLen <= 30 && path[pathLen-2] == 't' {
			suggestHandler(ctx, parseAccountId(path))
			return
		}
		// /accounts/<id>/recommend/
		if pathLen >= 21 && pathLen <= 31 && path[pathLen-2] == 'd' {
			recommendHandler(ctx, parseAccountId(path))
			return
		}

		// 404
		ctx.Error("{}", 404)
		return
	}

	if isPostRequest {
		// /accounts/new/
		if pathLen == 14 && path[pathLen-2] == 'w' {
			createUserHandler(ctx)
			return
		}
		// /accounts/likes/
		if pathLen == 16 && path[pathLen-2] == 's' {
			updateLikesHandler(ctx)
			return
		}
		// /accounts/<id>/
		if pathLen >= 12 && pathLen <= 21 && path[8] == 's' {
			updateUserHandler(ctx, parseAccountId(path))
			return
		}
		// 404
		ctx.Error("{}", 404)
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
	if strings.LastIndex(filename, "accounts_") != -1 {
		rawData, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}

		p := pp.Get()

		v, _ := p.ParseBytes(rawData)
		//if err != nil {
		//	log.Fatalf("unexpected error: %s", err)
		//}

		for _, jsonData := range v.GetArray("accounts") {
			NewAccountFromJson(jsonData)
		}

		pp.Put(p)

		if time.Now().Second()%3 == 0 {
			runtime.GC()
		}

		//parseAccountsMap(rawData)
	} else if strings.LastIndex(filename, "options.txt") != -1 {
		parseOptions(filename)
	}
}

func parseDataDir(dirPath string) {
	files, _ := ioutil.ReadDir(dirPath)
	for _, f := range files {
		log.Println("Started parsing file", f.Name())
		parseFile(dirPath + f.Name())
	}
}

func parseAccountsMap(fileBytes []byte) {
	//type jsonKey struct {
	//	Accounts []Account
	//}
	//
	//var accounts jsonKey
	//json.Unmarshal(fileBytes, &accounts)

	// 20 seconds: 2.12Gb peak, 1.31Gb result
	//sc.InitBytes(fileBytes)
	//
	//for sc.Next() {
	//	//v := sc.Value()
	//	//_ = v
	//}

	//for _, account := range accounts.Accounts {
	//	NewAccount(account)
	//}
}

func parseOptions(filename string) {
	if file, err := os.OpenFile(filename, os.O_RDONLY, 0644); err == nil {
		reader := bufio.NewReader(file)
		if line, _, err := reader.ReadLine(); err == nil {
			tmpUint, _ := strconv.ParseUint(string(line), 10, 32)
			now = uint32(tmpUint)
			log.Println("`Now` was updated from options.txt", now)
		}
	}
}
