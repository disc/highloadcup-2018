package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type Account map[string]interface{}

func GetIdFromKey(key string) int64 {
	chunks := strings.SplitN(key, ":", 3)
	if len(chunks) > 1 {
		if id, err := strconv.Atoi(chunks[1]); err == nil {
			return int64(id)
		}
	}
	return 0
}

func BuildAccountKey(id int64) string {
	return fmt.Sprintf("acc:%d", id)
}

func GetAccount(id int) json.RawMessage {
	//var result string
	//
	//err := db.View(func(tx *buntdb.Tx) error {
	//	result, _ = tx.Get(BuildAccountKey(id))
	//
	//	return nil
	//})
	//
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//return json.RawMessage(result)

	return accountMap[id]
}

type SexMap map[string][]int

type AccountMap map[int]json.RawMessage

var sexMap = make(SexMap, 0)
var accountMap = make(AccountMap, 0)

func UpdateAccount(data *gjson.Result) {
	record := data.Map()
	recordId := int(record["id"].Int())
	sex := record["sex"].String()

	if sex != "" {
		sexMap[sex] = append(sexMap[sex], recordId)
	}

	accountMap[recordId] = json.RawMessage(data.Raw)
}
