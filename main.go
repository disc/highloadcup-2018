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

	"github.com/tidwall/buntdb"
	"github.com/valyala/fasthttp"
)

var (
	addr = ":80"

	db *buntdb.DB

	now = int(time.Now().Unix())

	log = logrus.New()
)

func initDB() {
	err := db.Update(func(tx *buntdb.Tx) error {
		//err := tx.CreateIndex("id", "acc:*:id", buntdb.IndexInt)
		err := tx.CreateIndex("sex", "acc:*", buntdb.IndexJSON("sex"))
		err = tx.CreateIndex("email", "acc:*", buntdb.IndexJSON("email"))
		err = tx.CreateIndex("status", "acc:*", buntdb.IndexJSON("status"))
		err = tx.CreateIndex("fname", "acc:*", buntdb.IndexJSON("fname"))
		err = tx.CreateIndex("sname", "acc:*", buntdb.IndexJSON("sname"))
		err = tx.CreateIndex("country", "acc:*", buntdb.IndexJSON("country"))
		err = tx.CreateIndex("city", "acc:*", buntdb.IndexJSON("city"))
		err = tx.CreateIndex("birth", "acc:*", buntdb.IndexJSON("birth"))
		err = tx.CreateIndex("phone", "acc:*", buntdb.IndexJSON("phone"))
		err = tx.CreateIndex("phone_code", "phone_code:*", buntdb.IndexInt)
		err = tx.CreateIndex("birth_year", "birth_year:*", buntdb.IndexInt)
		err = tx.CreateIndex("premium_to", "premium_to:*", buntdb.IndexInt)

		//err = tx.CreateIndex("email", "acc:*:email", buntdb.IndexString)
		//err = tx.CreateIndex("email_domain", "acc:*:email:domain", buntdb.IndexString)
		//err = tx.CreateIndex("status", "acc:*:status", buntdb.IndexString)
		//err = tx.CreateIndex("fname", "acc:*:fname", buntdb.IndexString)
		//err = tx.CreateIndex("sname", "acc:*:sname", buntdb.IndexString)
		//err = tx.CreateIndex("phone_code", "acc:*:phone:code", buntdb.IndexInt)
		//err = tx.CreateIndex("country", "acc:*:country", buntdb.IndexString)
		//err = tx.CreateIndex("city", "acc:*:city", buntdb.IndexString)
		//err = tx.CreateIndex("birth", "acc:*:birth", buntdb.IndexInt)
		//err = tx.CreateIndex("birth_year", "acc:*:birth:year", buntdb.IndexInt)
		//err = tx.CreateIndex("premium_to", "acc:*:premium:to", buntdb.IndexInt)
		//err = tx.CreateIndex("interests", "acc:*:interests", buntdb.IndexString)
		// todo: interests
		// todo: likes

		return err
	})

	if err != nil {
		log.Fatal(err)
	}
}

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
		if pathLen > 15 && path[15] == 'r' {
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
		UpdateAccount(&account)
	}

	//for _, account := range data.Accounts {
	//	UpdateAccount(account)
	//}
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
