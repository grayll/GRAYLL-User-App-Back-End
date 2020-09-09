package api

import (
	"context"
	"strings"

	//"strings"

	"encoding/json"
	"strconv"
	"sync"

	//"encoding/json"
	"fmt"

	"log"
	"net/http"

	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
	"cloud.google.com/go/firestore"
	"github.com/SherClockHolmes/webpush-go"

	"github.com/avct/uasurfer"

	"github.com/gin-gonic/gin"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
)

func (h UserHandler) SetStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			UserId            string `json:"userId"`
			Pasword           string `json:"password"`
			StatusId          int    `json:"statusId"`
			LoginStatus       bool   `json:"loginStatus"`
			SignupStatus      bool   `json:"signupStatus"`
			Gry1Status        bool   `json:"gry1Status"`
			Gry2Status        bool   `json:"gry2Status"`
			Gry3Status        bool   `json:"gry3Status"`
			GrzStatus         bool   `json:"grzStatus"`
			Gry1NewPosition   bool   `json:"gry1NewPosition"`
			Gry2NewPosition   bool   `json:"gry2NewPosition"`
			Gry3NewPosition   bool   `json:"gry3NewPosition"`
			GrzNewPosition    bool   `json:"grzNewPosition"`
			SystemStatus      bool   `json:"systemStatus"`
			SystemNewPosition bool   `json:"systemNewPosition"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
			return
		}
		log.Println(input)
		ctx := context.Background()
		fieldName := ""
		value := true
		fieldName1 := ""
		value1 := true
		if input.UserId == "" {
			switch input.StatusId {
			case 1:
				fieldName = "loginStatus"
				value = input.LoginStatus
			case 2:
				fieldName = "signupStatus"
				value = input.SignupStatus
			case 3:
				fieldName = "gry1Status"
				value = input.Gry1Status
				fieldName1 = "gry1NewPosition"
				value1 = input.Gry1NewPosition
			case 4:
				fieldName = "gry2Status"
				value = input.Gry2Status
				fieldName1 = "gry2NewPosition"
				value1 = input.Gry2NewPosition
			case 5:
				fieldName = "gry3Status"
				value = input.Gry3Status
				fieldName1 = "gry3NewPosition"
				value1 = input.Gry3NewPosition
			case 6:
				fieldName = "grzStatus"
				value = input.GrzStatus
				fieldName1 = "grzNewPosition"
				value1 = input.GrzNewPosition
			case 7:
				fieldName = "systemStatus"
				value = input.SystemStatus
				fieldName1 = "systemNewPosition"
				value1 = input.SystemNewPosition
			}
			data := map[string]interface{}{
				fieldName: value,
			}
			if fieldName1 != "" {
				data = map[string]interface{}{
					fieldName:  value,
					fieldName1: value1,
				}
			}
			log.Println("data", data)
			_, err = h.apiContext.Store.Doc("admin/8efngc9fgm12nbcxeq").Set(ctx, data, firestore.MergeAll)
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
				return
			}
		} else {
			switch input.StatusId {
			case 1:
				fieldName = "loginStatus"
				value = input.LoginStatus
			case 2:
				fieldName = "signupStatus"
				value = input.SignupStatus
			case 3:
				fieldName = "gry1Status"
				value = input.Gry1Status
				fieldName1 = "gry1NewPosition"
				value1 = input.Gry1NewPosition
			case 4:
				fieldName = "gry2Status"
				value = input.Gry2Status
				fieldName1 = "gry2NewPosition"
				value1 = input.Gry2NewPosition
			case 5:
				fieldName = "gry3Status"
				value = input.Gry3Status
				fieldName1 = "gry3NewPosition"
				value1 = input.Gry3NewPosition
			case 6:
				fieldName = "grzStatus"
				value = input.GrzStatus
				fieldName1 = "grzNewPosition"
				value1 = input.GrzNewPosition
			case 7:
				fieldName = "systemStatus"
				value = input.SystemStatus
				fieldName1 = "systemNewPosition"
				value1 = input.SystemNewPosition
			}
			data := map[string]interface{}{
				fieldName: value,
			}
			if fieldName1 != "" {
				data = map[string]interface{}{
					fieldName:  value,
					fieldName1: value1,
				}
			}
			_, err := h.apiContext.Store.Doc("users_meta/" + input.UserId).Get(ctx)
			if err != nil {
				GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
				return
			}

			_, err = h.apiContext.Store.Doc("users_meta/"+input.UserId).Set(ctx, data, firestore.MergeAll)
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
				return
			}

		}
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) GetUsersMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := strconv.Atoi(c.Param("cursor"))
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, err.Error())
			return
		}
		limit := 250
		users := make([]map[string]interface{}, 0)
		ctx := context.Background()
		usersColl := h.apiContext.Store.Collection("users_meta")

		if cursor == 0 {
			userDocs, err := usersColl.OrderBy("CreatedAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
				return
			}

			for _, doc := range userDocs {
				// if strings.Contains(doc.Data()["PublicKey"].(string), "GAUBL3") {
				// 	log.Println("uid:", doc.Data()["UserId"].(string))
				// }
				users = append(users, doc.Data())
			}
		} else {
			userDocs, err := usersColl.OrderBy("CreatedAt", firestore.Desc).StartAfter(cursor).Limit(limit).Documents(ctx).GetAll()
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
				return
			}

			for _, doc := range userDocs {
				users = append(users, doc.Data())
			}
		}
		c.JSON(http.StatusOK, map[string]interface{}{"errCode": SUCCESS, "usersMeta": users})

	}
}
func (h UserHandler) GetUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		searchStr := c.Param("searchStr")
		log.Println(searchStr)

		userData := make(map[string]interface{})
		errCode := NOT_FOUND
		if strings.Contains(searchStr, "@") {
			doc, err := h.apiContext.Store.Collection("users_meta").Where("Email", "==", searchStr).Documents(ctx).GetAll()
			log.Println(err)
			if len(doc) > 0 {
				userData = doc[0].Data()
				errCode = SUCCESS
			}
		} else if strings.Contains(searchStr, "G") && len(searchStr) == 56 {
			doc, _ := h.apiContext.Store.Collection("users_meta").Where("PublicKey", "==", searchStr).Documents(ctx).GetAll()

			if len(doc) > 0 {
				userData = doc[0].Data()
				errCode = SUCCESS
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": errCode, "userData": userData,
		})

	}
}
func (h UserHandler) LoginAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := new(models.UserLogin)
		err := c.BindJSON(user)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
			return
		}
		// Validate user data
		if !user.Validate() {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
			return
		}

		if user.Email != "huykbc@gmail.com" && user.Email != "grayll@grayll.io" {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		userInfo, uid := GetUserByField(h.apiContext.Store, "Email", user.Email)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		// Check password
		ret, err := utils.VerifyPassphrase(user.Password, userInfo["HashPassword"].(string))
		if !ret || err != nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		if !userInfo["IsVerified"].(bool) {
			GinRespond(c, http.StatusOK, UNVERIFIED, "Email is not verified")
			return
		}
		// var gd geoIPData
		// gd.Country = c.GetHeader("X-AppEngine-Country")
		// gd.Region = c.GetHeader("X-AppEngine-Region")
		// gd.City = c.GetHeader("X-AppEngine-City")
		// log.Println("GeoIp data:", gd)

		currentIp := c.ClientIP()
		setting, ok := userInfo["Setting"].(map[string]interface{})
		if !ok {
			log.Println("Can not parse user setting. userInfo: ", userInfo)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}

		city, country := utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
		ua := uasurfer.Parse(c.Request.UserAgent())

		agent := fmt.Sprintf("Device - %s, Browser - %s, OS - %s.", ua.DeviceType.StringTrimPrefix(), ua.Browser.Name.StringTrimPrefix(), ua.OS.Name.StringTrimPrefix())
		ctx := context.Background()

		wg := new(sync.WaitGroup)
		wg.Add(1)
		isConfirmIp := false
		go func() {
			defer wg.Done()
			ipConfirm := setting["IpConfirm"].(bool)
			if ipConfirm {
				if currentIp != userInfo["Ip"].(string) {
					secondIp, ok := userInfo["SecondIp"]

					// secondIp still may not be set
					if !ok || (ok && currentIp != secondIp.(string)) {
						// Send confirm Ip mail
						encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, currentIp+"?"+uid)
						if encodeStr == "" {
							GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not login right now. Please try again later.")
							isConfirmIp = true
							return
						}

						mores := map[string]string{
							"loginTime": time.Now().Format("Mon, 02 Jan 2006 15:04:05 UTC"),
							"ip":        currentIp,
							"agent":     c.Request.UserAgent(),
							"city":      city,
							"country":   country,
						}
						err = mail.SendMail(userInfo["Email"].(string), userInfo["Name"].(string), ConfirmIpSub, ConfirmIp, encodeStr, h.apiContext.Config.Host, mores)
						if err != nil {
							GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not login right now. Please try again later.")
							isConfirmIp = true
							return
						}
						GinRespond(c, http.StatusOK, IP_CONFIRM, "Need to confirm Ip before login")

						// Send app and push notices
						title := "GRAYLL | IP Address Verification"
						body := fmt.Sprintf("This IP address %s is unknown! An IP address verification link has been sent to your email.", currentIp)
						notice := map[string]interface{}{
							"type":    "general",
							"title":   title,
							"isRead":  false,
							"body":    body,
							"time":    time.Now().Unix(),
							"vibrate": []int32{100, 50, 100},
							"icon":    "https://app.grayll.io/favicon.ico",
							"data": map[string]interface{}{
								"url": h.apiContext.Config.Host + "/notifications/overview",
							},
						}

						go func() {
							subs, err := h.apiContext.Cache.GetUserSubs(uid)
							if err == nil && subs != "" {
								//log.Println("subs: ", subs)
								noticeData := map[string]interface{}{
									"notification": notice,
								}
								webpushSub := webpush.Subscription{}
								err = json.Unmarshal([]byte(subs), &webpushSub)
								if err != nil {
									log.Println("Unmarshal subscription from redis error: ", err)
									return
								}
								err = PushNotice(noticeData, &webpushSub)
								if err != nil {
									log.Println("PushNotice error: ", err)
									//return
								}
							}
						}()

						// Save to firestore
						docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
						_, err = docRef.Set(ctx, notice)
						if err != nil {
							log.Println("SaveNotice error: ", err)
							return
						}
						// Set unread general
						_, err = h.apiContext.Store.Doc("users_meta/"+uid).Update(ctx, []firestore.Update{
							{Path: "UrGeneral", Value: firestore.Increment(1)},
						})
						if err != nil {
							log.Println("SaveNotice update error: ", err)
							//return
						}
						isConfirmIp = true
					}
				}
			}
		}()
		wg.Wait()
		if isConfirmIp {
			return
		}

		go func() {
			h.apiContext.Cache.SetPublicKey(uid, userInfo["PublicKey"].(string))
			settingFields := []string{"IpConfirm", "MulSignature", "AppGeneral", "AppWallet", "AppAlgo", "MailGeneral", "MailWallet", "MailAlgo"}
			for _, field := range settingFields {
				if val, ok := setting[field]; ok {
					h.apiContext.Cache.SetNotice(uid, field, val.(bool))
				}
			}
			if subs, ok := userInfo["Subs"]; ok {
				log.Println("Subs:", subs)
				s, err := json.Marshal(subs)
				if err != nil {
					log.Println("Can not find parse subs:", err)
				}
				h.apiContext.Cache.SetUserSubs(uid, string(s))

				if _, ok := userInfo["Subs"]; ok {
					userInfo["Subs"] = true
				}
			}
		}()

		// First login time, send mail notice
		go func() {
			if city == "" && country == "" {
				city, country = utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
			}
			mores := map[string]string{
				"loginTime": time.Now().Format("Mon, 02 Jan 2006 15:04:05 UTC"),
				"ip":        currentIp,
				"agent":     agent,
				"city":      city,
				"country":   country,
			}
			err = mail.SendLoginNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), LoginSuccess, mores)
			if err != nil {
				log.Println("Can not send login notice mail:", err)
			}
		}()

		tokenStr, err := h.apiContext.Jwt.GenToken(uid, 24*60)
		// localKey used by client to encrypt secret key and store encrypted secret key on local
		localKey := randStr(32, "alphanum")
		hashToken := Hash(tokenStr)

		go func() {
			// Set local key in redis, getUserInfo will get from redis cache
			h.apiContext.Cache.Client.Set(hashToken, localKey, time.Hour*24)
			h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{
				"LoginTime": time.Now().Unix(),
			}, firestore.MergeAll)

			// store login history
			docRef := h.apiContext.Store.Collection("logins").NewDoc()
			_, err := docRef.Set(ctx, map[string]interface{}{
				"Uid":       uid,
				"LoginTime": time.Now().Unix(),
				"Device":    ua.DeviceType.StringTrimPrefix(),
				"Country":   country,
			})
			if err != nil {
				log.Println("Can not create logins document:", err)
			}

		}()

		userBasicInfo := make(map[string]interface{})
		userBasicInfo["Tfa"] = false
		if _, ok := userInfo["Tfa"]; ok {
			tfaData := userInfo["Tfa"].(map[string]interface{})
			if tfaEnable, ok := tfaData["Enable"]; ok {
				userBasicInfo["Tfa"] = tfaEnable.(bool)
				if tfaEnable.(bool) {
					if _, ok := tfaData["Expire"]; ok {
						userBasicInfo["Expire"] = tfaData["Expire"]
					} else {
						userBasicInfo["Expire"] = 0
					}
				}
			}
		} else {
			userBasicInfo["Expire"] = 0
		}

		userBasicInfo["LoanPaidStatus"] = userInfo["LoanPaidStatus"].(int64)
		userBasicInfo["EnSecretKey"] = userInfo["EnSecretKey"]
		userBasicInfo["SecretKeySalt"] = userInfo["SecretKeySalt"]
		userBasicInfo["PublicKey"] = userInfo["PublicKey"]
		userBasicInfo["Setting"] = setting
		userBasicInfo["Uid"] = uid
		userInfo["Uid"] = uid
		userBasicInfo["LocalKey"] = localKey

		// check HMAC hex string for Intercom
		_hmc := ""
		if hmac, ok := userInfo["Hmac"]; !ok {
			_hmc = Hmac("kFOLecggKkSgaWGn_dyoFzZyuY8wFtzkvcncIU-J", userInfo["Email"].(string))
			userInfo["Hmac"] = _hmc
			_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
				"Hmac": _hmc,
			}, firestore.MergeAll)
		} else {
			_hmc = hmac.(string)
		}

		tokeExpTime := time.Now().Unix() + TokeExpiredTime
		userMeta := map[string]interface{}{"UrWallet": 0, "UrGRY1": 0, "UrGRY2": 0, "UrGRY3": 0, "UrGRZ": 0, "UrGeneral": 0, "OpenOrders": 0, "OpenOrdersGRX": 0,
			"OpenOrdersXLM": 0, "GRX": 0, "XLM": 0, "TokenExpiredTime": tokeExpTime}
		// set user meta data if account created before 7-Jan-2020
		snapShot, err := h.apiContext.Store.Doc("users_meta/" + uid).Get(context.Background())
		if err != nil {
			log.Println(uid+": Can not get users_meta error %v\n", err)
			_, err = h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), userMeta)
			if err != nil {
				log.Println(uid+": Set users_meta data error %v\n", err)
			}
		} else {
			userMeta = snapShot.Data()
			h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), map[string]interface{}{"TokenExpiredTime": tokeExpTime}, firestore.MergeAll)
		}
		userMeta["TokenExpiredTime"] = tokeExpTime

		grxP, err := h.apiContext.Cache.GetGRXPrice()
		if err != nil {
			grxP = "1"
		}
		xlmP, err := h.apiContext.Cache.GetXLMPrice()
		if err != nil {
			xlmP = "1"
		}
		userMeta["XlmP"] = xlmP
		userMeta["GrxP"] = grxP

		delete(userInfo, "LoanPaidStatus")
		delete(userInfo, "HashPassword")
		delete(userInfo, "EnSecretKey")
		delete(userInfo, "SecretKeySalt")
		delete(userInfo, "Setting")
		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS, "user": userInfo, "userMeta": userMeta, "userBasicInfo": userBasicInfo, "token": tokenStr, "tokenExpiredTime": tokeExpTime,
		})
	}
}
