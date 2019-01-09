package main

import (
	"github.com/emirpasic/gods/maps/treemap"

	"github.com/valyala/fasthttp"
)

type NamedIndex struct {
	name  []byte
	index *treemap.Map
}

func (n NamedIndex) New(name []byte, index *treemap.Map) *NamedIndex {
	return &NamedIndex{name, index}
}

func (n *NamedIndex) Update(name []byte, index *treemap.Map) *NamedIndex {
	n.name = name
	n.index = index

	return n
}

func emptyResponse(ctx *fasthttp.RequestCtx) {
	ctx.Success("application/json", []byte(`{"accounts":[]}`))
}

func prepareResponseBytes(found []*Account, responseProperties []string) []byte {
	vbuf := bytesPool.Get()
	bytesBuffer := vbuf.([]byte)

	bytesBuffer = append(bytesBuffer, `{"accounts":[`...)

	keysLen := len(responseProperties)
	foundLen := len(found)

	for accIdx, account := range found {
		lastAcc := accIdx == foundLen-1
		_ = lastAcc
		for keyIdx, key := range responseProperties {
			firstKey := keyIdx == 0
			lastKey := keyIdx == keysLen-1
			_ = lastKey
			_ = firstKey

			if firstKey {
				bytesBuffer = append(bytesBuffer, `{`...)
			}

			switch key {
			case "id":
				bytesBuffer = append(bytesBuffer, `"id":`...)
				bytesBuffer = fasthttp.AppendUint(bytesBuffer, account.ID)
			case "sex":
				bytesBuffer = append(bytesBuffer, `"sex":"`+account.Sex+`"`...)
			case "email":
				bytesBuffer = append(bytesBuffer, `"email":"`+account.Email+`"`...)
			case "status":
				bytesBuffer = append(bytesBuffer, `"status":"`+account.Status+`"`...)
			case "fname":
				bytesBuffer = append(bytesBuffer, `"fname":"`+account.Fname+`"`...)
			case "sname":
				bytesBuffer = append(bytesBuffer, `"sname":"`+account.Sname+`"`...)
			case "phone":
				bytesBuffer = append(bytesBuffer, `"phone":"`+account.Phone+`"`...)
			case "country":
				bytesBuffer = append(bytesBuffer, `"country":"`+account.Country+`"`...)
			case "city":
				bytesBuffer = append(bytesBuffer, `"city":"`+account.City+`"`...)
			case "birth":
				bytesBuffer = append(bytesBuffer, `"birth":`...)
				bytesBuffer = fasthttp.AppendUint(bytesBuffer, account.Birth)
			case "premium":
				if account.Premium == nil {
					bytesBuffer = append(bytesBuffer, `"premium":null`...)
				} else {
					bytesBuffer = append(bytesBuffer, `"premium":{"start":`...)
					bytesBuffer = fasthttp.AppendUint(bytesBuffer, account.Premium["start"])
					bytesBuffer = append(bytesBuffer, `,"finish":`...)
					bytesBuffer = fasthttp.AppendUint(bytesBuffer, account.Premium["finish"])
					bytesBuffer = append(bytesBuffer, `}`...)
				}
			}

			if lastKey {
				bytesBuffer = append(bytesBuffer, `}`...)
			} else {
				bytesBuffer = append(bytesBuffer, `,`...)
			}
		}

		if !lastAcc {
			bytesBuffer = append(bytesBuffer, `,`...)
		}
	}

	bytesBuffer = append(bytesBuffer, `]}`...)

	bytesPool.Put(vbuf)

	return bytesBuffer
}
