package api

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"

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
	"github.com/asaskevich/govalidator"
	"github.com/avct/uasurfer"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	stellar "github.com/huyntsgs/stellar-service"
	"google.golang.org/api/iterator"
	//"github.com/huyntsgs/stellar-service/assets"
	//"github.com/go-redis/redis"
)

const (
	ConfirmRegistrationSub = "Please Confirm Your GRAYLL App Registration Request"
	ConfirmIpSub           = "GRAYLL | New IP Address Verification"
	LoginSuccess           = "GRAYLL | Account Login Successful"
	ResetPasswordSub       = "GRAYLL | Reset Password Verification"
	ChangeEmailSub         = "GRAYLL | Confirm change email request"
	RevealSecretTokenSub   = "GRAYLL | Reveal Secret Token"

	VerifyEmail       = "verifyEmail"
	ResetPassword     = "resetPassword"
	ChangeEmail       = "changeEmail"
	ConfirmIp         = "confirmIp"
	UID               = "Uid"
	RevealSecretToken = "revealSecretToken"
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
		var gd geoIPData
		gd.Country = c.GetHeader("X-AppEngine-Country")
		gd.Region = c.GetHeader("X-AppEngine-Region")
		gd.City = c.GetHeader("X-AppEngine-City")
		log.Println("GeoIp data:", gd)

		currentIp := utils.RealIP(c.Request)
		setting, ok := userInfo["Setting"].(map[string]interface{})
		if !ok {
			log.Println("Can not parse user setting. userInfo: ", userInfo)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not parse user data")
			return
		}
		// ipConfirm := setting["IpConfirm"].(bool)
		// if currentIp != userInfo["Ip"].(string) && ipConfirm {
		// 	ipTemp, ok := userInfo["IpTemp"].(string)
		// 	if !ok || (ok && ipTemp == "") {
		// 		// Send confirm Ip mail
		// 		log.Println(`!ok || (ok && ipTemp != "")`)
		// 		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, currentIp+"?"+uid)
		// 		if encodeStr == "" {
		// 			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not login right now. Please try again later.")
		// 			return
		// 		}
		// 		city, country := utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
		// 		mores := map[string]string{
		// 			"loginTime": time.Now().Format("Mon, 02 Jan 2006 15:04:05 UTC"),
		// 			"ip":        currentIp,
		// 			"agent":     c.Request.UserAgent(),
		// 			"city":      city,
		// 			"country":   country,
		// 		}
		// 		err = mail.SendMail(userInfo["Email"].(string), userInfo["Name"].(string), ConfirmIpSub, ConfirmIp, encodeStr, h.apiContext.Config.Host, mores)
		// 		if err != nil {
		// 			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not login right now. Please try again later.")
		// 			return
		// 		}
		// 		GinRespond(c, http.StatusOK, IP_CONFIRM, "Need to confirm Ip before login")
		// 		return
		// 	} else if ok && currentIp == ipTemp {
		// 		log.Println(`ok && ipTemp == currentIp. Set IpTemp ""`)
		// 		_, err := h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"IpTemp": ""}, firestore.MergeAll)
		// 		if err != nil {
		// 			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not update temp ip")
		// 			return
		// 		}
		// 	}
		// }
		city, country := utils.GetCityCountry("http://www.geoplugin.net/json.gp?ip=" + currentIp)
		ua := uasurfer.Parse(c.Request.UserAgent())
		//log.Println("ua:", ua)
		agent := fmt.Sprintf("Device - %s, Browser - %s, OS - %s.", ua.DeviceType.StringTrimPrefix(), ua.Browser.Name.StringTrimPrefix(), ua.OS.Name.StringTrimPrefix())
		res := make(chan int)
		ctx := context.Background()
		go func(res chan int) {
			ipConfirm := setting["IpConfirm"].(bool)
			if currentIp != userInfo["Ip"].(string) && !ipConfirm {
				secondIp, ok := userInfo["SecondIp"]
				if ok && currentIp != secondIp.(string) {
					// already set second ip, warning email
					log.Println("Ip is not matched. Sent warning email")
					res <- 0
				}
			} else if currentIp != userInfo["Ip"].(string) && ipConfirm {
				secondIp, ok := userInfo["SecondIp"]
				// secondIp still may not be set
				if !ok || (ok && currentIp != secondIp.(string)) {
					//if !ok || (ok && ipTemp == "") {
					// Send confirm Ip mail
					encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, currentIp+"?"+uid)
					if encodeStr == "" {
						GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not login right now. Please try again later.")
						res <- 1
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
						GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not login right now. Please try again later.")
						res <- 1
						return
					}
					GinRespond(c, http.StatusOK, IP_CONFIRM, "Need to confirm Ip before login")
					res <- 1
				}
			}
			res <- 0
		}(res)
		go func() {
			if subs, ok := userInfo["Subs"]; ok {
				s, err := json.Marshal(subs)
				if err != nil {
					log.Println("Can not find parse subs:", err)
				}
				h.apiContext.Cache.SetUserSubs(uid, string(s))
				h.apiContext.Cache.SetPublicKey(uid, userInfo["PublicKey"].(string))

				settingFields := []string{"IpConfirm", "MulSignature", "AppGeneral", "AppWallet", "AppAlgo", "MailGeneral", "MailWallet", "MailAlgo"}
				for _, field := range settingFields {
					if val, ok := setting[field]; ok {
						h.apiContext.Cache.SetNotice(uid, field, val.(bool))
					}
				}
			}
		}()
		val := <-res
		if val == 1 {
			return
		}
		//userInfo["Uid"] = uid
		//if _, ok := userInfo["LoginTime"]; !ok {
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
		//}
		go func() {
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
		if _, ok := userInfo["Subs"]; ok {
			userInfo["Subs"] = true
		}
		tokenStr, err := h.apiContext.Jwt.GenToken(uid, 60)

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS, "user": userInfo, "token": tokenStr, "tokenExpiredTime": (time.Now().Unix() + int64(24*60*60-5)),
		})
	}
}

