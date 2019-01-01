package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tidwall/buntdb"
	"github.com/valyala/fasthttp"
)

var (
	addr = ":3000"

	db *buntdb.DB

	accountsMap = AccountMap{accounts: make(map[uint]*Account)}

	now = int(time.Now().Unix())

	log = logrus.New()
)

func initDB() {
	err := db.Update(func(tx *buntdb.Tx) error {
		err := tx.CreateIndex("sex", "acc:*:sex", buntdb.IndexString)
		err = tx.CreateIndex("email", "acc:*:email", buntdb.IndexString)
		err = tx.CreateIndex("email_domain", "acc:*:email:domain", buntdb.IndexString)
		// todo: email domain
		err = tx.CreateIndex("status", "acc:*:status", buntdb.IndexString)
		err = tx.CreateIndex("fname", "acc:*:fname", buntdb.IndexString)
		err = tx.CreateIndex("sname", "acc:*:sname", buntdb.IndexString)
		// todo: phone code
		err = tx.CreateIndex("country", "acc:*:country", buntdb.IndexString)
		err = tx.CreateIndex("city", "acc:*:city", buntdb.IndexString)
		err = tx.CreateIndex("birth", "acc:*:birth", buntdb.IndexInt)
		err = tx.CreateIndex("birth_year", "acc:*:birth:year", buntdb.IndexInt)
		// todo: birth year
		// todo: interests
		// todo: likes
		// todo: premium

		return err
	})

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("Started")

	db, _ = buntdb.Open(":memory:")
	defer db.Close()

	initDB()

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

	latestChar := path[len(path)-1]

	if isGetRequest {
		if path[15] == 'r' {
			// filter
			filterHandler(ctx)
			return
		}
		if latestChar == 'p' {
			// filter
			return
		}
		if latestChar == 'd' {
			// recommend
			return
		}
		if latestChar == 't' {
			// recommend
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
	type jsonKey struct {
		Accounts []map[string]interface{}
	}

	var data jsonKey
	_ = json.Unmarshal(fileBytes, &data)

	for _, account := range data.Accounts {
		UpdateAccount(account)
	}
}

func parseOptions(filename string) {
	if file, err := os.OpenFile(filename, os.O_RDONLY, 0644); err == nil {
		reader := bufio.NewReader(file)
		if line, _, err := reader.ReadLine(); err == nil {
			now, _ = strconv.Atoi(string(line))
			log.Println("`Now` was updated from options.txt", now)
		}
	}
}
