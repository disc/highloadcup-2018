package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/buntdb"
)

var columnList = []string{"id", "email", "fname", "sname", "phone", "sex", "birth", "country", "city", "joined", "status", "interests", "premium"}

type Timestamp int32

type Status uint8
type Sex rune

const (
	Free    Status = 0
	Busy    Status = 1
	Unknown Status = 2
	Male    Sex    = 'm'
	Female  Sex    = 'f'
)

type Interests []string

type Premium struct {
	Start  Timestamp `json:"start"`
	Finish Timestamp `json:"finish"`
}

type Like struct {
	ID uint      `json:"id"`
	Ts Timestamp `json:"ts"`
}

type Account struct {
	ID      uint      `json:"id"`
	Email   string    `json:"email"`
	Fname   string    `json:"fname"`
	Sname   string    `json:"sname"`
	Phone   string    `json:"phone"`
	Sex     Sex       `json:"sex"`
	Birth   Timestamp `json:"birth"`
	Country string    `json:"country"`
	City    string    `json:"city"`

	Joined    Timestamp `json:"joined"`
	Status    Status    `json:"status"`
	Interests Interests `json:"interests"`
	Premium   Premium   `json:"premium"`
}

type AccountAsMap map[string]string

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

func GetAccount(id int, columns []string) AccountAsMap {
	result := AccountAsMap{}

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

		for _, key := range columnList {
			if value, ok := data[key]; ok {
				val := fmt.Sprintf("%v", value)
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

type AccountMap struct {
	accounts map[uint]*Account
	sync.RWMutex
}

func (a *AccountMap) Get(id uint) *Account {
	a.RLock()
	defer a.RUnlock()

	return a.accounts[id]
}

func (a *AccountMap) Update(account Account) {
	a.Lock()
	a.accounts[account.ID] = &account
	a.Unlock()
}
