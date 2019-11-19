package api

import (
	"bytes"
	"log"

	//"context"
	//"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	//jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"
	//"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

type PhoneHandler struct {
	apiContext *ApiContext
}

var tokenMap *sync.Map = new(sync.Map)

// Creates new UserHandler.
// UserHandler accepts interface UserStore.
// Any data store implements UserStore could be the input of the handle.
func NewPhoneHandler(apiContext *ApiContext) PhoneHandler {
	return PhoneHandler{apiContext: apiContext}
}

// https://www.googleapis.com/identitytoolkit/v3/relyingparty/sendVerificationCode?key=api_key' \
//   -H 'content-type: application/json' \
//   -d '{
//  "phoneNumber": "phone_number_to_verify",
//  "recaptchaToken": "generated_recaptcha_token"
// }'
func (h PhoneHandler) SendSms() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			PhoneNumber    string `json:"phoneNumber"`
			RecaptchaToken string `json:"recaptchaToken"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "")
			return
		}
		log.Println("input:", input)
		data, err := json.Marshal(input)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INTERNAL_ERROR, "can not marshal json input")
			return
		}
		//url := fmt.Sprintf(API_URL, "sendVerificationCode")
		res, err := http.Post("https://www.googleapis.com/identitytoolkit/v3/relyingparty/sendVerificationCode?key=AIzaSyBoGmsWYP-CsATX8c1sW8AAc4ua4bl3_SY",
			"application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Println("error call sendVerificationCode:", err)
			GinRespond(c, http.StatusBadRequest, INTERNAL_ERROR, "can not marshal json input")
			return
		} else {
			log.Println("sendVerificationCode body:", res)
		}
		uid := c.GetString("Uid")
		var output struct {
			SessionInfo string `json:"sessionInfo"`
		}

		err = json.NewDecoder(res.Body).Decode(&output)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INTERNAL_ERROR, "can not decode json")
			return
		}

		log.Println("output:", output)

		tokenMap.Store(uid, output.SessionInfo)

		c.JSON(http.StatusOK, gin.H{
			"valid": true, "errCode": SUCCESS,
		})
	}
}

// https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPhoneNumber?key=api_key' \
//   -H 'content-type: application/json' \
//   -d '{
//   "sessionInfo": "session_token",
//   "code":"sms_code"
func (h PhoneHandler) VerifyCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Code        string `json:"code"`
			SessionInfo string `json:"sessionInfo,omitempty"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "")
			return
		}
		uid := c.GetString("Uid")
		sessionInfo, ok := tokenMap.Load(uid)
		if !ok {
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "")
			return
		}
		input.SessionInfo = sessionInfo.(string)
		log.Println("VerifyCode input:", input)
		data, err := json.Marshal(input)
		if err != nil {
			GinRespond(c, http.StatusBadRequest, INTERNAL_ERROR, "can not marshal json input")
			return
		}
		url := fmt.Sprintf(API_URL, "verifyPhoneNumber")
		res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		log.Println("res:", res)

		// err = json.NewDecoder(res.Body).Decode(&output)
		// if err != nil {
		// 	GinRespond(c, http.StatusBadRequest, INTERNAL_ERROR, "can not decode json")
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"valid": true, "errCode": SUCCESS,
		})
	}
}