// Register handles register router.
// Function validates parameters and call Register from UserStore.
func (h UserHandler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.UserInfo
		ctx := context.Background()
		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
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

		userInfo, _ := GetUserByField(h.apiContext.Store, "Email", input.Email)
		if userInfo != nil {
			GinRespond(c, http.StatusOK, EMAIL_IN_USED, "Email already registered")
			return
		}

		// Get IP of user at time registration
		//input.Token = ""
		input.Federation = input.Email + "*grayll.io"
		input.Ip = utils.RealIP(c.Request)
		input.CreatedAt = time.Now().Unix()
		input.Setting = models.Settings{IpConfirm: true, MulSignature: true, AppAlgo: true, AppWallet: true, AppGeneral: true,
			MailAlgo: true, MailWallet: true, MailGeneral: true}
		hash, err := utils.DerivePassphrase(input.HashPassword, 32)
		if err != nil {
			log.Printf("AddUserData:Add error %v\n", err)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}
		input.HashPassword = hash
		docRef, _, err := h.apiContext.Store.Collection("users").Add(ctx, input)
		if err != nil {
			log.Printf("AddUserData:Add error %v\n", err)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}
		uid := docRef.ID
		encodeStr := utils.EncryptItem(h.apiContext.Jwt.PublicKey, input.Email)
		if encodeStr == "" {
			_, err = h.apiContext.Store.Doc("users/" + uid).Delete(ctx)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}

		err = mail.SendMail(input.Email, input.Name, ConfirmRegistrationSub, VerifyEmail, encodeStr, h.apiContext.Config.Host, nil)
		if err != nil {
			_, err = h.apiContext.Store.Doc("users/" + uid).Delete(ctx)
			GinRespond(c, http.StatusInternalServerError, INTERNAL_ERROR, "Can not register right now")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}
}

