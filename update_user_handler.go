package main

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func updateUserHandler(ctx *fasthttp.RequestCtx, accountId int) {
	var account *Account
	if value, found := accountIndex.Get(accountId); !found {
		ctx.Error(`{"err":"user_not_found"}`, 404)
		return
	} else {
		account = value.(*Account)
	}

	p := pp.Get()

	jsonValue, err := p.ParseBytes(ctx.PostBody())

	if err != nil {
		ctx.Error("{}", 400)
		return
	}

	//TODO: Iterate by passed field

	// unique: Email
	if jsonValue.Exists("email") {
		email := string(jsonValue.Get("email").GetStringBytes())

		if email != "" {
			if len(email) > 100 {
				ctx.Error(`{"err":"email_too_long"}`, 400)
				return
			}
			if !strings.Contains(email, "@") {
				ctx.Error(`{"err":"incorrect_email_long"}`, 400)
				return
			}
			// unique email
			if emailsDict.GetId(email) > 0 {
				ctx.Error(`{"err":"email_already_exists"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_email_field"}`, 400)
			return
		}
	}

	// unique: Phone
	if jsonValue.Exists("phone") {
		phone := string(jsonValue.Get("phone").GetStringBytes())

		if phone != "" {
			if len(phone) > 16 {
				ctx.Error(`{"err":"phone_too_long"}`, 400)
				return
			}
			// unique phone
			if phonesDict.GetId(phone) > 0 {
				ctx.Error(`{"err":"phone_already_exists"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_phone_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("fname") {
		fname := string(jsonValue.Get("fname").GetStringBytes())

		if fname != "" {
			if len(fname) > 50 {
				ctx.Error(`{"err":"fname_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_fname_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("sname") {
		sname := string(jsonValue.Get("sname").GetStringBytes())

		if sname != "" {
			if len(sname) > 50 {
				ctx.Error(`{"err":"sname_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_sname_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("sex") {
		sex := string(jsonValue.Get("sex").GetStringBytes())

		if sex != "" {
			if len(sex) > 0 && sex != "m" && sex != "f" {
				ctx.Error(`{"err":"invalid_sex"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_sex_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("birth") {
		birth := jsonValue.GetInt("birth")

		if birth == 0 {
			ctx.Error(`{"err":"empty_birth_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("country") {
		country := string(jsonValue.Get("country").GetStringBytes())
		if country != "" {
			if len(country) > 50 {
				ctx.Error(`{"err":"country_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_country_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("city") {
		city := string(jsonValue.Get("city").GetStringBytes())

		if city != "" {
			if len(city) > 50 {
				ctx.Error(`{"err":"city_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_city_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("joined") {
		joined := jsonValue.GetInt("joined")

		if joined == 0 {
			ctx.Error(`{"err":"empty_joined_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("status") {
		status := string(jsonValue.Get("status").GetStringBytes())
		if status != "" {
			if len(status) > 0 && status != "свободны" && status != "заняты" && status != "всё сложно" {
				ctx.Error(`{"err":"invalid_status"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_status_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("interests") {
		interestsList := jsonValue.GetArray("interests")

		if interestsList != nil {
			for _, v := range interestsList {
				if len(v.GetStringBytes()) > 100 {
					ctx.Error(`{"err":"interest_too_long"}`, 400)
					return
				}
			}
		} else {
			ctx.Error(`{"err":"empty_interests_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("premium") {
		premiumObj := jsonValue.GetObject("premium")
		if premiumObj == nil {
			ctx.Error(`{"err":"empty_premium_field"}`, 400)
			return
		}
	}

	if jsonValue.Exists("likes") {
		likes := jsonValue.GetArray("likes")
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
	}

	pp.Put(p)

	// updating in goroutine
	go account.Update(ctx.PostBody())

	updatedSuccessResponse(ctx)
	return
}

func updatedSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(202)
}
