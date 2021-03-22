package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"sync"

	//"encoding/json"
	"fmt"
	"strings"

	"log"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
	"cloud.google.com/go/firestore"

	//"github.com/SherClockHolmes/webpush-go"
	"github.com/asaskevich/govalidator"
	"github.com/avct/uasurfer"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	stellar "github.com/huyntsgs/stellar-service"
	build "github.com/stellar/go/txnbuild"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"

	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	ConfirmRegistrationSub = "Please Confirm Your GRAYLL App Registration Request"
	ConfirmIpSub           = "GRAYLL | New IP Address Verification"
	LoginSuccess           = "GRAYLL | Account Login Successful"
	ResetPasswordSub       = "GRAYLL | Reset Password Verification"
	ChangeEmailSub         = "GRAYLL | Confirm change email request"
	RevealSecretTokenSub   = "GRAYLL | Reveal Secret Key"

	VerifyEmail       = "verifyEmail"
	ResetPassword     = "resetPassword"
	ChangeEmail       = "changeEmail"
	ConfirmIp         = "confirmIp"
	UID               = "Uid"
	RevealSecretToken = "revealSecretToken"
	TokeExpiredTime   = 24*60*60 - 2
	//TokeExpiredTime = 3 * 60
)

type UserHandler struct {
	apiContext *ApiContext
}

type geoIPData struct {
	Country string
	Region  string
	City    string
}

// Creates new UserHandler.
// UserHandler accepts interface UserStore.
// Any data store implements UserStore could be the input of the handle.
func NewUserHandler(apiContext *ApiContext) UserHandler {
	return UserHandler{apiContext: apiContext}
}

// Login handles login router.
// Function validates parameters and call Login from UserStore.
func (h UserHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		adminDoc, err := h.apiContext.Store.Doc("admin/8efngc9fgm12nbcxeq").Get(ctx)
		if err != nil {
			log.Println("Can not get admin doc:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
		}

		loginStatus := adminDoc.Data()["loginStatus"].(bool)
		if !loginStatus {
			log.Println("[HACK]-login is blocked", c.ClientIP(), err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "")
			return
		}

		currentIp := c.ClientIP()
		city, country := utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
		if country == "Indonesia" {
			log.Println("ERROR - LOGIN BOT - Indonesia bot", city, currentIp)
			//GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			//return
		}

		if h.apiContext.BlockIPs.Has(currentIp) {
			log.Println("ERROR - LOGIN blocked ip", currentIp, city)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}

		ua := uasurfer.Parse(c.Request.UserAgent())

		agent := fmt.Sprintf("Device - %s, Browser - %s, OS - %s.", ua.DeviceType.StringTrimPrefix(), ua.Browser.Name.StringTrimPrefix(), ua.OS.Name.StringTrimPrefix())

		user := new(models.UserLogin)
		err = c.BindJSON(user)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
			return
		}
		if !h.apiContext.Cache.CheckRecapchaToken(user.Email + "login") {
			log.Println("ERROR - HACK LOGIN BOT - bypass recapcha - not block IP", currentIp, city, user.Email)
			//h.apiContext.BlockIPs.Set(currentIp, "")
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		// Validate user data
		if !user.Validate() {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Data is invalid")
			return
		}

		userInfo, uid := GetUserLogin(h.apiContext.Store, "Email", user.Email)
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

		setting, ok := userInfo["Setting"].(map[string]interface{})
		if !ok {
			log.Println("Can not parse user setting. userInfo: ", userInfo)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}

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
							"type":   "general",
							"title":  title,
							"isRead": false,
							"body":   body,
							"time":   time.Now().Unix(),
							// "vibrate": []int32{100, 50, 100},
							// "icon":    "https://app.grayll.io/favicon.ico",
							// "data": map[string]interface{}{
							// 	"url": h.apiContext.Config.Host + "/notifications/overview",
							// },
						}

						// go func() {
						// 	subs, err := h.apiContext.Cache.GetUserSubs(uid)
						// 	if err == nil && subs != "" {
						// 		//log.Println("subs: ", subs)
						// 		noticeData := map[string]interface{}{
						// 			"notification": notice,
						// 		}
						// 		webpushSub := webpush.Subscription{}
						// 		err = json.Unmarshal([]byte(subs), &webpushSub)
						// 		if err != nil {
						// 			log.Println("Unmarshal subscription from redis error: ", err)
						// 			return
						// 		}
						// 		err = PushNotice(noticeData, &webpushSub)
						// 		if err != nil {
						// 			log.Println("PushNotice error: ", err)
						// 			//return
						// 		}
						// 	}
						// }()

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
			// if subs, ok := userInfo["Subs"]; ok {
			// 	log.Println("Subs:", subs)
			// 	s, err := json.Marshal(subs)
			// 	if err != nil {
			// 		log.Println("Can not find parse subs:", err)
			// 	}
			// 	h.apiContext.Cache.SetUserSubs(uid, string(s))

			// 	if _, ok := userInfo["Subs"]; ok {
			// 		userInfo["Subs"] = true
			// 	}
			// }
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
		// _hmc := ""
		// if hmac, ok := userInfo["Hmac"]; !ok {
		// 	_hmc = Hmac("kFOLecggKkSgaWGn_dyoFzZyuY8wFtzkvcncIU-J", userInfo["Email"].(string))
		// 	userInfo["Hmac"] = _hmc
		// 	_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
		// 		"Hmac": _hmc,
		// 	}, firestore.MergeAll)
		// } else {
		// 	_hmc = hmac.(string)
		// }

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

func (h UserHandler) Renew() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Password string
		}
		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input")
			return
		}
		uid := c.GetString("Uid")
		if input.Password != "" {
			// Check password
			userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
			ret, err := utils.VerifyPassphrase(input.Password, userInfo["HashPassword"].(string))
			if !ret || err != nil {
				GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
				return
			}
		}

		tokenStr, _ := h.apiContext.Jwt.GenToken(uid, 24*60)
		tokeExpTime := time.Now().Unix() + TokeExpiredTime
		h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), map[string]interface{}{"TokenExpiredTime": tokeExpTime}, firestore.MergeAll)
		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS, "token": tokenStr, "tokenExpiredTime": tokeExpTime,
		})
	}
}

func (h UserHandler) VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {

		email := c.Param("email")
		if email == "" {
			GinRespond(c, http.StatusOK, EMAIL_INVALID, "Email address is invalid")
			return
		}
		err := VerifyEmailNeverBounce(h.apiContext.Config.NeverBounceApiKey, email)
		if err != nil {
			GinRespond(c, http.StatusOK, EMAIL_INVALID, "Email address is invalid")
			return
		}
		GinRespond(c, http.StatusOK, SUCCESS, "")
		return

	}
}

func (h UserHandler) VerifyRecapchaToken() gin.HandlerFunc {
	return func(c *gin.Context) {

		var respApi = make(map[string]string)
		respApi["status"] = "fail"

		var respData struct {
			Success      bool      `json:"success"`      // whether this request was a valid reCAPTCHA token for your site
			Score        float64   `json:"score"`        // the score for this request (0.0 - 1.0)
			Action       string    `json:"action"`       // the action name for this request (important to verify)
			Challenge_ts time.Time `json:"challenge_ts"` // timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
			Hostname     string    `json:"hostname"`     // the hostname of the site where the reCAPTCHA was solved

		}

		token, err := ExtractToken(c.Request)
		if err != nil {
			fmt.Println("VerifyRecapchaToken: Authorization header does not contain Bearer\n", err)
			//respData.Success = false
			// w.WriteHeader(http.StatusUnauthorized)
			// json.NewEncoder(w).Encode(respData)
			//respApi["status"] = "fail"
			c.JSONP(http.StatusUnauthorized, respApi)
			return
		}

		email := c.Param("email")
		action := c.Param("action")

		log.Println("ERROR VerifyRecapchaToken - ", email, action)

		if email == "" || action == "" {
			c.JSON(http.StatusUnauthorized, respApi)
			return
		}

		url := "https://www.google.com/recaptcha/api/siteverify"
		secret := "6LfYI7EUAAAAAKGxMquwzN5EsJHlp-0_bfspQhGI"
		url = fmt.Sprintf("%s?secret=%s&response=%s", url, secret, token)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
		if err != nil {
			log.Println("verifyRecapchaToken: Can not create new req")
			//respApi["status"] = "fail"
			// w.WriteHeader(http.StatusInternalServerError)
			// json.NewEncoder(w).Encode(&respApi)
			c.JSON(http.StatusUnauthorized, respApi)
			return
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("verifyRecapchaToken: call client.Do() error %v\n", err)
			c.JSONP(http.StatusUnauthorized, respApi)
			return
		}
		//defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&respData)
		if err != nil {
			log.Printf("verifyRecapchaToken: call resp.Body error %v\n", err)
			c.JSONP(http.StatusUnauthorized, respApi)
			return
		}

		if respData.Score > 0.5 {
			respApi["status"] = "success"
		}
		if action == "login" || action == "register" {
			h.apiContext.Cache.SetRecapchaToken(email+action, token)
		} else {
			respApi["status"] = "fail"
		}

		c.JSONP(http.StatusOK, respApi)
	}
}

