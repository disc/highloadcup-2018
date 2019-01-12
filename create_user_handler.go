package main

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func createUserHandler(ctx *fasthttp.RequestCtx) {
	acc := Account{}
	if err := json.Unmarshal(ctx.PostBody(), &acc); err != nil {
		ctx.Error("{}", 400)
		return
	}

	if acc.ID == 0 || acc.Email == "" {
		ctx.Error(`{"err":"empty_req_fields"}`, 400)
		return
	}

	if len(acc.Email) > 100 {
		ctx.Error(`{"err":"email_too_long"}`, 400)
		return
	}

	if len(acc.Fname) > 50 || len(acc.Sname) > 50 {
		ctx.Error(`{"err":"fname_sname_too_long"}`, 400)
		return
	}

	if len(acc.Phone) > 16 {
		ctx.Error(`{"err":"phone_too_long"}`, 400)
		return
	}

	if len(acc.Sex) > 0 && acc.Sex != "m" && acc.Sex != "f" {
		ctx.Error(`{"err":"invalid_sex"}`, 400)
		return
	}

	// todo birth validate

	if len(acc.Country) > 50 {
		ctx.Error(`{"err":"country_too_long"}`, 400)
		return
	}

	if len(acc.City) > 50 {
		ctx.Error(`{"err":"city_too_long"}`, 400)
		return
	}

	//todo joined validate

	if len(acc.Status) > 0 && acc.Status != "свободны" && acc.Status != "заняты" && acc.Status != "всё сложно" {
		ctx.Error(`{"err":"invalid_status"}`, 400)
		return
	}

	if len(acc.Interests) > 0 {
		for _, v := range acc.Interests {
			if len(v) > 100 {
				ctx.Error(`{"err":"interest_too_long"}`, 400)
				return
			}
		}
	}

	//todo premium validate

	// todo likes validate

	// unique id
	if _, found := accountMap.Get(acc.ID); found {
		ctx.Error(`{"err":"id_already_exists"}`, 400)
		return
	}

	// unique email
	if _, found := emailIndex[acc.Email]; found {
		ctx.Error(`{"err":"email_already_exists"}`, 400)
		return
	}

	// unique phone
	if _, found := phoneIndex[acc.Phone]; found {
		ctx.Error(`{"err":"phone_already_exists"}`, 400)
		return
	}

	createAccount(acc)

	// unique
	createdSuccessResponse(ctx)
	return
}

func createdSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(201)
}
