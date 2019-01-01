package main

import (
	"github.com/gin-gonic/gin/json"
	"github.com/tidwall/buntdb"
	"github.com/valyala/fasthttp"
)

func filterHandler(ctx *fasthttp.RequestCtx) {
	sex := ctx.QueryArgs().Peek("sex")

	var resultIds []uint
	var results []AccountAsMap

	db.View(func(tx *buntdb.Tx) error {
		tx.AscendEqual("sex", string(sex), func(key, val string) bool {
			resultIds = append(resultIds, GetIdFromKey(key))
			return true
		})
		return nil
	})

	for id := range resultIds {
		results = append(results, GetAccount(uint(id), []string{}))
	}

	jsonData, _ := json.Marshal(resultIds)

	ctx.Success("application/json", jsonData)
	return
}
