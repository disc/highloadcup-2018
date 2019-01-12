package main

import (
	"encoding/json"
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

	var data map[string]interface{}
	if err := json.Unmarshal(ctx.PostBody(), &data); err != nil {
		ctx.Error(`{"err":"empty_req_fields"}`, 400)
	}

	//TODO: Iterate by passed field

	//if _, found := data["id"]; found {
	//	ctx.Error(`{"err":"id_passed"}`, 400)
	//}

	// unique: Email
	if value, ok := data["email"]; ok {
		if value != nil {
			email := value.(string)
			if len(email) > 100 {
				ctx.Error(`{"err":"email_too_long"}`, 400)
				return
			}
			if !strings.Contains(email, "@") {
				ctx.Error(`{"err":"incorrect_email_long"}`, 400)
				return
			}
			// unique email
			if emailIndex.Exists(email) {
				ctx.Error(`{"err":"email_already_exists"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_email_field"}`, 400)
			return
		}
	}

	// unique: Phone
	if value, ok := data["phone"]; ok {
		if value != nil {
			phone := value.(string)
			if len(phone) > 16 {
				ctx.Error(`{"err":"phone_too_long"}`, 400)
				return
			}
			// unique phone
			if phoneIndex.Exists(phone) {
				ctx.Error(`{"err":"phone_already_exists"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_phone_field"}`, 400)
			return
		}
	}

	if value, ok := data["fname"]; ok {
		if value != nil {
			fname := value.(string)
			if len(fname) > 50 {
				ctx.Error(`{"err":"fname_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_fname_field"}`, 400)
			return
		}
	}

	if value, ok := data["sname"]; ok {
		if value != nil {
			sname := value.(string)
			if len(sname) > 50 {
				ctx.Error(`{"err":"sname_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_sname_field"}`, 400)
			return
		}
	}

	if value, ok := data["sex"]; ok {
		if value != nil {
			sex := value.(string)
			if len(sex) > 0 && sex != "m" && sex != "f" {
				ctx.Error(`{"err":"invalid_sex"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_sex_field"}`, 400)
			return
		}
	}

	if value, ok := data["birth"]; ok {
		if value != nil {
			_, ok := value.(int)
			if !ok {
				ctx.Error(`{"err":"incorrect_birth"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_birth_field"}`, 400)
			return
		}
	}

	if value, ok := data["country"]; ok {
		if value != nil {
			country := value.(string)
			if len(country) > 50 {
				ctx.Error(`{"err":"country_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_country_field"}`, 400)
			return
		}
	}

	if value, ok := data["city"]; ok {
		if value != nil {
			city := value.(string)
			if len(city) > 50 {
				ctx.Error(`{"err":"city_too_long"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_city_field"}`, 400)
			return
		}
	}

	if value, ok := data["joined"]; ok {
		if value != nil {
			_, ok := value.(int)
			if !ok {
				ctx.Error(`{"err":"incorrect_joined"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_joined_field"}`, 400)
			return
		}
	}

	if value, ok := data["status"]; ok {
		if value != nil {
			status := value.(string)
			if len(status) > 0 && status != "свободны" && status != "заняты" && status != "всё сложно" {
				ctx.Error(`{"err":"invalid_status"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_status_field"}`, 400)
			return
		}
	}

	if value, ok := data["interests"]; ok {
		if value != nil {
			var interests []string
			for _, v := range value.([]interface{}) {
				interest := v.(string)
				if len(interest) > 100 {
					ctx.Error(`{"err":"interest_too_long"}`, 400)
					return
				}
				interests = append(interests, interest)
			}
		} else {
			ctx.Error(`{"err":"empty_interests_field"}`, 400)
			return
		}
	}

	if value, ok := data["premium"]; ok {
		if value != nil {
			_, ok := value.(map[string]interface{})
			if !ok {
				ctx.Error(`{"err":"incorrect_premium"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_premium_field"}`, 400)
			return
		}
	}

	if value, ok := data["likes"]; ok {
		if value != nil {
			_, ok := value.([]map[string]interface{})
			if !ok {
				ctx.Error(`{"err":"incorrect_likes_format"}`, 400)
				return
			}
		} else {
			ctx.Error(`{"err":"empty_premium_field"}`, 400)
			return
		}
	}

	// updating in goroutine
	go account.Update(data)

	updatedSuccessResponse(ctx)
	return
}

func updatedSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.SetStatusCode(202)
}
