package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/tidwall/buntdb"
)

type Account map[string]interface{}

func GetIdFromKey(key string) int {
	chunks := strings.SplitN(key, ":", 3)
	if len(chunks) > 1 {
		if id, err := strconv.Atoi(chunks[1]); err == nil {
			return id
		}
	}
	return 0
}

func BuildKey(id interface{}, key string) string {
	return fmt.Sprintf("%s:%d", key, id)
}

func BuildAccountKey(id int64) string {
	return fmt.Sprintf("acc:%d", id)
}

func GetAccount(id int64) json.RawMessage {
	var result string

	err := db.View(func(tx *buntdb.Tx) error {
		result, _ = tx.Get(BuildAccountKey(id))

		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	return json.RawMessage(result)
}

func UpdateAccount(data *gjson.Result) {
	record := data.Map()
	recordId := record["id"].Int()
	err := db.Update(func(tx *buntdb.Tx) error {
		//// email-domain
		//if email, ok := data["email"]; ok {
		//	email := fmt.Sprintf("%v", email)
		//	components := strings.Split(email, "@")
		//	domain := components[1]
		//	_, _, err := tx.Set(BuildKey(data["id"], "email:domain"), domain, nil)
		//	if err != nil {
		//		log.Fatal("Email-domain setting error", err)
		//	}
		//
		//}
		// birth-year
		if record["birth"].Exists() {
			//TODO: Rewrite way of getting birth date (float -> string)
			tm := time.Unix(int64(record["birth"].Num), 0)

			_, _, err := tx.Set(BuildKey(recordId, "birth_year"), tm.Format("2006"), nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}
		// phone_code
		if record["phone"].Exists() {
			phoneCode := strings.SplitN(strings.SplitN(record["phone"].String(), "(", 2)[1], ")", 2)[0]
			_, _, err := tx.Set(BuildKey(recordId, "phone_code"), phoneCode, nil)
			if err != nil {
				log.Fatal("Phone-code setting error", err)
			}
		}
		// premium-to
		if record["premium"].Exists() {
			premiumMap := record["premium"].Map()

			_, _, err := tx.Set(BuildKey(recordId, "premium_to"), premiumMap["finish"].Str, nil)
			if err != nil {
				log.Fatal("Premium-to setting error", err)
			}
		}

		_, _, err := tx.Set(BuildAccountKey(recordId), data.Raw, nil)
		if err != nil {
			log.Fatal("Setting error", err)
		}

		return nil

	})

	if err != nil {
		log.Fatalln("Transaction error", err)
		return
	}
}
