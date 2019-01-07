package main

import (
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"github.com/mailru/easyjson/buffer"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

type AccountResponse map[string]interface{}

var (
	inverseIntComparator = func(a, b interface{}) int {
		return -utils.IntComparator(a, b)
	}
	inverseFloat64Comparator = func(a, b interface{}) int {
		return -utils.Float64Comparator(a, b)
	}
	accountMap     = treemap.NewWith(inverseIntComparator)
	countryMap     = map[string]*treemap.Map{}
	cityMap        = map[string]*treemap.Map{}
	birthYearMap   = map[int]*treemap.Map{}
	fnameMap       = map[string]*treemap.Map{}
	snameMap       = map[string]*treemap.Map{}
	similarityMap  = map[int]*treemap.Map{}
	globalLikesMap = map[*Account][]*Account{}
)

type Account struct {
	ID        int            `json:"id"`
	Email     string         `json:"email"`
	Fname     string         `json:"fname"`
	Sname     string         `json:"sname"`
	Phone     string         `json:"phone"`
	Sex       string         `json:"sex"`
	Birth     int            `json:"birth"`
	Country   string         `json:"country"`
	City      string         `json:"city"`
	Joined    int            `json:"joined"`
	Status    string         `json:"status"`
	Interests []string       // temp data
	Premium   map[string]int `json:"premium"`
	//TempLikes []map[string]int `json:"likes"` // temp data

	interestsMap  map[string]struct{} // try map[string]struct{}{}
	emailDomain   string
	phoneCode     int
	birthYear     int
	premiumFinish int64
	uniqLikes     map[int]int
}

func createAccount(acc Account) {
	// TODO: unset uniqLikes, interests

	if len(acc.Interests) > 0 {
		acc.interestsMap = make(map[string]struct{})
		for _, interest := range acc.Interests {
			acc.interestsMap[interest] = struct{}{}
		}
		acc.Interests = nil
	}

	if acc.Email != "" {
		components := strings.Split(acc.Email, "@")
		if len(components) > 0 {
			acc.emailDomain = components[1]
		}
	}

	if acc.Phone != "" {
		phoneCodeStr := strings.SplitN(strings.SplitN(acc.Phone, "(", 2)[1], ")", 2)[0]
		if phoneCode, err := strconv.Atoi(phoneCodeStr); err == nil {
			acc.phoneCode = phoneCode
		}
	}

	if acc.Birth != 0 {
		tm := time.Unix(int64(acc.Birth), 0)
		acc.birthYear = tm.Year()
	}
	if finish, ok := acc.Premium["finish"]; ok {
		acc.premiumFinish = int64(finish)
	}

	//if len(acc.TempLikes) > 0 {
	//	likesMap := make(map[int][]int, 0)
	//	for _, like := range acc.TempLikes {
	//		likesMap[like["id"]] = append(likesMap[like["id"]], like["ts"])
	//
	//		//globalLikesMap[accId] = append(globalLikesMap[accId], &acc)
	//	}
	//	acc.TempLikes = nil
	//
	//	uniqLikeMap := map[int]int{}
	//	for id, timestamps := range likesMap {
	//		var ts int
	//		if len(timestamps) > 1 {
	//			var total = 0
	//			for _, value := range timestamps {
	//				total += value
	//			}
	//			ts = total / int(len(timestamps))
	//		} else {
	//			ts = timestamps[0]
	//		}
	//		uniqLikeMap[id] = ts
	//	}
	//	//acc.uniqLikes = uniqLikeMap
	//}

	if acc.Country != "" {
		if _, ok := countryMap[acc.Country]; !ok {
			countryMap[acc.Country] = treemap.NewWith(inverseIntComparator)
		}
		countryMap[acc.Country].Put(acc.ID, &acc)
	}
	if acc.City != "" {
		if _, ok := cityMap[acc.City]; !ok {
			cityMap[acc.City] = treemap.NewWith(inverseIntComparator)
		}
		cityMap[acc.City].Put(acc.ID, &acc)
	}
	if acc.birthYear > 0 {
		if _, ok := birthYearMap[acc.birthYear]; !ok {
			birthYearMap[acc.birthYear] = treemap.NewWith(inverseIntComparator)
		}
		birthYearMap[acc.birthYear].Put(acc.ID, &acc)
	}
	if acc.Fname != "" {
		if _, ok := fnameMap[acc.Fname]; !ok {
			fnameMap[acc.Fname] = treemap.NewWith(inverseIntComparator)
		}
		fnameMap[acc.Fname].Put(acc.ID, &acc)
	}
	if acc.Sname != "" {
		if _, ok := snameMap[acc.Sname]; !ok {
			snameMap[acc.Sname] = treemap.NewWith(inverseIntComparator)
		}
		snameMap[acc.Sname].Put(acc.ID, &acc)
	}

	accountMap.Put(acc.ID, &acc)
}

func calculateSimilarityIndex() {
	accountMap.Each(func(key interface{}, value interface{}) {
		calculateSimilarityForUser(value.(*Account))
	})
	// calculate for ine user
	//value, _ := accountMap.Get(24156)
	//calculateSimilarityForUser(value.(*Account))
}

func calculateSimilarityForUser(account *Account) {
	//user1Likes := account.uniqLikes
	//for likeId, ts1 := range user1Likes {
	//	for _, acc2 := range globalLikesMap[likeId] {
	//		ts2 := acc2.uniqLikes[likeId]
	//		var similarity float64
	//		if ts1 == ts2 {
	//			similarity += 1
	//		} else {
	//			similarity += 1 / math.Abs(float64(ts1-ts2))
	//		}
	//		if similarity > 0 {
	//			user1Id := account.ID
	//			if _, ok := similarityMap[user1Id]; !ok {
	//				similarityMap[user1Id] = treemap.NewWith(inverseFloat64Comparator)
	//			}
	//			similarityMap[user1Id].Put(similarity, acc2)
	//		}
	//	}
	//}
}

func hashFunc(str string) uint32 {
	return crc32.Checksum([]byte(str), crc32q)
}

func accountToJsonBytes(acc Account) []byte {
	resultBuf := buffer.Buffer{}

	//out.RawByte('{')
	resultBuf.AppendByte('{')

	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		if first {
			first = false
			resultBuf.AppendString(prefix[1:])
		} else {
			resultBuf.AppendString(prefix)
		}
		//out.Int(int(acc.ID))
		strconv.AppendInt(resultBuf.Buf, int64(acc.ID), 10)
	}
	{
		const prefix string = ",\"email\":"
		if first {
			first = false
			resultBuf.AppendString(prefix[1:])
		} else {
			resultBuf.AppendString(prefix)
		}
		resultBuf.AppendString("\"" + acc.Email + "\"")
		//out.String(string(acc.Email))
	}
	//{
	//	const prefix string = ",\"fname\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Fname))
	//}
	//{
	//	const prefix string = ",\"sname\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Sname))
	//}
	//{
	//	const prefix string = ",\"phone\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Phone))
	//}
	//{
	//	const prefix string = ",\"sex\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Sex))
	//}
	//{
	//	const prefix string = ",\"birth\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.Int(int(acc.Birth))
	//}
	//{
	//	const prefix string = ",\"country\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Country))
	//}
	//{
	//	const prefix string = ",\"city\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.City))
	//}
	//{
	//	const prefix string = ",\"joined\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.Int(int(acc.Joined))
	//}
	//{
	//	const prefix string = ",\"status\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	out.String(string(acc.Status))
	//}
	//{
	//	const prefix string = ",\"Interests\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	if acc.Interests == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
	//		resultBuf.AppendString("null")
	//	} else {
	//		out.RawByte('[')
	//		for v3, v4 := range acc.Interests {
	//			if v3 > 0 {
	//				out.RawByte(',')
	//			}
	//			out.String(string(v4))
	//		}
	//		out.RawByte(']')
	//	}
	//}
	//{
	//	const prefix string = ",\"premium\":"
	//	if first {
	//		first = false
	//		resultBuf.AppendString(prefix[1:])
	//	} else {
	//		resultBuf.AppendString(prefix)
	//	}
	//	if acc.Premium == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
	//		resultBuf.AppendString(`null`)
	//	} else {
	//		out.RawByte('{')
	//		v5First := true
	//		for v5Name, v5Value := range acc.Premium {
	//			if v5First {
	//				v5First = false
	//			} else {
	//				out.RawByte(',')
	//			}
	//			out.String(string(v5Name))
	//			out.RawByte(':')
	//			out.Int(int(v5Value))
	//		}
	//		out.RawByte('}')
	//	}
	//}

	//out.RawByte('}')
	resultBuf.AppendByte('}')

	return resultBuf.BuildBytes()
}
