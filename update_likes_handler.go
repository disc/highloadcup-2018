package main

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type Like struct {
	Likee int
	Liker int
	Ts    int
}
type LikesPayload struct {
	Likes []Like
}

func updateLikesHandler(ctx *fasthttp.RequestCtx) {
	var likes LikesPayload
	jsonData := ctx.PostBody()
	if err := json.Unmarshal(jsonData, &likes); err != nil {
		ctx.Error(`{"err":"invalid_payload"}`, 400)
		return
	}

	for _, v := range likes.Likes {
		if _, found := accountIndex.Get(v.Likee); !found {
			ctx.Error(`{"err":"likee_not_found"}`, 400)
			return
		}
		if _, found := accountIndex.Get(v.Liker); !found {
			ctx.Error(`{"err":"liker_not_found"}`, 400)
			return
		}
	}

	// updating in goroutine
	go updateLikes(jsonData)

	updatedSuccessResponse(ctx)
	return
}