// Register handles register router.
// Function validates parameters and call Register from UserStore.
func (h UserHandler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.RegistrationInfo
		currentIp := c.ClientIP()
		city, country := utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
		if country == "Indonesia" {
			log.Println("ERROR - REGISTER BOT - Indonesia bot", city, input.Email, currentIp)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}

		if h.apiContext.BlockIPs.Has(currentIp) {
			log.Println("ERROR -REGISTER - blocked ip", currentIp, country)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}

		ctx := context.Background()
		adminDoc, err := h.apiContext.Store.Doc("admin/8efngc9fgm12nbcxeq").Get(ctx)
		if err != nil {
			log.Println("Can not get admin doc:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		signupStatus := adminDoc.Data()["signupStatus"].(bool)
		if !signupStatus {
			log.Println("[HACK]-singup is blocked", c.ClientIP(), err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "")
			return
		}

		err = c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		// if strings.Count(input.Email, ".") >= 2 {
		// 	log.Println("ERROR - REGISTER BOT - Email has two dot", city, input.Email, currentIp)
		// 	GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
		// 	return
		// }

		if !h.apiContext.Cache.CheckRecapchaToken(input.Email + "register") {
			log.Println("ERROR - HACK REGISTER BOT - bypass recapcha - blocked IP", city, input.Email, currentIp)
			h.apiContext.BlockIPs.Set(currentIp, "")
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}
		//log.Println("input:", input)
		// Validate user data
		if !(input.Validate()) {
			log.Println("Validate err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Input data is invalid")
			return
		}
		// err = VerifyEmailNeverBounce(h.apiContext.Config.NeverBounceApiKey, input.Email)
		// if err != nil {
		// 	GinRespond(c, http.StatusOK, EMAIL_INVALID, "Email address is invalid")
		// 	return
		// }
		userInfo, _ := GetUserByField(h.apiContext.Store, "Email", input.Email)
		if userInfo != nil {
			GinRespond(c, http.StatusOK, EMAIL_IN_USED, "Email already registered")
			return
		}

		// Get IP of user at time registration
		hmc := Hmac("kFOLecggKkSgaWGn_dyoFzZyuY8wFtzkvcncIU-J", input.Email)
		input.Federation = input.Email + "*grayll.io"
		input.LoanPaidStatus = 0
		input.Ip = c.ClientIP()
		input.CreatedAt = time.Now().Unix()
		input.Setting = models.Settings{IpConfirm: true, MulSignature: false, AppAlgo: true, AppWallet: true, AppGeneral: true,
			MailAlgo: true, MailWallet: true, MailGeneral: true}
		hash, err := utils.DerivePassphrase(input.HashPassword, 32)
		if err != nil {
			log.Printf("AddUserData:Add error %v\n", err)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}
		input.HashPassword = hash
		input.Hmac = hmc
		batch := h.apiContext.Store.Batch()
		userDoc := h.apiContext.Store.Collection("users").NewDoc()

		// docRef, _, err := h.apiContext.Store.Collection("users").Add(ctx, input)
		// if err != nil {
		// 	log.Printf("AddUserData:Add error %v\n", err)
		// 	GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
		// 	return
		// }

		uid := userDoc.ID // Check referer user
		refererUid := ""
		referralUGeneral := 0
		referralSendGridId := ""
		if input.Referer != "" {
			uidb, err := base64.StdEncoding.DecodeString(input.Referer)
			if err != nil {
				log.Println("[ERROR] - Register - can not decode referer:", input.Email, err)
			} else {
				refererUid = string(uidb)
				referer, err := h.apiContext.Store.Doc("users/" + refererUid).Get(ctx)
				if err != nil {
					log.Println("[ERROR] - get referer uid:", err)
				} else {
					t := time.Now().Unix()

					refererDoc := h.apiContext.Store.Doc("referrals/" + uid + "/referer/" + refererUid)
					batch.Set(refererDoc, map[string]interface{}{
						"time":   t,
						"name":   referer.Data()["Name"],
						"lname":  referer.Data()["LName"],
						"email":  referer.Data()["Email"],
						"uid":    refererUid,
						"feeGRX": 0,
					})

					referralData := map[string]interface{}{
						"time":         t,
						"name":         input.Name,
						"lname":        input.LName,
						"email":        input.Email,
						"totalFeeGRX":  0,
						"totalPayment": 0,
					}
					invitedDate := ""
					if input.DocId != "" {
						invitedDoc, err := h.apiContext.Store.Doc("referrals/" + refererUid + "/invite/" + input.DocId).Get(ctx)
						if err == nil && invitedDoc != nil {
							log.Println("get from invite data:")
							referralData := invitedDoc.Data()

							invitedDate = time.Unix(invitedDoc.Data()["sentRemind"].(int64), 0).Format("01-02-2006")

							delete(referralData, "remindTime")
							delete(referralData, "status")
							delete(referralData, "lastSentRemind")
							referralData["time"] = time.Now().Unix()
							referralData["totalFeeGRX"] = 0
							referralData["totalPayment"] = 0

						}
					}
					//log.Println("referralData:", referralData)

					referralData["uid"] = uid
					referralDoc := h.apiContext.Store.Doc("referrals/" + refererUid + "/referral/" + uid)
					batch.Set(referralDoc, referralData)

					// Get invited sendgrid id
					invitedDoc, err := h.apiContext.Store.Doc("referrals/" + refererUid + "/invite/" + input.DocId).Get(ctx)
					if err == nil {
						if val, ok := invitedDoc.Data()["sendGridId"]; ok {
							referralSendGridId = val.(string)
						}

						batch.Delete(invitedDoc.Ref)
					}

					// Check whether metrics data exist
					metricDoc, err := h.apiContext.Store.Doc("referrals/" + refererUid + "/metrics/referral").Get(ctx)
					//log.Println("metricDoc, err:", err)
					if err != nil && grpc.Code(err) == codes.NotFound {
						// already exist
						log.Println("!existed referral")
						metricDoc := h.apiContext.Store.Doc("referrals/" + refererUid + "/metrics/referral")
						batch.Set(metricDoc, map[string]interface{}{
							"confirmed":    1,
							"pending":      0,
							"totalFeeGRX":  0,
							"totalPayment": 0,
						})

					} else {
						log.Println("existed referral")
						batch.Update(metricDoc.Ref, []firestore.Update{
							{
								Path:  "confirmed",
								Value: firestore.Increment(1),
							},
							{
								Path:  "pending",
								Value: firestore.Increment(-1),
							},
						})
					}

					// Referer user
					title, content, contents := GenInvitationConfirmedSender(input.Name, input.LName, invitedDate)
					mailGeneral, err := h.apiContext.Cache.GetNotice(refererUid, "MailGeneral")
					if err != nil {
						log.Println("Can not get MailWallet setting from cache:", err)
					} else {
						// check setting and send mail
						if mailGeneral == "1" {
							err = mail.SendNoticeMail(referer.Data()["Email"].(string), referer.Data()["Name"].(string), title, contents)
							if err != nil {
								log.Println("[ERROR]- Invite - can not send mail invite:", err)
								GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
								return
							}
						}
					}

					// App notice
					notice := map[string]interface{}{
						"title":  title,
						"body":   content,
						"isRead": false,
						"time":   time.Now().Unix(),
					}
					docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(refererUid).NewDoc()
					batch.Set(docRef, notice)

					userMeta := h.apiContext.Store.Doc("users_meta/" + refererUid)
					batch.Update(userMeta, []firestore.Update{
						{Path: "UrGeneral", Value: firestore.Increment(1)},
					})

					// Referral user
					title, content, contents = GenInvitationConfirmed(referer.Data()["Name"].(string), referer.Data()["LName"].(string), invitedDate)
					err = mail.SendNoticeMail(input.Email, input.Name, title, contents)
					if err != nil {
						log.Println("[ERROR]- Invite - can not send mail invite:", err)
						GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
						return
					}

					// App notice
					notice = map[string]interface{}{
						"title":  title,
						"body":   content,
						"isRead": false,
						"time":   time.Now().Unix(),
					}
					docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
					batch.Set(docRef, notice)

					referralUGeneral = 1
				}
			}
		}

		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, input.Email)
		if encodeStr == "" {
			_, err = h.apiContext.Store.Doc("users/" + uid).Delete(ctx)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}

		err = mail.SendMail(input.Email, input.Name, ConfirmRegistrationSub, VerifyEmail, encodeStr, h.apiContext.Config.Host, nil)
		if err != nil {
			log.Println("[ERROR] Can not send registration confirm email:", input.Email, err)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}

		batch.Set(userDoc, input)
		userMeta := map[string]interface{}{"UrWallet": 0, "UrGRY1": 0, "UrGRY2": 0, "UrGRY3": 0, "UrGRZ": 0, "UrGeneral": referralUGeneral,
			"Email": input.Email, "Name": input.Name, "LName": input.LName, "UserId": uid, "CreatedAt": input.CreatedAt,
			"OpenOrders": 0, "OpenOrdersGRX": 0, "OpenOrdersXLM": 0, "GRX": float64(0), "XLM": float64(0)}

		userMetaDoc := h.apiContext.Store.Doc("users_meta/" + uid)
		batch.Set(userMetaDoc, userMeta)

		_, err = batch.Commit(ctx)
		if err != nil {
			log.Println("[ERROR] - Register - Commit:", input.Email, err)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}
		if refererUid != "" {
			h.apiContext.Cache.SetRefererUid(uid, refererUid)
		}

		// Save registration infor to SendGrid db
		//12592774 : confirmed invite referral
		//12592770: pending invite
		go func() {
			if referralSendGridId != "" {
				mail.AddRecipienttoList(referralSendGridId, 12592774)
				mail.RemoveRecipientFromList(referralSendGridId, 12592770)
			} else {
				receiptId, err := mail.SaveRegistrationInfo(input.Name, input.LName, input.Email, input.CreatedAt, 0)
				if receiptId != "" && err != nil {
					h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{
						"SendGridId": receiptId,
					}, firestore.MergeAll)
				}
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}

//cities := client.Collection("cities")
// Get the first 25 cities, ordered by population.
// firstPage := cities.OrderBy("population", firestore.Asc).Limit(25).Documents(ctx)
// docs, err := firstPage.GetAll()
// if err != nil {
//         return err
// }
//
// // Get the last document.
// lastDoc := docs[len(docs)-1]
//
// // Construct a new query to get the next 25 cities.
// secondPage := cities.OrderBy("population", firestore.Asc).
//         StartAfter(lastDoc.Data()["population"]).
//         Limit(25)

func (h UserHandler) GetNotices() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Limit      int    `json:"limit"`
			Cursor     int64  `json:"cursor"`
			NoticeType string `type:"type"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}
		if input.NoticeType != "all" && input.NoticeType != "wallet" && input.NoticeType != "algo" && input.NoticeType != "general" {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "invalid notice type")
			return
		}
		if input.Limit >= 200 {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "invalid limit param")
			return
		}
		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		switch input.NoticeType {
		case "all":
			wallets := h.getDetailNotice("wallet", input.Limit, input.Cursor, uid)
			algos := h.getDetailNotice("algo", input.Limit, input.Cursor, uid)
			generals := h.getDetailNotice("general", input.Limit, input.Cursor, uid)
			c.JSON(http.StatusOK, gin.H{
				"errCode": SUCCESS, "wallets": wallets, "algos": algos, "generals": generals,
			})
		default:
			notices := h.getDetailNotice(input.NoticeType, input.Limit, input.Cursor, uid)
			c.JSON(http.StatusOK, gin.H{
				"errCode": SUCCESS, "notices": notices,
			})
		}

	}
}

func (h UserHandler) getDetailNotice(noticeType string, limit int, cursor int64, uid string) []map[string]interface{} {
	notices := make([]map[string]interface{}, 0)
	var it *firestore.DocumentIterator
	if cursor > 0 {
		it = h.apiContext.Store.Collection("notices/"+noticeType+"/"+uid).Limit(limit).StartAfter(cursor).OrderBy("time", firestore.Desc).Documents(context.Background())
	} else {
		it = h.apiContext.Store.Collection("notices/"+noticeType+"/"+uid).Limit(limit).OrderBy("time", firestore.Desc).Documents(context.Background())
	}
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("err reading db: ", err)
		}
		var data map[string]interface{}
		if noticeType == "wallet" || noticeType == "algo" {
			data = map[string]interface{}{
				"id":     doc.Ref.ID,
				"type":   doc.Data()["type"].(string),
				"title":  doc.Data()["title"].(string),
				"body":   doc.Data()["body"].(string),
				"time":   doc.Data()["time"].(int64),
				"txId":   doc.Data()["txId"].(string),
				"isRead": doc.Data()["isRead"].(bool),
			}
		} else {
			data = map[string]interface{}{
				"id":     doc.Ref.ID,
				"type":   doc.Data()["type"].(string),
				"title":  doc.Data()["title"].(string),
				"body":   doc.Data()["body"].(string),
				"time":   doc.Data()["time"].(int64),
				"isRead": doc.Data()["isRead"].(bool),
			}
		}
		notices = append(notices, data)
	}
	return notices
}

func (h UserHandler) ValidatePhone() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Phone string `json:"phone"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, err.Error())
			return
		}
		uid := c.GetString("Uid")
		// userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		// if userInfo == nil {
		// 	GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
		// 	return
		// }

		// if phone, ok := userInfo["Phone"]; ok {
		// 	if phone == input.Phone {
		// 		GinRespond(c, http.StatusOK, PHONE_EXIST, "Phone number already registered.")
		// 		return
		// 	}
		// }

		// check phone number already registered or not
		userInfo, _ := GetUserByField(h.apiContext.Store, "Phone", input.Phone)
		if userInfo != nil {
			GinRespond(c, http.StatusOK, PHONE_EXIST, "Phone number already registered.")
			return
		}

		res, err := http.Get(h.apiContext.Config.Numberify + input.Phone)
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		var values map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&values)
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		if valid, ok := values["valid"]; ok {
			if valid.(bool) {
				_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
					"Phone": input.Phone,
				}, firestore.MergeAll)
				if err != nil {
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
				} else {
					GinRespond(c, http.StatusOK, SUCCESS, "")
				}
				return
			}
		}
		GinRespond(c, http.StatusOK, INVALID_CODE, "")
	}
}

