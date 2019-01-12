package main

import (
	"encoding/json"
	"strings"

	"github.com/valyala/fasthttp"
)

func createUserHandler(ctx *fasthttp.RequestCtx) {
	account := Account{}
	if err := json.Unmarshal(ctx.PostBody(), &account); err != nil {
		ctx.Error("{}", 400)
		return
	}

	//TODO: Iterate by passed field

	if account.ID == 0 || account.Email == "" {
		ctx.Error(`{"err":"empty_req_fields"}`, 400)
		return
	}

	if len(account.Email) > 100 {
		ctx.Error(`{"err":"email_too_long"}`, 400)
		return
	}

	if !strings.Contains(account.Email, "@") {
		ctx.Error(`{"err":"incorrect_email_long"}`, 400)
		return
	}

	if len(account.Fname) > 50 || len(account.Sname) > 50 {
		ctx.Error(`{"err":"fname_sname_too_long"}`, 400)
		return
	}

	if len(account.Phone) > 16 {
		ctx.Error(`{"err":"phone_too_long"}`, 400)
		return
	}

	if len(account.Sex) > 0 && account.Sex != "m" && account.Sex != "f" {
		ctx.Error(`{"err":"invalid_sex"}`, 400)
		return
	}

	if len(account.Country) > 50 {
		ctx.Error(`{"err":"country_too_long"}`, 400)
		return
	}

	if len(account.City) > 50 {
		ctx.Error(`{"err":"city_too_long"}`, 400)
		return
	}

	if len(account.Status) > 0 && account.Status != "свободны" && account.Status != "заняты" && account.Status != "всё сложно" {
		ctx.Error(`{"err":"invalid_status"}`, 400)
		return
	}

	if len(account.Interests) > 0 {
		for _, v := range account.Interests {
			if len(v) > 100 {
				ctx.Error(`{"err":"interest_too_long"}`, 400)
				return
			}
		}
	}

	// unique id
	if _, found := accountIndex.Get(account.ID); found {
		ctx.Error(`{"err":"id_already_exists"}`, 400)
		return
	}

	// unique email
	if emailIndex.Exists(account.Email) {
		ctx.Error(`{"err":"email_already_exists"}`, 400)
		return
	}

	// unique phone
	if phoneIndex.Exists(account.Phone) {
		ctx.Error(`{"err":"phone_already_exists"}`, 400)
		return
	}

	// creating in goroutine
	go NewAccount(account)

	// unique
	createdSuccessResponse(ctx)
	return
}

func createdSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(201)
}