func (h UserHandler) GetNotices() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}
		notices := make([]map[string]interface{}, 0)
		it := h.apiContext.Store.Collection("notices/users/"+uid).Limit(200).OrderBy("time", firestore.Desc).Documents(context.Background())
		for {
			doc, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Println("err reading db: ", err)
			}

			data := map[string]interface{}{
				"id":     doc.Ref.ID,
				"type":   doc.Data()["type"].(string),
				"title":  doc.Data()["title"].(string),
				"body":   doc.Data()["body"].(string),
				"time":   doc.Data()["time"].(int64),
				"txId":   doc.Data()["txId"].(string),
				"isRead": doc.Data()["isRead"].(bool),
			}
			notices = append(notices, data)
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS, "notices": notices,
		})
	}
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
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusOK, INVALID_PARAMS, err.Error())
			return
		}
		ctx := context.Background()
		if len(input.WalletIds) > 0 {
			for _, id := range input.WalletIds {
				h.apiContext.Store.Doc("notices/users/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}
			cnt := 0 - len(input.WalletIds)
			_, err := h.apiContext.Store.Doc("users/"+uid).Update(ctx, []firestore.Update{
				{Path: "UrWallet", Value: firestore.Increment(cnt)},
			})
			if err != nil {
				log.Println("SaveNotice update error: ", err)
				return
			}
		}
		if len(input.AlgoIds) > 0 {
			for _, id := range input.WalletIds {
				h.apiContext.Store.Doc("notices/users/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}
			cnt := 0 - len(input.AlgoIds)
			_, err := h.apiContext.Store.Doc("users/"+uid).Update(ctx, []firestore.Update{
				{Path: "UrAlgo", Value: firestore.Increment(cnt)},
			})
			if err != nil {
				log.Println("SaveNotice update error: ", err)
				return
			}
		}
		if len(input.GeneralIds) > 0 {
			for _, id := range input.GeneralIds {
				h.apiContext.Store.Doc("notices/users/"+uid+"/"+id).Set(ctx, map[string]interface{}{
					"isRead": true,
				}, firestore.MergeAll)
			}
			cnt := 0 - len(input.GeneralIds)
			_, err := h.apiContext.Store.Doc("users/"+uid).Update(ctx, []firestore.Update{
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
			_, err = h.apiContext.Cache.SetUserSubs(uid, string(rawData))
			if err != nil {
				// GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
				// return
				log.Println("Can not save the subs to cache: ", err)
			}

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
		log.Println("Login; user: ", changeMail)

		uid := c.GetString("Uid")

		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
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
				_, err := h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"IsVerified": false, "Email": items[1]}, firestore.MergeAll)
				if err != nil {
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Error set IsVerified false")
					return
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
		err = mail.SendMail(input.Email, input.Name, ResetPasswordSub, ResetPassword, encodeStr, h.apiContext.Config.Host, nil)
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
			"HashPassword": hash,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("AddUserData:Add error 1%v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, "Can not register right now")
			return
		}
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
			seq, hash, err := stellar.SendXLMCreateAccount(input.PublicKey, float64(1.50001), h.apiContext.Config.XlmLoanerSeed)
			if err != nil {
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
					time.Sleep(1000)
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
		activatedData := map[string]interface{}{
			"PublicKey":      input.PublicKey,
			"EnSecretKey":    input.EnSecretKey,
			"SecretKeySalt":  input.Salt,
			"LoanPaidStatus": 1,
			"ActivatedAt":    time.Now().Unix(),
			// "Setting": map[string]interface{}{
			// 	"IpConfirm": true, "MulSignature": true, "AppGeneral": true, "AppWallet": true, "AppAlgo": true, "MailGeneral": true, "MailWallet": true, "MailAlgo": true,
			// },
		}
		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), activatedData, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set activated data error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		// Set PublicAddress to cache
		go func() {
			_, err = h.apiContext.Cache.SetPublicKey(input.PublicKey, uid)
			if err != nil {
				log.Printf(uid+": SetPublicKey cache error %v\n", err)
			}

			// Set setting cache
			data := map[string]bool{"IpConfirm": true, "MulSignature": true, "AppGeneral": true, "AppWallet": true,
				"AppAlgo": true, "MailGeneral": true, "MailWallet": true, "MailAlgo": true}
			for k, v := range data {
				_, err = h.apiContext.Cache.SetNotice(uid, k, v)
				if err != nil {
					log.Printf(uid+": SetNotice cache error %v\n", err)
				}
			}
			// _, err = h.apiContext.Cache.SetNotices(uid, "IpConfirm", "1", "MulSignature", "1", "AppGeneral", "1", "AppWallet", "1",
			// 	"AppAlgo", "1", "MailGeneral", "1", "MailWallet", "1", "MailAlgo", "1")
			// if err != nil {
			// 	log.Printf(uid+": SetNotice cache error %v\n", err)
			// }

		}()

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
		if _, ok := input["isloan"]; ok {
			if value, ok := userInfo["IsLoan"]; ok {
				res["isloan"] = value.(bool)
			} else {
				res["isloan"] = false
			}
		}
		res["errCode"] = SUCCESS
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
		log.Println("reveal token 1:", revealToken)
		// Set token to database
		_, err := h.apiContext.Store.Doc("resetpwd/"+uid).Set(context.Background(), map[string]interface{}{
			"reavealtoken": revealToken,
			"expire":       time.Now().Unix() + int64(5*60),
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
		fmt.Printf("uid %v\n", uid)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			//GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "User does not exist.")
			output = Output{Valid: false, ErrCode: INVALID_UNAME_PASSWORD, Message: "User does not exist."}
			c.JSON(http.StatusOK, output)
			return
		}
		tfa, ok := userInfo["Tfa"].(map[string]interface{})
		if tfa == nil || !ok {
			fmt.Println("Tfa is nil:", err)
			output = Output{Valid: false, ErrCode: INTERNAL_ERROR, Message: "Tfa is not enabled."}
			c.JSON(http.StatusOK, output)
			return
		}

		fmt.Printf("input %v\n", input)

		otpc := &dgoogauth.OTPConfig{
			//Secret:      input.Secret,
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
		output = Output{Valid: true}
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

func (h UserHandler) TxVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Ledger int64  `json:"ledger"`
			Action string `json:"action"`
			Algo   string `json:"algo,omitempty"`
		}

		err := c.BindJSON(&request)
		if err != nil {
			log.Println("BindJSON error: ", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Error can not parse input data")
			return
		}

		uid := c.GetString(UID)
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		url := h.apiContext.Config.HorizonUrl + fmt.Sprintf("ledgers/%d/payments", request.Ledger)
		log.Println("url payment:", url)
		_, _, amount, err := GetLedgerInfo(url, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerAddress)
		if err != nil {
			log.Printf("Can not query ledger %d. Error: %v. Will re-try\n", request.Ledger, err)
			// re-try
			for i := 0; i < 3; i++ {
				time.Sleep(1000)
				_, _, amount, err = GetLedgerInfo(url, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerAddress)
				if err == nil {
					break
				}
			}
			if err != nil {
				log.Printf("Can not query ledger %d. Error: %v. Afer re-try three times\n", request.Ledger, err)
				GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse ledger information")
				return
			}
		}
		log.Println("amount:", amount)
		switch request.Action {
		case "payoff":
			if amount >= 1.5 {
				// Set IsLoan to true
				_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"LoanPaidStatus": 2}, firestore.MergeAll)
				if err != nil {
					log.Printf(uid+": Set IsLoand false error %v\n", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
					return
				}
				log.Println("set isloan:")
				GinRespond(c, http.StatusOK, SUCCESS, "")
				return
			} else {
				log.Println("not set isloan:")
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
		GinRespond(c, http.StatusOK, INVALID_CODE, "Can not parse ledger information")
	}
}
