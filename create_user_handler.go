package main

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func createUserHandler(ctx *fasthttp.RequestCtx) {
	p := pp.Get()

	jsonData, err := p.ParseBytes(ctx.PostBody())

	if err != nil {
		ctx.Error("{}", 400)
		return
	}

	//TODO: Iterate by passed field

	email := string(jsonData.Get("email").GetStringBytes())
	if jsonData.Get("id").GetInt() == 0 || email == "" {
		ctx.Error(`{"err":"empty_req_fields"}`, 400)
		return
	}

	if len(email) > 100 {
		ctx.Error(`{"err":"email_too_long"}`, 400)
		return
	}

	if !strings.Contains(email, "@") {
		ctx.Error(`{"err":"incorrect_email_long"}`, 400)
		return
	}

	if len(jsonData.Get("fname").GetStringBytes()) > 50 || len(jsonData.Get("sname").GetStringBytes()) > 50 {
		ctx.Error(`{"err":"fname_sname_too_long"}`, 400)
		return
	}

	phone := string(jsonData.Get("phone").GetStringBytes())
	if len(phone) > 16 {
		ctx.Error(`{"err":"phone_too_long"}`, 400)
		return
	}

	sex := string(jsonData.Get("sex").GetStringBytes())
	if len(sex) > 0 && sex != "m" && sex != "f" {
		ctx.Error(`{"err":"invalid_sex"}`, 400)
		return
	}

	if len(jsonData.Get("country").GetStringBytes()) > 50 {
		ctx.Error(`{"err":"country_too_long"}`, 400)
		return
	}

	if len(jsonData.Get("city").GetStringBytes()) > 50 {
		ctx.Error(`{"err":"city_too_long"}`, 400)
		return
	}

	status := string(jsonData.Get("status").GetStringBytes())
	if len(status) > 0 && status != "свободны" && status != "заняты" && status != "всё сложно" {
		ctx.Error(`{"err":"invalid_status"}`, 400)
		return
	}

	for _, v := range jsonData.GetArray("interests") {
		if len(v.GetStringBytes()) > 100 {
			ctx.Error(`{"err":"interest_too_long"}`, 400)
			return
		}
	}

	//todo: premium?

	// unique id
	if _, found := accountIndex.Get(jsonData.GetInt("id")); found {
		ctx.Error(`{"err":"id_already_exists"}`, 400)
		return
	}

	// unique email
	if emailIndex.Exists(email) {
		ctx.Error(`{"err":"email_already_exists"}`, 400)
		return
	}

	// unique phone
	if phoneIndex.Exists(phone) {
		ctx.Error(`{"err":"phone_already_exists"}`, 400)
		return
	}

	pp.Put(p)

	// creating in goroutine
	go NewAccountFromByte(ctx.PostBody())

	createdSuccessResponse(ctx)
	return
}

func createdSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(201)
}