func (h UserHandler) UpdateAllAsRead() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		noticeType := c.Param("noticeType")
		log.Println("noticeType:", noticeType)
		docPath := ""

		urMap := make(map[string]interface{})
		switch noticeType {
		case "algo", "gry1", "gry2", "gry3", "grz":
			docPath = "notices/algo/" + uid
			urMap = map[string]interface{}{"UrGRZ": 0, "UrGRY1": 0, "UrGRY2": 0, "UrGRY3": 0}
		case "wallet":
			docPath = "notices/wallet/" + uid
			urMap = map[string]interface{}{"UrWallet": 0}
		case "general":
			docPath = "notices/general/" + uid
			urMap = map[string]interface{}{"UrGeneral": 0}

		default:
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Invalid notice type")
			return
		}
		ctx := context.Background()
		iter := h.apiContext.Store.Collection(docPath).Where("isRead", "==", false).Documents(ctx)
		batch := h.apiContext.Store.Batch()
		cnt := 0
		total := 0
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if noticeType != "grz" && noticeType != "gry1" && noticeType != "gry2" && noticeType != "gry3" {
				batch.Set(doc.Ref, map[string]interface{}{"isRead": true}, firestore.MergeAll)
				cnt++
				total++
			} else {
				docType := doc.Data()["type"].(string)
				if noticeType == "grz" && docType == "GRZ" {
					batch.Set(doc.Ref, map[string]interface{}{"isRead": true}, firestore.MergeAll)
					cnt++
					total++
				}
				if noticeType == "gry1" && docType == "GRY 1" {
					batch.Set(doc.Ref, map[string]interface{}{"isRead": true}, firestore.MergeAll)
					cnt++
					total++
				}
				if noticeType == "gry2" && docType == "GRY 2" {
					batch.Set(doc.Ref, map[string]interface{}{"isRead": true}, firestore.MergeAll)
					cnt++
					total++
				}
				if noticeType == "gry3" && docType == "GRY 3" {
					batch.Set(doc.Ref, map[string]interface{}{"isRead": true}, firestore.MergeAll)
					cnt++
					total++
				}
			}
			if cnt >= 300 {
				_, err = batch.Commit(ctx)
				if err != nil {
					log.Println("[ERROR] batch commit:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not update read all")
					return
				}
				batch = h.apiContext.Store.Batch()
				cnt = 0
			}
		}

		if total > 0 {
			userMeta := h.apiContext.Store.Doc("users_meta/" + uid)
			if noticeType == "grz" || noticeType == "gry1" || noticeType == "gry2" || noticeType == "gry3" {
				switch noticeType {
				case "grz":
					urMap = map[string]interface{}{"UrGRZ": 0}
				case "gry1":
					urMap = map[string]interface{}{"UrGRY1": 0}
				case "gry2":
					urMap = map[string]interface{}{"UrGRY2": 0}
				case "gry3":
					urMap = map[string]interface{}{"UrGRY3": 0}
				}
			}
			batch.Set(userMeta, urMap, firestore.MergeAll)
			_, err = batch.Commit(ctx)
			if err != nil {
				log.Println("[ERROR] batch commit:", err)
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not update read all")
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})
	}
}

