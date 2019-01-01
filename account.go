package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/buntdb"
)

var columnList = []string{"id", "email", "fname", "sname", "phone", "sex", "birth", "country", "city", "joined", "status", "interests", "premium"}

type Account map[string]string

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
	return fmt.Sprintf("acc:%v:%s", id, key)
}

func GetAccount(id int, columns []string) Account {
	result := Account{}

	if len(columns) == 0 {
		columns = columnList
	}

	err := db.View(func(tx *buntdb.Tx) error {
		for _, key := range columns {
			val, _ := tx.Get(BuildKey(id, key))
			result[key] = val
		}

		return nil
	})

	result["id"] = fmt.Sprintf("%v", id)

	if err != nil {
		log.Fatalln(err)
	}

	return result
}

func UpdateAccount(data map[string]interface{}) {
	err := db.Update(func(tx *buntdb.Tx) error {
		// email-domain
		if email, ok := data["email"]; ok {
			email := fmt.Sprintf("%v", email)
			components := strings.Split(email, "@")
			domain := components[1]
			_, _, err := tx.Set(BuildKey(data["id"], "email:domain"), domain, nil)
			if err != nil {
				log.Fatal("Email-domain setting error", err)
			}

		}
		// birth-year
		if birth, ok := data["birth"]; ok {
			//TODO: Rewrite way of getting birth date (float -> string)
			birthF64 := birth.(float64)
			birth := strconv.FormatFloat(birthF64, 'f', 0, 64)
			tm := time.Unix(int64(birthF64), 0)
			data["birth"] = string(birth)

			_, _, err := tx.Set(BuildKey(data["id"], "birth:year"), tm.Format("2006"), nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}
		// phone_code
		if phone, ok := data["phone"]; ok {
			phoneCode := strings.SplitN(strings.SplitN(phone.(string), "(", 2)[1], ")", 2)[0]
			_, _, err := tx.Set(BuildKey(data["id"], "phone:code"), phoneCode, nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}
		// premium-to
		if premium, ok := data["premium"]; ok {
			premiumMap, _ := premium.(map[string]interface{})
			//TODO: Rewrite way of getting birth date (float -> string)
			premiumFinishF64 := premiumMap["finish"].(float64)
			premiumFinish := strconv.FormatFloat(premiumFinishF64, 'f', 0, 64)

			_, _, err := tx.Set(BuildKey(data["id"], "premium:to"), string(premiumFinish), nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}

		for _, key := range columnList {
			if value, ok := data[key]; ok {
				val := fmt.Sprintf("%v", value)
				if key == "interests" {
					//log.Printf("%#v", data[key])
				}
				_, _, err := tx.Set(BuildKey(data["id"], key), val, nil)
				if err != nil {
					log.Fatal("Setting error", err)
				}
			}
		}

		return nil

	})

	if err != nil {
		log.Fatalln("Transaction error", err)
		return
	}
}
