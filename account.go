package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/buntdb"
)

var columnList = []string{"id", "email", "fname", "sname", "phone", "sex", "birth", "country", "city", "joined", "status", "interests", "premium"}

type Account map[string]string

type AccountRequest struct {
	ID      uint   `json:"id"`
	Email   string `json:"email,omitempty"`
	Fname   string `json:"fname,omitempty"`
	Sname   string `json:"sname,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Sex     string `json:"sex,omitempty"`
	Birth   int64  `json:"birth,omitempty"`
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`

	Joined    int32  `json:"joined,omitempty"`
	Status    string `json:"status,omitempty"`
	Interests string `json:"interests,string,omitempty"`
	Premium   string `json:"premium,string,omitempty"`
}

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

func UpdateAccount(data AccountRequest) {
	err := db.Update(func(tx *buntdb.Tx) error {
		// email-domain
		if data.Email != "" {
			components := strings.Split(data.Email, "@")
			domain := components[1]
			_, _, err := tx.Set(BuildKey(data.ID, "email:domain"), domain, nil)
			if err != nil {
				log.Fatal("Email-domain setting error", err)
			}

		}
		// birth-year
		if &data.Birth != nil {
			//TODO: Rewrite way of getting birth date (float -> string)
			tm := time.Unix(data.Birth, 0)

			_, _, err := tx.Set(BuildKey(data.ID, "birth:year"), tm.Format("2006"), nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}
		// phone_code
		if &data.Phone != nil {
			log.Println(data.Phone, &data.Phone)
			phoneCode := strings.SplitN(strings.SplitN(data.Phone, "(", 2)[1], ")", 2)[0]
			_, _, err := tx.Set(BuildKey(data.ID, "phone:code"), phoneCode, nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}
		// premium-to
		if &data.Premium != nil {
			var objmap map[string]*json.RawMessage
			json.Unmarshal([]byte(data.Premium), &objmap)

			log.Println(objmap["finish"])

			//TODO: Rewrite way of getting birth date (float -> string)
			//premiumFinishF64 := objmap["finish"].(float64)
			//premiumFinish := strconv.FormatFloat(premiumFinishF64, 'f', 0, 64)

			_, _, err := tx.Set(BuildKey(data.ID, "premium:to"), string("1"), nil)
			if err != nil {
				log.Fatal("Birth-year setting error", err)
			}
		}

		//for _, key := range columnList {
		//	if value, ok := data[key]; ok {
		//		val := fmt.Sprintf("%v", value)
		//		if key == "interests" {
		//			//log.Printf("%#v", data[key])
		//		}
		//		_, _, err := tx.Set(BuildKey(data["id"], key), val, nil)
		//		if err != nil {
		//			log.Fatal("Setting error", err)
		//		}
		//	}
		//}

		return nil

	})

	if err != nil {
		log.Fatalln("Transaction error", err)
		return
	}
}