func (h UserHandler) UpdateReadNotices() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		var input struct {
			WalletIds  []string `json:"walletIds"`
			AlgoIds    []string `json:"algoIds"`
			GeneralIds []string `json:"generalIds"`
			UrGRZ      int64    `json:"urgrz,omitempty"`
			UrGRY1     int64    `json:"urgry1,omitempty"`
			UrGRY2     int64    `json:"urgry2,omitempty"`
			UrGRY3     int64    `json:"urgry3,omitempty"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, err.Error())
			return
		}
		ctx := context.Background()
		if len(input.WalletIds) > 0 {
			for _, id := range input.WalletIds {
				h.apiContext.Store.Doc("notices/wallet/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}
			cnt := 0 - len(input.WalletIds)
			_, err := h.apiContext.Store.Doc("users_meta/"+uid).Update(ctx, []firestore.Update{
				{Path: "UrWallet", Value: firestore.Increment(cnt)},
			})
			if err != nil {
				log.Println("SaveNotice update error: ", err)
				return
			}
		}
		if len(input.AlgoIds) > 0 {
			for _, id := range input.AlgoIds {
				h.apiContext.Store.Doc("notices/algo/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}

			if err != nil {
				log.Println("SaveNotice update error: ", err)
				return
			}
			log.Println("Input:", input)
			_, err := h.apiContext.Store.Doc("users_meta/"+uid).Set(ctx, map[string]interface{}{
				"UrGRZ": input.UrGRZ, "UrGRY1": input.UrGRY1, "UrGRY2": input.UrGRY2, "UrGRY3": input.UrGRY3,
			}, firestore.MergeAll)
			if err != nil {
				log.Println("SaveNotice un read notice error: ", err)
				return
			}
		}
		if len(input.GeneralIds) > 0 {
			for _, id := range input.GeneralIds {
				h.apiContext.Store.Doc("notices/general/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}
			cnt := 0 - len(input.GeneralIds)
			_, err := h.apiContext.Store.Doc("users_meta/"+uid).Update(ctx, []firestore.Update{
				{Path: "UrGeneral", Value: firestore.Increment(cnt)},
			})
			if err != nil {
				log.Println("SaveNotice update error: ", err)
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})
	}
}

// Login handles login router.
// Function validates parameters and call Login from UserStore.
func (h UserHandler) SaveSubcriber() gin.HandlerFunc {
	return func(c *gin.Context) {
		// PushSubscription {endpoint: "https://fcm.googleapis.com/fcm/send/eup4tFLqovM:AP…btRIe1kDd5uI9Fy0mH3cdIQQpR99ARkj4pVIz4Q9vtsHhQlOO",
		// expirationTime: null, options: PushSubscriptionOptions}
		// var subs struct {
		// 	Endpoint       string `json:"endpoint"`
		// 	ExpirationTime string `json:"expirationTime,omitempty"`
		// }
		subs := make(map[string]interface{}, 0)

		rawData, err := c.GetRawData()
		if err != nil {
			log.Println("GetRawData error: ", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Error can not parse input data")
			return
		}
		err = json.Unmarshal(rawData, &subs)
		if err != nil {
			log.Println("BindJSON error: ", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Error can not parse input data")
			return
		}
		// Validate user data

		log.Println("rawData: ", string(rawData))
		uid := c.GetString("Uid")

		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		// Save to db
		// data := map[string]bool{"IpConfirm": true, "MulSignature": false, "AppGeneral": true, "AppWallet": true,
		// 	"AppAlgo": true, "MailGeneral": true, "MailWallet": true, "MailAlgo": true}

		// err = h.apiContext.Store.RunTransaction(context.Background(), func(ctx context.Context, tx *firestore.Transaction) error {
		// 	docRef := h.apiContext.Store.Collection("subsriptions").Doc(uid)
		// 	err := tx.Set(docRef, subs)
		// 	if err != nil {
		// 		log.Println("Add subs error: ", err)
		// 		return err
		// 	}
		//
		// 	docRef = h.apiContext.Store.Doc("users/" + uid)
		// 	return tx.Set(docRef, map[string]interface{}{"Setting": data}, firestore.MergeAll)
		// })

		// _, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(),
		// 	map[string]interface{}{"Setting": data, "Subs": subs}, firestore.MergeAll)
		//
		// if err != nil {
		// 	log.Println("Add subs error 1: ", err)
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
		// 	return
		// }
		//_, err = h.apiContext.Store.Collection("subsriptions").Doc(uid).Set(context.Background(), subs)
		// Set setting cache

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(),
			map[string]interface{}{"Subs": subs}, firestore.MergeAll)

		if err != nil {
			log.Println("Add subs error 1: ", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		// Add subs to cache
		go func() {
			// _, err = h.apiContext.Cache.SetUserSubs(uid, string(rawData))
			// if err != nil {
			// 	// GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			// 	// return
			// 	log.Println("Can not save the subs to cache: ", err)
			// }

			// for k, v := range data {
			// 	_, err = h.apiContext.Cache.SetNotice(uid, k, v)
			// 	if err != nil {
			// 		log.Printf(uid+": SetNotice cache error %v\n", err)
			// 	}
			// }
		}()

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})
	}
}

// Login handles login router.
// Function validates parameters and call Login from UserStore.
func (h UserHandler) ChangeEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var changeMail struct {
			Email    string `json:"email"`
			NewEmail string `json:"newemail"`
			Password string `json:"password"`
		}

		err := c.BindJSON(&changeMail)
		if err != nil {
			log.Println("BindJSON error: ", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Error can not parse input data")
			return
		}
		// Validate user data
		if govalidator.IsNull(changeMail.Password) || !govalidator.IsEmail(changeMail.Email) || !govalidator.IsEmail(changeMail.NewEmail) {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Input data is invalid")
			return
		}
		// Check exist email
		userInfo, _ := GetUserByField(h.apiContext.Store, "Email", changeMail.NewEmail)
		if userInfo != nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Email existed")
			return
		}

		uid := c.GetString("Uid")

		userInfo, _ = GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		if userInfo["Email"].(string) != changeMail.Email {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email does not exist")
			return
		}

		// Check password
		ret, err := utils.VerifyPassphrase(changeMail.Password, userInfo["HashPassword"].(string))
		if !ret || err != nil {
			log.Println("VerifyPassword: pass is not matched")
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, changeMail.Email+"?"+changeMail.NewEmail)
		if encodeStr == "" {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not change email right now. Please try again later.")
			return
		}
		err = mail.SendMail(changeMail.Email, userInfo["Name"].(string), ChangeEmailSub, ChangeEmail,
			encodeStr, h.apiContext.Config.Host, map[string]string{"newemail": changeMail.NewEmail})
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not login right now. Please try again later.")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})
	}
}

// Register handles register router.
// Function validates parameters and call Register from UserStore.
func (h UserHandler) ValidateCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		oobCode := c.Query("oobCode")
		mode := c.Query("mode")
		itemData := utils.DecryptItem(h.apiContext.Jwt.PrivateKey, oobCode)
		if itemData == "" {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Invalid code")
			return
		}
		log.Println("itemData:", itemData)
		switch mode {
		case "verifyEmail":
			userInfo, uid := GetUserByField(h.apiContext.Store, "Email", itemData)
			if userInfo == nil {
				GinRespond(c, http.StatusBadRequest, INVALID_CODE, "Invalid code")
				return
			}
			log.Println("uid:", uid)
			_, err := h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"IsVerified": true}, firestore.MergeAll)
			if err != nil {
				GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not update verifed email status")
				return
			}
		case "confirmIp":
			items := strings.Split(itemData, "?")
			if len(items) == 2 {
				//0- is ip
				//1- is uid
				userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", items[1])
				if userInfo == nil {
					GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not verify ip changes")
					return
				}

				secondIp, ok := userInfo["SecondIp"]
				if !ok {
					log.Println("Set second ip to: ", items[0])
					_, err := h.apiContext.Store.Doc("users/"+items[1]).Set(ctx, map[string]interface{}{"SecondIp": items[0]}, firestore.MergeAll)
					if err != nil {
						GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not verify ip changes")
						return
					}
				} else { // secondip already set, set to Ip field
					log.Println("Set second ip to: ", items[0])
					log.Println("Set ip to: ", secondIp.(string))
					_, err := h.apiContext.Store.Doc("users/"+items[1]).Set(ctx, map[string]interface{}{"Ip": secondIp.(string),
						"SecondIp": items[0]}, firestore.MergeAll)
					if err != nil {
						GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not verify ip changes")
						return
					}
				}
			}
		case "resetPassword":
		case "changeEmail":
			items := strings.Split(itemData, "?")
			if len(items) == 2 {
				user, uid := GetUserByField(h.apiContext.Store, "Email", items[0])
				_, err := h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"IsVerified": false, "Email": items[1], "Federation": items[1] + "*grayll.io"}, firestore.MergeAll)
				if err != nil {
					h.apiContext.Store, _ = ReconnectFireStore("grayll-app-f3f3f3", 60)
					_, err = h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"IsVerified": false, "Email": items[1], "Federation": items[1] + "*grayll.io"}, firestore.MergeAll)
					if err != nil {
						log.Println("ERROR - changeemail - unable change email", items, err)
						GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Error set IsVerified false")
						return
					}
				}
				encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, items[1])
				if encodeStr == "" {
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Error gen token string")
					return
				}
				err = mail.SendMail(items[1], user["Name"].(string), ConfirmRegistrationSub, VerifyEmail, encodeStr, h.apiContext.Config.Host, nil)
				if err != nil {
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Error send confirm registration email")
					return
				}
			}

		}

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) ResendEmailConfirm() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}
		log.Println("ResendEmailConfirm: info: ", input)
		// Validate user data
		if govalidator.IsNull(input.Name) || !govalidator.IsEmail(input.Email) {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Input data is invalid")
			return
		}

		userInfo, _ := GetUserByField(h.apiContext.Store, "Email", input.Email)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email account is not registered yet")
			return
		}

		if userInfo["IsVerified"].(bool) {
			GinRespond(c, http.StatusOK, EMAIL_VERIFIED, "Email is verified")
			return
		}
		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, input.Email)
		if encodeStr == "" {
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not send confirm email right now")
			return
		}
		err = mail.SendMail(input.Email, input.Name, ConfirmRegistrationSub, VerifyEmail, encodeStr, h.apiContext.Config.Host, nil)
		if err != nil {
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can resend email register right now")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}
func (h UserHandler) SendEmailResetPwd() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input data")
			return
		}
		log.Println("ResendEmailConfirm: info: ", input)
		// Validate user data
		if !govalidator.IsEmail(input.Email) {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Input data is invalid")
			return
		}

		userInfo, uid := GetUserByField(h.apiContext.Store, "Email", input.Email)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email account is not registered yet")
			return
		}

		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, input.Email)
		if encodeStr == "" {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not encrypt token")
			return
		}

		err = mail.SendMail(input.Email, userInfo["Name"].(string), ResetPasswordSub, ResetPassword, encodeStr, h.apiContext.Config.Host, nil)
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not send reset password email right now")
			return
		}

		// Set token to database
		_, err = h.apiContext.Store.Doc("resetpwd/"+uid).Set(context.Background(), map[string]interface{}{
			"token": encodeStr,
		})

		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not save reset password token")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}
func (h UserHandler) ResetPassword1() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			OobCode     string `json:"oobCode"`
			NewPassword string `json:"newPassword"`
			EnSecretKey string `json:"enSecretKey,omitempty"`
			Salt        string `json:"salt,omitempty"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		log.Println("ResetPassword: info: ", input)
		// Validate user data
		if govalidator.IsNull(input.OobCode) || govalidator.IsNull(input.NewPassword) {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Input data is invalid")
			return
		}

		itemData := utils.DecryptItem(h.apiContext.Jwt.PrivateKey, input.OobCode)
		if itemData == "" {
			GinRespond(c, http.StatusOK, INVALID_CODE, "Invalid code")
			return
		}

		userInfo, uid := GetUserByField(h.apiContext.Store, "Email", itemData)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email account is not registered yet")
			return
		}

		docSs, err := h.apiContext.Store.Doc("resetpwd/" + uid).Get(context.Background())
		if err != nil {
			log.Printf("can not get resetpwd token: %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can get resetpwd")
			return
		}

		if resetpwdToken, ok := docSs.Data()["token"]; ok {
			if resetpwdToken.(string) != input.OobCode {
				log.Println("Reset pwd token invalid")
				GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
				return
			}
		} else {
			log.Println("Reset pwd token invalid")
			GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
		}

		hash, err := utils.DerivePassphrase(input.NewPassword, 32)
		if err != nil {
			log.Printf("ResetPassword error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"HashPassword":  hash,
			"EnSecretKey":   input.EnSecretKey,
			"SecretKeySalt": input.Salt,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("AddUserData:Add error 1%v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}
		// Send mail reset password successfully
		go func() {
			mail.SendMailResetPwdSuccess(userInfo["Email"].(string), userInfo["Name"].(string), "GRAYLL | Reset Password Successfully",
				[]string{
					"Your password has been reset successfully.",
					"If you didn’t request and approve your GRAYLL account password reset, please contact us immediately!",
					"support@grayll.io",
				})
		}()
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			OobCode     string `json:"oobCode"`
			NewPassword string `json:"newPassword"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		//log.Println("ResetPassword: info: ", input)
		// Validate user data
		if govalidator.IsNull(input.OobCode) || govalidator.IsNull(input.NewPassword) {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Input data is invalid")
			return
		}

		itemData := utils.DecryptItem(h.apiContext.Jwt.PrivateKey, input.OobCode)
		if itemData == "" {
			GinRespond(c, http.StatusOK, INVALID_CODE, "Invalid code")
			return
		}

		userInfo, uid := GetUserByField(h.apiContext.Store, "Email", itemData)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email account is not registered yet")
			return
		}

		docSs, err := h.apiContext.Store.Doc("resetpwd/" + uid).Get(context.Background())
		if err != nil {
			log.Printf("can not get resetpwd token: %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can get resetpwd")
			return
		}

		if resetpwdToken, ok := docSs.Data()["token"]; ok {
			if resetpwdToken.(string) != input.OobCode {
				log.Println("Reset pwd token invalid")
				GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
				return
			}
		} else {
			log.Println("Reset pwd token invalid")
			GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
		}

		hash, err := utils.DerivePassphrase(input.NewPassword, 32)
		if err != nil {
			log.Printf("ResetPassword error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"HashPassword":  hash,
			"EnSecretKey":   "",
			"SecretKeySalt": "",
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("AddUserData:Add error 1%v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}
		// Send mail reset password successfully
		go func() {
			mail.SendMailResetPwdSuccess(userInfo["Email"].(string), userInfo["Name"].(string), "GRAYLL | Reset Password Successful",
				[]string{
					"Your password has been reset successfully.",
					"If you didn’t request and approve your GRAYLL account password reset, please contact us immediately!",
					"support@grayll.io",
				})
		}()
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Password    string `json:"password"`
			NewPassword string `json:"newPassword"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		log.Println("ResetPassword: info: ", input)
		// Validate user data
		if govalidator.IsNull(input.Password) || govalidator.IsNull(input.NewPassword) {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Input data is invalid")
			return
		}
		// itemData := utils.DecryptItem(h.apiContext.Jwt.PrivateKey, input.OobCode)
		// if itemData == "" {
		// 	GinRespond(c, http.StatusOK, INVALID_CODE, "Invalid code")
		// 	return
		// }

		// userInfo, uid := GetUserByField(h.apiContext.Store, "Email", itemData)
		// if userInfo == nil {
		// 	GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Email account is not registered yet")
		// 	return
		// }

		// docSs, err := h.apiContext.Store.Doc("resetpwd/" + uid).Get(context.Background())
		// if err != nil {
		// 	log.Printf("can not get resetpwd token: %v\n", err)
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can get resetpwd")
		// 	return
		// }

		// if resetpwdToken, ok := docSs.Data()["token"]; ok {
		// 	if resetpwdToken.(string) != input.OobCode {
		// 		log.Println("Reset pwd token invalid")
		// 		GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
		// 		return
		// 	}
		// } else {
		// 	log.Println("Reset pwd token invalid")
		// 	GinRespond(c, http.StatusOK, INVALID_CODE, "Reset token invalid")
		// }

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		// Check password
		ret, err := utils.VerifyPassphrase(input.Password, userInfo["HashPassword"].(string))
		if !ret || err != nil {
			log.Println(uid + ": Current password is not matched")
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		hash, err := utils.DerivePassphrase(input.NewPassword, 32)
		if err != nil {
			log.Printf("ResetPassword error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"HashPassword":  hash,
			"EnSecretKey":   "",
			"SecretKeySalt": "",
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("AddUserData:Add error 1%v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not change password right now")
			return
		}
		// Send mail reset password successfully
		go func() {
			mail.SendMailResetPwdSuccess(userInfo["Email"].(string), userInfo["Name"].(string), "GRAYLL | Change Password Successful",
				[]string{
					"Your password has been changed successfully.",
					"If you didn’t request and approve your GRAYLL account password reset, please contact us immediately!",
					"support@grayll.io",
				})
		}()
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) ValidateAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Password    string `json:"password"`
			PublicKey   string `json:"publicKey"`
			EnSecretKey string `json:"enSecretKey"`
			Salt        string `json:"salt"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		// Check password
		ret, err := utils.VerifyPassphrase(input.Password, userInfo["HashPassword"].(string))
		if !ret || err != nil {
			log.Println(uid + ": VerifyPassword: pass is not matched")
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		// Fund 1.5 XLM
		if !AddressSanityCheck(input.PublicKey) {
			GinRespond(c, http.StatusOK, INVALID_ADDRESS, "Invalid address")
			return
		}

		if stellar.IsMainNet {
			seq, hash, err := stellar.SendXLMCreateAccount(input.PublicKey, float64(2.1), h.apiContext.Config.XlmLoanerSeed)
			if err != nil {
				log.Println("ERROR - AccountValidate - unable to create account", uid, err)
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
				return
			}

			_, err = h.apiContext.Store.Doc("loans").Set(context.Background(), map[string]interface{}{
				"uid":       uid,
				"txhash":    hash,
				"seq":       seq,
				"source":    h.apiContext.Config.XlmLoanerAddress,
				"address":   input.PublicKey,
				"createdAt": time.Now(),
				"type":      "loan",
			}, firestore.MergeAll)

			if err != nil {
				for i := 0; i < 3; i++ {
					time.Sleep(time.Second * 2)
					_, err = h.apiContext.Store.Doc("loans").Set(context.Background(), map[string]interface{}{
						"txhash":    hash,
						"seq":       seq,
						"source":    h.apiContext.Config.XlmLoanerAddress,
						"address":   input.PublicKey,
						"createdAt": time.Now(),
						"type":      "loan",
					})
					if err == nil {
						break
					}
				}
			}
		} else {
			err = stellar.FundXLMTestNet(input.PublicKey)
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
				return
			}
			// seq, hash, err := assets.SendAsset(h.apiContext.Asset, input.PublicKey, float64(100), h.apiContext.Config.XlmLoanerSeed, "")
			// log.Printf("FundXLMTestNet: seq : %v - hash:  %v - err : %v \n", seq, hash, err)

		}
		activatedAt := time.Now().Unix()
		activatedData := map[string]interface{}{
			"PublicKey":      input.PublicKey,
			"EnSecretKey":    input.EnSecretKey,
			"SecretKeySalt":  input.Salt,
			"LoanPaidStatus": 1,
			"ActivatedAt":    activatedAt,
		}
		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), activatedData, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set activated data error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		// _, err = h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), map[string]interface{}{
		// 	"PublicKey":   input.PublicKey,
		// 	"ActivatedAt": activatedAt,
		// }, firestore.MergeAll)
		// if err != nil {
		// 	log.Printf(uid+": Set activated data user meta error %v\n", err)
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
		// 	return
		// }

		// Set PublicAddress to cache
		go func(uid, publicKey string) {
			_, err = h.apiContext.Cache.SetPublicKey(uid, publicKey)
			//_, err = h.apiContext.Cache.SetPublicKey(input.PublicKey, uid)
			if err != nil {
				log.Printf(uid+": SetPublicKey cache error %v\n", err)
			}

			// Set setting cache
			data := map[string]bool{"IpConfirm": true, "MulSignature": false, "AppGeneral": true, "AppWallet": true,
				"AppAlgo": true, "MailGeneral": true, "MailWallet": true, "MailAlgo": true}
			for k, v := range data {
				_, err = h.apiContext.Cache.SetNotice(uid, k, v)
				if err != nil {
					log.Printf(uid+": SetNotice cache error %v\n", err)
				}
			}

			// set user meta data
			_, err = h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(),
				map[string]interface{}{
					"XLM":         2.1,
					"PublicKey":   publicKey,
					"ActivatedAt": activatedAt}, firestore.MergeAll)
			if err != nil {
				log.Println(uid+": Set users_meta data error %v\n", err)
			}
		}(uid, input.PublicKey)

		go func() {
			createLoanReminder(uid, int64(1), int64(activatedAt))
		}()

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) SaveUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			TotalXLM   float64 `json:"totalXLM"`
			TotalGRX   float64 `json:"totalGRX"`
			OpenOrders int     `json:"openOrders"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		accountData := map[string]interface{}{
			"TotalXLM":   input.TotalXLM,
			"TotalGRX":   input.TotalGRX,
			"OpenOrders": input.OpenOrders,
		}
		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), accountData, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set accountData error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) SaveUserMetaData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			TotalGRZCurrentPositionRoi   float64 `json:"total_grz_current_position_ROI_$"`
			TotalGRZCurrentPositionValue float64 `json:"total_grz_current_position_value_$"`
			TotalGRZOpenPositions        float64 `json:"total_grz_open_positions"`

			TotalGRY1CurrentPositionRoi   float64 `json:"total_gry1_current_position_ROI_$"`
			TotalGRY1CurrentPositionValue float64 `json:"total_gry1_current_position_value_$"`
			TotalGRY1OpenPositions        float64 `json:"total_gry1_open_positions"`

			TotalGRY2CurrentPositionRoi   float64 `json:"total_gry2_current_position_ROI_$"`
			TotalGRY2CurrentPositionValue float64 `json:"total_gry2_current_position_value_$"`
			TotalGRY2OpenPositions        float64 `json:"total_gry2_open_positions"`

			TotalGRY3CurrentPositionRoi   float64 `json:"total_gry3_current_position_ROI_$"`
			TotalGRY3CurrentPositionValue float64 `json:"total_gry3_current_position_value_$"`
			TotalGRY3OpenPositions        float64 `json:"total_gry3_open_positions"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		// userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		// if userInfo == nil {
		// 	GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
		// 	return
		// }

		accountData := map[string]interface{}{
			"total_grz_current_position_ROI_$":   input.TotalGRZCurrentPositionRoi,
			"total_grz_current_position_value_$": input.TotalGRZCurrentPositionValue,
			"total_grz_open_positions":           input.TotalGRZOpenPositions,

			"total_gry1_current_position_ROI_$":   input.TotalGRY1CurrentPositionRoi,
			"total_gry1_current_position_value_$": input.TotalGRY1CurrentPositionValue,
			"total_gry1_open_positions":           input.TotalGRY1OpenPositions,

			"total_gry2_current_position_ROI_$":   input.TotalGRY2CurrentPositionRoi,
			"total_gry2_current_position_value_$": input.TotalGRY2CurrentPositionValue,
			"total_gry2_open_positions":           input.TotalGRY2OpenPositions,

			"total_gry3_current_position_ROI_$":   input.TotalGRY3CurrentPositionRoi,
			"total_gry3_current_position_value_$": input.TotalGRY3CurrentPositionValue,
			"total_gry3_open_positions":           input.TotalGRY3OpenPositions,
		}
		_, err = h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), accountData, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set accountData error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}
		h.apiContext.Cache.Client.MSet(uid+"_total_grz_open_positions", input.TotalGRZOpenPositions, uid+"_total_gry1_open_positions", input.TotalGRY1OpenPositions,
			uid+"_total_gry2_open_positions", input.TotalGRY2OpenPositions, uid+"_total_gry3_open_positions", input.TotalGRY3OpenPositions)

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) SaveEnSecretKeyData() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			EnSecretKey string `json:"enSecretKey"`
			Salt        string `json:"salt"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		accountData := map[string]interface{}{
			"EnSecretKey":   input.EnSecretKey,
			"SecretKeySalt": input.Salt,
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), accountData, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set accountData error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

// AddressSanityCheck checks whether address in wellformed or not
func AddressSanityCheck(address string) bool {
	if stellar.AccountExists(address) {
		return false
	}
	if len(address) != 56 || !strings.HasPrefix(address, "G") {
		return false
	}
	return true
}

func (h UserHandler) UpdateTfa() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tfa models.TfaUpdate
		err := c.BindJSON(&tfa)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		otpc := &dgoogauth.OTPConfig{
			Secret:      tfa.Secret,
			WindowSize:  3,
			HotpCounter: 0,
		}
		val, err := otpc.Authenticate(tfa.OneTimePassword)
		if err != nil || !val {
			fmt.Println("error validate one-tme password:", err)
			GinRespond(c, http.StatusOK, TOKEN_INVALID, "Token is invalid")
			return
		}

		uid := c.GetString(UID)

		tfaData := map[string]interface{}{
			"BackupCode": tfa.BackupCode,
			//"DataURL":    tfa.DataURL,
			"Secret": tfa.Secret,
			"Enable": tfa.Enable,
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"Tfa": tfaData,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("Settfa:Add error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not update right now")
			return
		}
		//c.JSON(http.StatusOK, {})
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) GetFieldInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		input := make(map[string]interface{})
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Account does not exist.")
			return
		}
		res := make(map[string]interface{})
		res["tfa"] = false
		if _, ok := input["tfa"]; ok {
			if _, ok := userInfo["Tfa"]; ok {
				tfaData := userInfo["Tfa"].(map[string]interface{})
				if tfaEnable, ok := tfaData["Enable"]; ok {
					res["tfa"] = tfaEnable.(bool)
					if tfaEnable.(bool) {
						res["secret"] = tfaData["Secret"]
						if _, ok := tfaData["Expire"]; ok {
							res["expire"] = tfaData["Expire"]
						} else {
							res["expire"] = 0
						}
					}
				}
			}
		}
		if _, ok := input["LoanPaidStatus"]; ok {
			if value, ok := userInfo["LoanPaidStatus"]; ok {
				res["isloan"] = value.(int)
			} else {
				res["isloan"] = 0
			}
		}
		if _, ok := input["keyInfo"]; ok {
			res["EnSecretKey"] = userInfo["EnSecretKey"]
			res["SecretKeySalt"] = userInfo["SecretKeySalt"]
			res["LoanPaidStatus"] = userInfo["LoanPaidStatus"]
		}
		res["errCode"] = SUCCESS
		c.JSON(http.StatusOK, res)
	}
}

