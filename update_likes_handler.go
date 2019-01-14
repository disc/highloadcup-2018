package main

import (
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
	p := pp.Get()

	jsonData, err := p.ParseBytes(ctx.PostBody())

	if err != nil {
		ctx.Error("{}", 400)
		return
	}

	likes := jsonData.GetArray("likes")

	for _, v := range likes {
		if _, found := accountIndex.Get(v.GetInt("likee")); !found {
			ctx.Error(`{"err":"likee_not_found"}`, 400)
			return
		}
		if _, found := accountIndex.Get(v.GetInt("liker")); !found {
			ctx.Error(`{"err":"liker_not_found"}`, 400)
			return
		}
		if v.GetInt("ts") == 0 {
			ctx.Error(`{"err":"invalid_ts"}`, 400)
			return
		}
	}

	pp.Put(p)

	// updating in goroutine
	go updateLikes(ctx.PostBody())

	updatedSuccessResponse(ctx)
	return
}