func (h UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Account does not exist.")
			return
		}
		token := c.GetString("Token")
		hashToken := Hash(token)
		tokenCache := h.apiContext.Cache.Client.Get(hashToken)

		localKey, err := tokenCache.Result()
		if err != nil {
			log.Println("Can not get token in cache:", err)
		}
		res := make(map[string]interface{})
		res["Tfa"] = false
		if _, ok := userInfo["Tfa"]; ok {
			tfaData := userInfo["Tfa"].(map[string]interface{})
			if tfaEnable, ok := tfaData["Enable"]; ok {
				res["Tfa"] = tfaEnable.(bool)
				if tfaEnable.(bool) {
					//res["secret"] = tfaData["Secret"]
					if _, ok := tfaData["Expire"]; ok {
						res["Expire"] = tfaData["Expire"]
					} else {
						res["Expire"] = 0
					}
				}
			}
		}
		// if value, ok := userInfo["LoanPaidStatus"]; ok {
		//
		// } else {
		// 	res["LoanPaidStatus"] = 0
		// }
		res["LoanPaidStatus"] = userInfo["LoanPaidStatus"].(int64)
		setting := userInfo["Setting"].(map[string]interface{})
		res["EnSecretKey"] = userInfo["EnSecretKey"]
		res["SecretKeySalt"] = userInfo["SecretKeySalt"]
		res["Setting"] = setting
		res["Uid"] = uid
		res["errCode"] = SUCCESS
		res["PublicKey"] = userInfo["PublicKey"]
		res["LocalKey"] = localKey

		c.JSON(http.StatusOK, res)
	}
}

func (h UserHandler) SendRevealSecretToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Account does not exist.")
			return
		}

		revealToken := randStr(6, "alphanum")

		// Set token to database
		_, err := h.apiContext.Store.Doc("resetpwd/"+uid).Set(context.Background(), map[string]interface{}{
			"reavealtoken": revealToken,
			"expire":       time.Now().Unix() + int64(10*60),
		}, firestore.MergeAll)

		if err != nil {
			log.Println("error set reveal token:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not save reset reveal token")
			return
		}

		err = mail.SendMail(userInfo["Email"].(string), userInfo["Name"].(string), RevealSecretTokenSub, RevealSecretToken, revealToken, h.apiContext.Config.Host, nil)
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not send reset password email right now")
			return
		}

		// err = mail.SendMail(userInfo["Email"].(string), userInfo["Name"].(string), RevealSecretTokenSub, RevealSecretToken, revealToken, h.apiContext.Config.Host, nil)
		// if err != nil {
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not send reset password email right now")
		// 	return
		// }
		// encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, "grayll@gmail.com")
		// if encodeStr == "" {
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not encrypt token")
		// 	return
		// }
		// err = mail.SendMail("grayll@gmail.com", "huy", ResetPasswordSub, ResetPassword, revealToken, h.apiContext.Config.Host, nil)
		// if err != nil {
		// 	GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not send reset password email right now")
		// 	return
		// }
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) VerifyRevealSecretToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Token string `json:"token"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse json input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Account does not exist.")
			return
		}
		// Get token from database
		doc, err := h.apiContext.Store.Doc("resetpwd/" + uid).Get(context.Background())
		if err != nil {
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not verify reveal secret token")
			return
		}

		if _, ok := doc.Data()["reavealtoken"]; !ok {
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not verify reveal secret token")
				return
			}
		}
		if _, ok := doc.Data()["expire"]; !ok {
			if err != nil {
				GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not verify reveal secret token")
				return
			}
		}

		if input.Token == doc.Data()["reavealtoken"].(string) && time.Now().Unix() < doc.Data()["expire"].(int64) {
			// Set token expired
			_, err = h.apiContext.Store.Doc("resetpwd/"+uid).Set(context.Background(), map[string]interface{}{
				"expire": 0,
			}, firestore.MergeAll)
			GinRespond(c, http.StatusOK, SUCCESS, "")
		} else {
			GinRespond(c, http.StatusOK, INVALID_CODE, "The provided token is invalid or expired")
		}
	}
}

// Expire > 0 is used to update 2FA
// Expire = -1, then will verify password too
// Expire = -2 only verify the 2FA code
func (h UserHandler) VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Token  string `json:"token"`
			Secret string `json:"secret"`
			Expire int64  `json:"expire"`
		}
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			//GinRespond(c, http.StatusOK, INVALID_PARAMS, "Can not parse input data")
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		if govalidator.IsNull(input.Secret) || govalidator.IsNull(input.Token) {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		fmt.Printf("verifytoken- uid %v\n", uid)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			//GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "User does not exist.")
			h.apiContext.Store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
			userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
			if userInfo == nil {
				log.Println("ERROR unable find user with id:", uid)
				output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
				c.JSON(http.StatusOK, output)
				return
			}
		}
		tfa, ok := userInfo["Tfa"].(map[string]interface{})
		if tfa == nil || !ok {
			fmt.Println("verifytoken - Tfa is nil:", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Tfa is not enabled."}
			c.JSON(http.StatusOK, output)
			return
		}

		fmt.Println("verifytoken - input parsed", input)
		otpc := &dgoogauth.OTPConfig{

			Secret:      tfa["Secret"].(string),
			WindowSize:  3,
			HotpCounter: 0,
		}
		val, err := otpc.Authenticate(input.Token)
		if err != nil || !val {
			fmt.Println("error json decode:", err)
			//GinRespond(c, http.StatusOK, TOKEN_INVALID, "Token is invalid")
			output = Output{Valid: false, ErrCode: TOKEN_INVALID, Message: "Token is invalid."}
			c.JSON(http.StatusOK, output)
			return
		}

		if input.Expire > 0 {
			tfa["Expire"] = input.Expire
			_, err := h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
				"Tfa": tfa,
			}, firestore.MergeAll)
			if err != nil {
				//GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not tfa data.")
				output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update expire of tfa data."}
				c.JSON(http.StatusOK, output)
				return
			}
		} else if input.Expire == -1 {
			if ret, err := utils.VerifyPassphrase(input.Secret, userInfo["HashPassword"].(string)); ret {
				_, err = h.apiContext.Store.Doc("users/"+uid).Update(context.Background(), []firestore.Update{
					{Path: "Tfa", Value: firestore.Delete},
				})
				if err != nil {
					//GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not delete tfa data.")
					output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not delete tfa data."}
					c.JSON(http.StatusOK, output)
					return
				}
			} else {
				//GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid password")
				output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD}
				c.JSON(http.StatusOK, output)
				return
			}
		}
		fmt.Println("verifytoken SUCCESS")
		output = Output{Valid: true, ErrCode: SUCCESS}
		c.JSON(http.StatusOK, output)
	}
}
func (h UserHandler) UpdateSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Field  string `json:"field"`
			Status bool   `json:"status"`
		}
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		if govalidator.IsNull(input.Field) {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}
		setting, ok := userInfo["Setting"].(map[string]interface{})
		if setting == nil || !ok {
			fmt.Println("setting is nil:", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not get setting data."}
			c.JSON(http.StatusOK, output)
			return
		}

		//fmt.Printf("input %v\n", input)

		switch input.Field {
		case "IpConfirm", "MulSignature", "AppGeneral", "AppWallet", "AppAlgo", "MailGeneral", "MailWallet", "MailAlgo":
			currentValue, ok := setting[input.Field]
			value := false
			if ok {
				value = currentValue.(bool)
			}
			if value != input.Status {
				setting[input.Field] = input.Status
				_, err := h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
					"Setting": setting,
				}, firestore.MergeAll)
				if err != nil {
					output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update setting."}
					c.JSON(http.StatusOK, output)
					return
				}

				_, err = h.apiContext.Cache.SetNotice(uid, input.Field, input.Status)
				if err != nil {
					log.Println("Can not set notice cache fo redis: ", err)
				}
			}
			//if input.Field != "IpConfirm" && input.Field != "MulSignature" {
			// Set cache h.apiContext.Cache.SetNotice(uid, field, val.(bool))
			log.Println("Set value setting to redis:", value)

		default:
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update setting."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}

func (h UserHandler) UpdateKycDoc() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input map[string]interface{}
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			log.Println("Parse kyc err", err)
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}
		uid := c.GetString(UID)
		// userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		// if userInfo == nil {
		// 	output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
		// 	c.JSON(http.StatusOK, output)
		// 	return
		// }
		ctx := context.Background()
		userSnap, err := h.apiContext.Store.Doc("users_meta/" + uid).Get(ctx)
		if err != nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}
		userInfo := userSnap.Data()
		if val, ok := userInfo["Status"]; ok {
			if val.(string) == "Approved" {
				output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "The KYC application has been approved."}
				c.JSON(http.StatusOK, output)
			}
		}
		fieldName := ""
		for k, _ := range input {
			fieldName = k
			newKey := k + "Time"
			input[newKey] = time.Now().Unix()
			break
		}
		batch := h.apiContext.Store.Batch()

		docRef := h.apiContext.Store.Doc("users_meta/" + uid)
		batch.Set(docRef, map[string]interface{}{"KycDocs": input}, firestore.MergeAll)
		kycDocs := userInfo["KycDocs"]
		for k, v := range input {
			(kycDocs.(map[string]interface{}))[k] = v
		}

		log.Println("userInfo", userInfo)
		ret, msg := VerifyKycStatus(userInfo)
		appType := userInfo["Kyc"].(map[string]interface{})["AppType"].(string)
		//log.Println("verify result ret,msg: ", ret, msg)
		// All required documents subitted
		if ret == 0 && len(msg) == 0 {
			// docRef = h.apiContext.Store.Doc("users/" + uid)
			// batch.Set(docRef, map[string]interface{}{"Status": "Submitted"}, firestore.MergeAll)

			docRef = h.apiContext.Store.Doc("users_meta/" + uid)
			batch.Set(docRef, map[string]interface{}{"Status": "Submitted"}, firestore.MergeAll)

			title, content, contents := GenSubmitCompleted(appType)
			mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)

			notice := map[string]interface{}{
				"title":  title,
				"body":   content,
				"isRead": false,
				"time":   time.Now().Unix(),
			}
			docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
			batch.Set(docRef, notice)

			pk, xlm, grx, algoValue := h.GetUserValue(uid)
			title, content, contents = GenSubmitCompletedGrayll(userInfo["Name"].(string), userInfo["LName"].(string), uid, pk, appType, xlm, grx, algoValue)
			mail.SendNoticeMail(SUPER_ADMIN_EMAIL, SUPER_ADMIN_NAME, title, contents)

		} else if len(msg) != 0 {
			docName := GetFriendlyName(fieldName)
			title, content, contents := GenDocSubmitOk(appType, docName, msg)
			mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)

			notice := map[string]interface{}{
				"title":  title,
				"body":   content,
				"isRead": false,
				"time":   time.Now().Unix(),
			}
			docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
			batch.Set(docRef, notice)

		}

		_, err = batch.Commit(context.Background())
		if err != nil {
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update kyc doc."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}

func (h UserHandler) UpdateKyc() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input KYC
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			log.Println("Parse kyc err", err)
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		if govalidator.IsNull(input.Name) || govalidator.IsNull(input.LName) || govalidator.IsNull(input.GovId) ||
			govalidator.IsNull(input.Nationality) || govalidator.IsNull(input.DoB) {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		//userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		ctx := context.Background()
		userSnap, err := h.apiContext.Store.Doc("users_meta/" + uid).Get(ctx)
		if err != nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}
		userInfo := userSnap.Data()
		if val, ok := userInfo["Status"]; ok {
			if val.(string) == "Approved" {
				output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "The KYC documents has been approved."}
				c.JSON(http.StatusOK, output)
			}
		}
		_, err = h.apiContext.Store.Doc("users_meta/"+uid).Set(context.Background(), map[string]interface{}{"Kyc": input}, firestore.MergeAll)
		if err != nil {
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update profile."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}
func (h UserHandler) UpdateKycCom() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input KYCCom
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		log.Println("Parse kyc err", err)
		if err != nil {

			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		if govalidator.IsNull(input.Name) || (govalidator.IsNull(input.Address1) && govalidator.IsNull(input.Address2)) || govalidator.IsNull(input.Registration) ||
			govalidator.IsNull(input.City) || govalidator.IsNull(input.Country) {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Input data is invalid"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}

		if val, ok := userInfo["Status"]; ok {
			if val.(string) == "Approved" {
				output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "The KYC documents has been approved."}
				c.JSON(http.StatusOK, output)
			}
		}

		status := map[string]interface{}{"AppType": "Company"}
		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"KycCom": input, "Kyc": status}, firestore.MergeAll)
		if err != nil {
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update profile."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}
func (h UserHandler) FirebaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
			Token   string `json:"token"`
		}
		var output Output
		uid := c.GetString(UID)
		opt := option.WithCredentialsFile("grayll-kyc-firebase-adminsdk-g2ga0-8b0dc5462e.json")
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Println("Can not create new firebase app: %v\n", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not create new firebase app"}
			c.JSON(http.StatusOK, output)
			return
		}
		client, err := app.Auth(context.Background())
		if err != nil {
			log.Println("Can not create new firebase app: %v\n", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not authenticate firebase app"}
			c.JSON(http.StatusOK, output)
			return
		}

		token, err := client.CustomToken(context.Background(), uid)
		if err != nil {
			log.Println("Can not create new firebase app: %v\n", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not create firebase authentication token"}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true, Token: token}
		c.JSON(http.StatusOK, output)
	}
}

func (h UserHandler) UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name  string `json:"name,omitempty"`
			LName string `json:"lname,omitempty"`
			Phone string `json:"phone,omitempty"`
		}
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		log.Println(input)

		if govalidator.IsNull(input.Name) && govalidator.IsNull(input.LName) && govalidator.IsNull(input.Phone) {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"Name": input.Name, "LName": input.LName, "Phone": input.Phone,
		}, firestore.MergeAll)
		if err != nil {
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update profile."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}
func (h UserHandler) EditFederation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Federation string `json:"federation"`
		}
		type Output struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
			ErrCode string `json:"errCode"`
		}
		var output Output
		var err = c.BindJSON(&input)
		if err != nil {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Can not parse input data"}
			c.JSON(http.StatusOK, output)
			return
		}

		if len(input.Federation) < 4 || len(input.Federation) > 36 {
			output = Output{Valid: false, ErrCode: INVALID_PARAMS, Message: "Federation must be between 4 and 20"}
			c.JSON(http.StatusOK, output)
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}

		if !strings.HasSuffix(input.Federation, "*grayll.io") {
			input.Federation = input.Federation + "*grayll.io"
		}

		other, _ := GetUserByField(h.apiContext.Store, "Federation", input.Federation)
		if other != nil {
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "Federation address in used."}
			c.JSON(http.StatusOK, output)
			return
		}

		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{
			"Federation": input.Federation,
		}, firestore.MergeAll)
		if err != nil {
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Can not update profile."}
			c.JSON(http.StatusOK, output)
			return
		}

		output = Output{Valid: true}
		c.JSON(http.StatusOK, output)
	}
}
func (h UserHandler) SetupTfa() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Account string `json:"account"`
		}
		type Output struct {
			Secret string `json:"secret"`
			URL    string `json:"dataURL"`
		}

		err := c.BindJSON(&input)
		if err != nil || govalidator.IsNull(input.Account) {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Account is not provided")
			return
		}

		// Generate random secret instead of using the test value above.
		secret := make([]byte, 10)
		_, err = rand.Read(secret)
		if err != nil {
			panic(err)
		}

		secretBase32 := base32.StdEncoding.EncodeToString(secret)
		issuer := "GrayLL"

		URL, err := url.Parse("otpauth://totp")
		if err != nil {
			panic(err)
		}

		URL.Path += "/" + url.PathEscape(issuer) + ":" + url.PathEscape(input.Account)

		params := url.Values{}
		params.Add("secret", secretBase32)
		params.Add("issuer", issuer)

		URL.RawQuery = params.Encode()
		fmt.Printf("URL is %s\n", URL.String())
		out := Output{secretBase32, URL.String()}

		c.JSON(http.StatusOK, out)
	}
}

//https://YOUR_DOMAIN/.well-known/stellar.toml
//FEDERATION_SERVER="https://api.yourdomain.com/federation"
//https://YOUR_FEDERATION_SERVER/federation?q=jed*stellar.org&type=name
//https://YOUR_FEDERATION_SERVER/federation?q=GD6WU64OEP5C4LRBH6NK3MHYIA2ADN6K6II6EXPNVUR3ERBXT4AN4ACD&type=id
//{
// "stellar_address": <username*domain.tld>,
// "account_id": <account_id>,
// "memo_type": <"text", "id" , or "hash"> *optional*
// "memo": <memo to attach to any payment. if "hash" type then will be base64 encoded> *optional*
// }
func (h UserHandler) Federation() gin.HandlerFunc {
	return func(c *gin.Context) {

		type Output struct {
			StellarAddress string `json:"stellar_address"`
			AccountId      string `json:"account_id"`
			MemoType       string `json:"memo_type,omitempty"`
			Memo           string `json:"memo,omitempty"`
		}
		var output Output
		typeQ := c.Query("type")
		q := c.Query("q")
		log.Println("query:", typeQ, q)
		if govalidator.IsNull(typeQ) || govalidator.IsNull(q) {
			c.JSON(http.StatusBadRequest, output)
			return
		}

		switch typeQ {
		case "name":
			userInfo, _ := GetUserByField(h.apiContext.Store, "Federation", q)
			if userInfo == nil {
				c.JSON(http.StatusNotFound, output)
				return
			}

			output.StellarAddress = userInfo["Federation"].(string)
			output.AccountId = userInfo["PublicKey"].(string)
			c.JSON(http.StatusOK, output)
		case "id":
			userInfo, _ := GetUserByField(h.apiContext.Store, "PublicKey", q)
			if userInfo == nil {
				log.Println("user is nil")
				c.JSON(http.StatusNotFound, output)
				return
			}

			output.StellarAddress = userInfo["Federation"].(string)
			output.AccountId = userInfo["PublicKey"].(string)
			c.JSON(http.StatusOK, output)

		default:
			c.JSON(http.StatusBadRequest, output)
		}
	}
}

func (h UserHandler) Invite() gin.HandlerFunc {
	return func(c *gin.Context) {
		input := Contact{}
		uid := c.GetString(UID)
		err := c.BindJSON(&input)
		if err != nil {
			log.Println("[ERROR]- Invite - can not bind json:", uid, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "")
			return
		}

		// err = VerifyEmailNeverBounce(h.apiContext.Config.NeverBounceApiKey, input.Email)
		// if err != nil {
		// 	log.Println("[ERROR]- Invite - email invalid:", err)
		// 	GinRespond(c, http.StatusOK, INVALID_ADDRESS, "invalid email address")
		// 	return
		// }

		//check whether email already registered with Grayll
		_, referralUid := GetUserByField(h.apiContext.Store, "Email", input.Email)
		if referralUid != "" {
			log.Println("[ERROR]- Invite - email already used:", uid, input.Email, err)
			GinRespond(c, http.StatusOK, EMAIL_IN_USED, "email in used")
			return
		}

		batch := h.apiContext.Store.Batch()
		ctx := context.Background()

		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			log.Println("[ERROR]- Invite - can not save invite contact:", uid, err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}

		// save to pending invite sendgrid list
		currentTime := time.Now().Unix()
		sendgridId, err := mail.SaveRegistrationInfo(input.Name, input.LName, input.Email, currentTime, 12592770)

		log.Println("sendgridId:", uid, sendgridId)

		doc := h.apiContext.Store.Collection("referrals/" + uid + "/invite").NewDoc()
		invited := map[string]interface{}{
			"id":             doc.ID,
			"name":           input.Name,
			"lname":          input.LName,
			"email":          input.Email,
			"businessName":   input.BusinessName,
			"phone":          input.Phone,
			"remindTime":     0,
			"status":         "pending",
			"lastSentRemind": currentTime,
			"sentRemind":     currentTime,
			"sendGridId":     sendgridId,
		}

		title, url, contents := GenInvite(uid, userInfo["Name"].(string), userInfo["LName"].(string), doc.ID)
		err = mail.SendMailRegistrationInvite(input.Email, input.Name, title, url, contents)
		if err != nil {
			log.Println("[ERROR]- Invite - can not send mail invitee:", uid, err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
			return
		}
		batch.Set(doc, invited)

		title, content, contents := GenInviteSender(input.Name, input.LName)
		mailGeneral, err := h.apiContext.Cache.GetNotice(uid, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail inviter:", uid, err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice := map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
		batch.Set(docRef, notice)

		userMeta := h.apiContext.Store.Doc("users_meta/" + uid)
		batch.Update(userMeta, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})

		metricDoc, err := h.apiContext.Store.Doc("referrals/" + uid + "/metrics/referral").Get(ctx)

		if err != nil && grpc.Code(err) == codes.NotFound {
			// already exist
			log.Println("!existed referral", uid)
			metricDoc := h.apiContext.Store.Doc("referrals/" + uid + "/metrics/referral")
			batch.Set(metricDoc, map[string]interface{}{
				"confirmed":    0,
				"pending":      1,
				"totalFeeGRX":  0,
				"totalPayment": 0,
			})

		} else {
			log.Println("existed referral", uid)
			batch.Update(metricDoc.Ref, []firestore.Update{
				{Path: "pending", Value: firestore.Increment(1)},
			})
		}

		_, err = batch.Commit(ctx)

		if err != nil {
			log.Println("[ERROR]- Invite - can not commit batch:", uid, input.Email, err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}
		log.Println("Invite - successfully", uid, err)
		GinRespond(c, http.StatusOK, SUCCESS, "")

	}
}
func (h UserHandler) ReportClosing() gin.HandlerFunc {
	return func(c *gin.Context) {

		type Input struct {
			GrayllTxId       string  `json:"grayllTxId"`
			Algorithm        string  `json:"algorithm"`
			GrxUsd           float64 `json:"grxUsd"`
			PositionValue    float64 `json:"positionValue"`
			PositionValueGRX float64 `json:"positionValueGRX"`
			Duration         string  `json:"duration"`
		}
		var reportData struct {
			Positions  []Input `json:"positions"`
			UserId     string  `json:"userId"`
			Name       string  `json:"name"`
			Lname      string  `json:"lname"`
			PublicKey  string  `json:"publicKey"`
			PauseUntil int     `json:"pauseUntil"`
			ClientTime int     `json:"clientTime"`
		}
		uid := c.GetString(UID)
		err := c.BindJSON(&reportData)
		if err != nil {
			log.Println("[ERROR]-ReportClosing - unable bind json ", uid, err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
			return
		}

		content := []string{
			time.Now().Format(`15:04 | 02-01-2006`),

			fmt.Sprintf(`%s %s is attempting to close algo positions outside of the current GRX market volatility parameters.`, reportData.Name, reportData.Lname),

			fmt.Sprintf(`User Account: %s`, reportData.PublicKey),

			fmt.Sprintf(`GRAYLL User ID: %s`, reportData.UserId),
		}
		for _, position := range reportData.Positions {
			positionContent := []string{
				fmt.Sprintf(`==========`),
				fmt.Sprintf(`GRX Rate | $ %7f`, position.GrxUsd),

				fmt.Sprintf(`Algo Position Duration: %s`, position.Duration),

				fmt.Sprintf(`USD Algo Position Value | $ %7f`, position.PositionValue),

				fmt.Sprintf(`GRX Algo Position Value | %7f GRX `, position.PositionValueGRX),

				fmt.Sprintf(`%s | GRAYLL | Transaction ID | %s `, position.Algorithm, position.GrayllTxId),
			}
			content = append(content, positionContent...)
		}

		err = mail.SendNoticeMail("grayll@grayll.io", "GRAYLL", "GRAYLL | GRX Market Volatility | Algo System Intervention", content)
		err = mail.SendNoticeMail("huykbc@gmail.com", "GRAYLL", "GRAYLL | GRX Market Volatility | Algo System Intervention", content)
		if err != nil {
			log.Println("[ERROR]- ReportClosing - can not send mail invite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
			return
		}

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) RemveReferral() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get docid from url
		referralId := c.Param("referralId")
		if referralId == "" {
			log.Println("[ERROR]- RemveReferral - invalid referralId:")
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "referralId not exist")
			return
		}

		ctx := context.Background()
		batch := h.apiContext.Store.Batch()

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)

		doc, err := h.apiContext.Store.Doc("referrals/" + uid + "/referral/" + referralId).Get(ctx)
		if err != nil {
			log.Println("[ERROR]- RemveReferral - find the referral with uid:", referralId, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "referralId not found")
			return
		}

		// Referer user
		title, content, contents := GenRemoveRefererralSender(doc.Data()["name"].(string), doc.Data()["lname"].(string))
		mailGeneral, err := h.apiContext.Cache.GetNotice(uid, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail invite:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice := map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
		batch.Set(docRef, notice)
		userMeta := h.apiContext.Store.Doc("users_meta/" + uid)
		batch.Update(userMeta, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})

		// Referral user
		title, content, contents = GenRemoveRefererral(userInfo["Name"].(string), userInfo["LName"].(string))
		mailGeneral, err = h.apiContext.Cache.GetNotice(referralId, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(doc.Data()["email"].(string), doc.Data()["name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail invite:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice = map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(referralId).NewDoc()
		batch.Set(docRef, notice)
		userMeta = h.apiContext.Store.Doc("users_meta/" + referralId)
		batch.Update(userMeta, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})

		refDoc := h.apiContext.Store.Doc("referrals/" + uid + "/referral/" + referralId)
		batch.Delete(refDoc)

		referDoc := h.apiContext.Store.Doc("referrals/" + referralId + "/referer/" + uid)
		batch.Delete(referDoc)

		_, err = batch.Commit(ctx)

		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not batch commit reinvite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}
		h.apiContext.Cache.DelRefererUid(referralId)
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) EditReferral() gin.HandlerFunc {
	return func(c *gin.Context) {
		referral := Contact{}

		err := c.BindJSON(&referral)
		if err != nil {
			log.Println("[ERROR]- EditReferral- Can not parse edit data:", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "")
			return
		}

		ctx := context.Background()

		uid := c.GetString(UID)
		//userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)

		_, err = h.apiContext.Store.Doc("referrals/" + uid + "/referral/" + referral.RefererUid).Get(ctx)
		if err != nil {
			log.Println("[ERROR]- EditReferral - find the referral with uid:", referral.RefererUid, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "referralId not found")
			return
		}

		_, err = h.apiContext.Store.Doc("referrals/"+uid+"/referral/"+referral.RefererUid).Set(ctx, map[string]interface{}{
			"name": referral.Name, "lname": referral.LName, "email": referral.Email, "businessName": referral.BusinessName, "phone": referral.Phone,
		}, firestore.MergeAll)

		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not batch commit reinvite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}

		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) RemveReferer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get docid from url
		refererId := c.Param("refererId")
		if refererId == "" {
			log.Println("[ERROR]- RemveReferral - invalid refererId:")
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "refererId not exist")
			return
		}

		ctx := context.Background()
		batch := h.apiContext.Store.Batch()

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)

		doc, err := h.apiContext.Store.Doc("referrals/" + uid + "/referer/" + refererId).Get(ctx)
		if err != nil {
			log.Println("[ERROR]- RemveReferral - find the refererId with uid:", refererId, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "refererId not found")
			return
		}

		// Referral user
		title, content, contents := GenRemoveReferer(doc.Data()["name"].(string), doc.Data()["lname"].(string))
		mailGeneral, err := h.apiContext.Cache.GetNotice(uid, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail invite:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice := map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
		batch.Set(docRef, notice)
		userMeta := h.apiContext.Store.Doc("users_meta/" + uid)
		batch.Update(userMeta, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})

		// Referer user
		title, content, contents = GenRemoveRefererSender(userInfo["Name"].(string), userInfo["LName"].(string))
		mailGeneral, err = h.apiContext.Cache.GetNotice(refererId, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(doc.Data()["email"].(string), doc.Data()["name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail invite:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice = map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(refererId).NewDoc()
		batch.Set(docRef, notice)
		userMeta = h.apiContext.Store.Doc("users_meta/" + refererId)
		batch.Update(userMeta, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})
		refDoc := h.apiContext.Store.Doc("referrals/" + uid + "/referer/" + refererId)
		batch.Delete(refDoc)

		referDoc := h.apiContext.Store.Doc("referrals/" + refererId + "/referral/" + uid)
		batch.Delete(referDoc)
		_, err = batch.Commit(ctx)

		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not batch commit reinvite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}
		h.apiContext.Cache.DelRefererUid(uid)
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) ReInvite() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get docid from url
		docId := c.Param("docId")
		if docId == "" {
			log.Println("[ERROR]- ReSendInvite - invalid docid:")
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "docId not exist")
			return
		}

		ctx := context.Background()
		batch := h.apiContext.Store.Batch()

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)

		doc, err := h.apiContext.Store.Doc("referrals/" + uid + "/invite/" + docId).Get(context.Background())
		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not find invite document:", err, uid)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "docId not exist")
			return
		}

		// Referer user
		title, content, contents := GenReminderSender(doc.Data()["name"].(string), doc.Data()["lname"].(string))
		mailGeneral, err := h.apiContext.Cache.GetNotice(uid, "MailGeneral")
		if err != nil {
			log.Println("Can not get MailWallet setting from cache:", err)
		} else {
			// check setting and send mail
			if mailGeneral == "1" {
				err = mail.SendNoticeMail(userInfo["Email"].(string), userInfo["Name"].(string), title, contents)
				if err != nil {
					log.Println("[ERROR]- Invite - can not send mail invite:", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
					return
				}
			}
		}

		// App notice
		notice := map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   time.Now().Unix(),
		}
		docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
		batch.Set(docRef, notice)

		// Referral user
		title, _, contents = GenReminder(uid, userInfo["Name"].(string), userInfo["LName"].(string), doc.Ref.ID)
		err = mail.SendNoticeMail(doc.Data()["email"].(string), doc.Data()["name"].(string), title, contents)
		if err != nil {
			log.Println("[ERROR]- Invite - can not send mail invite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "email in used")
			return
		}

		// App notice
		// notice = map[string]interface{}{
		// 	"title":  title,
		// 	"body":   content,
		// 	"isRead": false,
		// 	"time":   time.Now().Unix(),
		// }
		// docRef = h.apiContext.Store.Collection("notices").Doc("general").Collection(refererUid).NewDoc()
		// batch.Set(docRef, notice)

		refDoc := h.apiContext.Store.Doc("referrals/" + uid + "/invite/" + docId)
		batch.Update(refDoc, []firestore.Update{
			{Path: "remindTime", Value: firestore.Increment(1)},
			{Path: "lastSentRemind", Value: time.Now().Unix()},
		})

		_, err = batch.Commit(ctx)

		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not batch commit reinvite:", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "")
			return
		}
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}
func (h UserHandler) DelInvite() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get docid from url
		docId := c.Param("docId")
		if docId == "" {
			log.Println("[ERROR]- ReSendInvite - invalid docid:")
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "docId not exist")
			return
		}
		ctx := context.Background()

		uid := c.GetString(UID)

		pendingIniteDoc, err := h.apiContext.Store.Doc("referrals/" + uid + "/invite/" + docId).Get(ctx)
		if err == nil {
			if sendgridId, ok := pendingIniteDoc.Data()["sendGridId"]; ok {
				mail.RemoveRecipientFromList(sendgridId.(string), 12592770)
			}
		}

		_, err = h.apiContext.Store.Doc("referrals/" + uid + "/invite/" + docId).Delete(ctx)
		if err != nil {
			log.Println("[ERROR]- ReSendInvite - can not find invite document:", err, uid)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "docId not exist")
			return
		}
		GinRespond(c, http.StatusOK, SUCCESS, "")
	}
}

func (h UserHandler) TxVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Ledger int64                  `json:"ledger"`
			TxHash string                 `json:"txHash"`
			Action string                 `json:"action"`
			Algo   string                 `json:"algo,omitempty"`
			Data   map[string]interface{} `json:"data,omitempty"`
		}

		err := c.BindJSON(&request)
		if err != nil {
			log.Println("BindJSON error: ", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Error can not parse input data")
			return
		}
		log.Println("request: ", request)
		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		payments, err := ParsePaymentFromTxHash(request.TxHash, stellar.GetHorizonClient())
		if err != nil {
			time.Sleep(500)
			payments, err = ParsePaymentFromTxHash(request.TxHash, stellar.GetHorizonClient())
		}
		if len(payments) == 0 {
			if err != nil {
				log.Printf("Can not query payment from tx hash %s. Error: %v. Afer re-try three times\n", request.TxHash, err)
				GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse ledger information")
				return
			}
		}
		hasPayment := false
		paymentId := ""
		var from, to string
		var amount float64 = 0
		for _, payment := range payments {
			if payment.From == userInfo["PublicKey"].(string) {
				hasPayment = true
				from = payment.From
				to = payment.To
				amount, _ = strconv.ParseFloat(payment.Amount, 64)
				paymentId = payment.ID
				break
			}
		}
		if !hasPayment {
			log.Printf("Can not query ledger %s. Error: %v. Afer re-try three times\n", request.TxHash, err)
			GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse tx hash not belong to user")
			return
		}

		// url := h.apiContext.Config.HorizonUrl + fmt.Sprintf("ledgers/%d/payments", request.Ledger)
		// log.Println("url payment:", url)
		// from, to, amount, err := GetLedgerInfo(url, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerAddress)
		// if err != nil {
		// 	log.Printf("Can not query ledger %d. Error: %v. Will re-try\n", request.Ledger, err)
		// 	// re-try
		// 	for i := 0; i < 3; i++ {
		// 		time.Sleep(1000)
		// 		_, _, amount, err = GetLedgerInfo(url, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerAddress)
		// 		if err == nil {
		// 			break
		// 		}
		// 	}
		// 	if err != nil {
		// 		log.Printf("Can not query ledger %d. Error: %v. Afer re-try three times\n", request.Ledger, err)
		// 		GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse ledger information")
		// 		return
		// 	}
		// }
		log.Println("amount:", from, to, amount)
		switch request.Action {
		case "payoff":
			if to == h.apiContext.Config.XlmLoanerAddress && amount >= 2.1 {
				// Set IsLoan to true
				_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"LoanPaidStatus": 2}, firestore.MergeAll)
				if err != nil {
					log.Printf(uid+": Set IsLoand false error %v\n", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
					return
				}

				_, _, err = stellar.RemoveSigner(from, h.apiContext.Config.XlmLoanerSeed)
				if err != nil {
					log.Println("Can not remove signer", err)
				}

				// remove from sendgrid

				GinRespond(c, http.StatusOK, SUCCESS, "")
				return
			} else {
				log.Println("not set isloan:")
				GinRespond(c, http.StatusOK, TX_FAIL, "Transaction is invalid")
			}
		case "buying":
			if to == h.apiContext.Config.SuperAdminAddress {
				// send fund from super admin account to 'from' account
				// Set price

				var grxPrice, grxAmount, xlmAmount float64
				var ok bool
				grxPrice, ok = request.Data["grxPrice"].(float64)
				if !ok {
					log.Println("Can not get grx price", err)
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					return
				}
				grxAmount, ok = request.Data["grxAmount"].(float64)
				if !ok {
					log.Println("Can not get grxAmount", err)
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					return
				}
				xlmAmount, ok = request.Data["xlmAmount"].(float64)
				if !ok {
					log.Println("Can not get xlmAmount", err)
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					return
				}
				grxUsd, ok := request.Data["grxUsd"].(float64)
				if !ok {
					log.Println("Can not get grxUsd", err)
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					return
				}
				totalUsd, ok := request.Data["totalUsd"].(float64)
				if !ok {
					log.Println("Can not get totalUsd", err)
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					return
				}

				log.Println("grxXlmPrice, grxAmount, xlmAmount: ", grxPrice, grxAmount, xlmAmount)

				if grxPrice < h.apiContext.Config.SellingPrice {
					log.Println("txverify - price is lower than setting", grxPrice, h.apiContext.Config.SellingPrice)
					GinRespond(c, http.StatusOK, PRICE_LOWER_LIMIT, "")
					return
				}

				if amount != xlmAmount {
					GinRespond(c, http.StatusOK, PRICE_LOWER_LIMIT, "")
					log.Println("tx PRICE_LOWER_LIMIT")
					return
				}

				_, _, err = stellar.SendAsset(from, grxAmount, h.apiContext.Config.SuperAdminSeed,
					build.CreditAsset{Code: h.apiContext.Config.AssetCode, Issuer: h.apiContext.Config.IssuerAddress}, "")
				if err != nil {
					GinRespond(c, http.StatusOK, TX_FAIL, "")
					// Need to store txid and account information for checking
				} else {
					log.Println("tx success")
					GinRespond(c, http.StatusOK, SUCCESS, "")
				}

				// Save to trade collection
				data := map[string]interface{}{
					"time":     time.Now().Unix(),
					"type":     "BUY",
					"asset":    "GRX",
					"amount":   grxAmount, //
					"xlmp":     grxPrice,
					"totalxlm": xlmAmount, //
					"priceusd": grxUsd,
					"totalusd": totalUsd,
					"offerId":  paymentId,
				}

				docRef := h.apiContext.Store.Collection("trades").Doc("users").Collection(uid).NewDoc()
				_, err = docRef.Set(context.Background(), data)
				if err != nil {
					log.Println("[ERROR] SaveNotice: ", err)
					return
				}
				data["id"] = docRef.ID
				data["uid"] = uid
				_, err = h.apiContext.OrderIndex.SaveObject(data)
				if err != nil {
					log.Println("[ERROR] Algolia OrderIndex.SaveObject: ", err)
					return
				}

			}
		case "open":
			// userInfo, _ := GetUserByField(h.apiContext.Store, "PublicKey", q)
			// if userInfo == nil {
			// 	c.JSON(http.StatusNotFound, output)
			// 	return
			// }

			// output.StellarAddress = userInfo["Federation"].(string)
			// output.AccountId = userInfo["PublicKey"].(string)
			// c.JSON(http.StatusOK, output)
		case "close":

		}
		//GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse ledger information")
	}
}

func ReconnectFireStore(projectId string, timeout int) (*firestore.Client, error) {
	cnt := 0
	var client *firestore.Client
	var err error
	ctx := context.Background()
	for {
		cnt++
		time.Sleep(1 * time.Second)
		client, err = firestore.NewClient(ctx, projectId)
		if err == nil {
			break
		}
		if cnt > timeout {
			log.Println("[ERROR] Can not connect to firestore after retry times", cnt, projectId, err)
			break
		}

	}
	return client, err
}
