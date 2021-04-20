package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"

	//"encoding/json"

	"log"
	"net/http"
	"time"

	//"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	"cloud.google.com/go/firestore"

	//"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/timestamp"
	stellar "github.com/huyntsgs/stellar-service"
	"google.golang.org/api/iterator"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	//"github.com/mitchellh/mapstructure"
	"github.com/fatih/structs"
	//"github.com/jinzhu/now"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

const ()

type PriceData struct {
	Ts    int64   `json:"ts"`
	Price float64 `json:"price"`
}

const (

	// new sub collection
	frame01m = "frame_01m"
	frame05m = "frame_05m"
	frame15m = "frame_15m"
	frame30m = "frame_30m"
	frame01h = "frame_01h"
	frame04h = "frame_04h"
	frame01d = "frame_01d"
	frame01w = "frame_01w"
	frame1mo = "frame_1mo"

	// new col
	xrpusd_01m = "xrpusd_01m"
	xrpusd_05m = "xrpusd_05m"

	gryusd_05m = "gryusd_05m"
	gryusd_15m = "gryusd_15m"
	gryusd_30m = "gryusd_30m"
	gryusd_01h = "gryusd_01h"
)

const (
	// gray price
	UNIX_timestamp = "UNIX_timestamp"
	price          = "price"
	pair           = "pair"
	rate           = "rate"
)

const (
	// time interval of frames in minutes
	n01m = 1
	n05m = 5
	n15m = 15
	n30m = 30
	n01h = 60
	n04h = 240
	n01d = 1440
	n01w = 10080
	n1mo = 43200
)

// type UserHandler struct {
// 	apiContext *ApiContext
// }
//
// type geoIPData struct {
// 	Country string
// 	Region  string
// 	City    string
// }

func (h UserHandler) GetFramesData() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Limit int    `json:"limit"`
			Coins string `json:"coins"`
			Frame string `json:"frame"`
		}
		mt := new(sync.Mutex)
		res := make(map[string][]PriceData, 0)

		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}
		fmt.Printf("got input data: %v\n", input)
		coins := strings.Split(input.Coins, ",")
		wg := new(sync.WaitGroup)
		wg.Add(len(coins))
		for _, pair := range coins {
			if input.Limit == 1 {
				go func(pairstr string) {
					defer wg.Done()
					prices := QueryFrameData(h.apiContext.Store, input.Limit, pairstr, input.Frame)
					mt.Lock()
					res[pairstr] = prices
					mt.Unlock()
				}(pair)
			} else {
				go func(pairstr string) {
					defer wg.Done()
					prices := QueryFrameDataWithTs(h.apiContext.Store, input.Limit, pairstr, input.Frame)
					mt.Lock()
					res[pairstr] = prices
					mt.Unlock()
				}(pair)
			}
		}
		wg.Wait()

		//log.Println("res", res)

		c.JSON(http.StatusOK, gin.H{
			"status": "success", "res": res,
		})
	}
}

func (h UserHandler) GetFramesDataGet() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Limit int    `json:"limit"`
			Coins string `json:"coins"`
			Frame string `json:"frame"`
		}
		var err error
		log.Println(c.Params)
		input.Limit, err = strconv.Atoi(c.Param("limit"))
		if err != nil || input.Limit > 300 {
			log.Println("Parse limit err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		input.Coins = c.Param("coins")
		if err != nil {
			log.Println("Parse coin err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}
		input.Frame = c.Param("frame")
		if err != nil {
			log.Println("Parse frame err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		mt := new(sync.Mutex)
		res := make(map[string][]PriceData, 0)

		// err = c.BindJSON(&input)
		// if err != nil {
		// 	log.Println("BindJSON err:", err)
		// 	GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
		// 	return
		// }
		// fmt.Printf("got input data: %v\n", input)
		coins := strings.Split(input.Coins, ",")
		wg := new(sync.WaitGroup)
		wg.Add(len(coins))
		for _, pair := range coins {
			if input.Limit == 1 {
				go func(pairstr string) {
					defer wg.Done()
					prices := QueryFrameData(h.apiContext.Store, input.Limit, pairstr, input.Frame)
					mt.Lock()
					res[pairstr] = prices
					mt.Unlock()
				}(pair)
			} else {
				go func(pairstr string) {
					defer wg.Done()
					prices := QueryFrameDataWithTs(h.apiContext.Store, input.Limit, pairstr, input.Frame)
					mt.Lock()
					res[pairstr] = prices
					mt.Unlock()
				}(pair)
			}
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"status": "success", "res": res,
		})
	}
}

func (h UserHandler) MakeTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			XDR string `json:"xdr"`
			TX  string `json:"tx"`
		}

		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}

		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		if input.TX == "loanpaid" {
			txrs, err := stellar.ParseLoanXDR(input.XDR, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerSeed, h.apiContext.Config.XlmLoanerAddress, float64(2.0001))
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"errCode": txrs,
				})
			} else {
				_, err := h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"LoanPaidStatus": 2}, firestore.MergeAll)
				if err != nil {
					log.Printf(uid+": Set IsLoand false error %v\n", err)
					GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
					return
				}
				log.Println("isloan:")
				// _, _, err = stellar.RemoveSigner(userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerSeed)
				// if err != nil {
				// 	log.Println("Can not remove signer", err)
				// }
				c.JSON(http.StatusOK, gin.H{
					"errCode": txrs,
				})
			}

		} else {
			txrs, err := stellar.ParseXDR(input.XDR)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"errCode": txrs,
				})
			} else {

				c.JSON(http.StatusOK, gin.H{
					"errCode": txrs,
				})
			}
		}
	}
}
func (h UserHandler) PayLoan() gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, "Uid", uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, INVALID_UNAME_PASSWORD, "Invalid user name or password")
			return
		}

		pk := userInfo["PublicKey"].(string)

		_, _, err := stellar.PayLoan(pk, h.apiContext.Config.XlmLoanerSeed)
		if err != nil {
			log.Println("ERROR - PayLoan - unable to pay loan", uid, err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}
		log.Println("Paid loan", uid)
		_, err = h.apiContext.Store.Doc("users/"+uid).Set(context.Background(), map[string]interface{}{"LoanPaidStatus": 2}, firestore.MergeAll)
		if err != nil {
			log.Printf(uid+": Set IsLoand false error %v\n", err)
			GinRespond(c, http.StatusOK, INTERNAL_ERROR, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})
	}
}

func (h UserHandler) UpdateHomeDomain() gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("Uid")
		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)

		// move user data to backup
		_, err := h.apiContext.Store.Doc("accounts_exported/"+uid).Set(context.Background(), userInfo)
		if err != nil {
			log.Println("Can not set account closure:", err)
		} else {
			_, err := h.apiContext.Store.Doc("users/" + uid).Delete(context.Background())
			if err != nil {
				log.Println("Can not delete account:", err)
			}
		}

		// Remove form xlm loan sendgrid list
		if receiptId, ok := userInfo["SendGridId"]; ok {
			mail.RemoveRecipientFromList(receiptId.(string), 10196670)
			mail.RemoveRecipientFromList(receiptId.(string), 10761027)
			// Add to account closure list
			mail.AddRecipienttoList(receiptId.(string), 12586061)
		}

		c.Status(200)
	}
}

func (h UserHandler) XlmLoanReminder() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract the request body for further task details.
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("ReadAll: %v", err)
			//http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		data := make(map[string]interface{})
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Println("error parse task data")
			return
		}
		//log.Println("Task data:", data)
		uid := data["uid"].(string)

		orderId := int64(data["orderId"].(float64))
		activatedAt := int64(data["activatedAt"].(float64))

		userInfo, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if userInfo == nil {
			GinRespond(c, http.StatusOK, EMAIL_NOT_EXIST, "Account does not exist.")
			return
		}

		loanPaidStatus := userInfo["LoanPaidStatus"].(int64)
		if loanPaidStatus == 1 {
			log.Println("Send mail and push notice")
			// Send mail and push notice
			if orderId < 40 {
				content := genReminderContent(orderId)
				title := "GRAYLL | XLM Loan Repayment Reminder"
				mail.SendLoanReminder(userInfo["Email"].(string), userInfo["Name"].(string), title, h.apiContext.Config.Host, content, true)

				// app notice
				body := ""
				for _, con := range content {
					body = body + con
				}
				url := fmt.Sprintf("%s/dashboard/overview/(popup:xlm-loan)", h.apiContext.Config.Host)
				// Send app and push notices
				notice := map[string]interface{}{
					"type":   "general",
					"title":  title,
					"isRead": false,
					"url":    url,
					"body":   body,
					"time":   time.Now().Unix(),
					// "vibrate": []int32{100, 50, 100},
					// "icon":    "https://app.grayll.io/favicon.ico",
					// "data": map[string]interface{}{
					// 	"url": h.apiContext.Config.Host + "/notifications/overview",
					// },
				}

				// Save to firestore
				ctx := context.Background()
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

				// check loan paid status again
				// if not paid, schedule to send reminder

				createLoanReminder(uid, orderId+1, activatedAt)
				mail.SaveLoanPaidInfo(userInfo["Name"].(string), userInfo["LName"].(string), userInfo["Email"].(string), "no", userInfo["CreatedAt"].(int64), orderId)
			} else {
				content := genReminderContent(orderId)
				title := "GRAYLL | Account Closure Reminder"
				mail.SendLoanReminder(userInfo["Email"].(string), userInfo["Name"].(string), title, h.apiContext.Config.Host, content, false)

				// merge account
				err := MergeAccount(userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerSeed)

				if err != nil {
					log.Println("ERROR unable MergeAccount error:", uid, userInfo["PublicKey"].(string), err)
				}

				// move user data to backup
				_, err = h.apiContext.Store.Doc("accounts_closure/"+uid).Set(context.Background(), userInfo)
				if err != nil {
					log.Println("Can not set account closure:", err)
				} else {
					_, err := h.apiContext.Store.Doc("users/" + uid).Delete(context.Background())
					if err != nil {
						log.Println("Can not delete account:", err)
					}
				}

				// Remove form xlm loan sendgrid list
				if receiptId, ok := userInfo["SendGridId"]; ok {
					mail.RemoveRecipientFromList(receiptId.(string), 10196670)
					mail.RemoveRecipientFromList(receiptId.(string), 10761027)
					// Add to account closure list
					mail.AddRecipienttoList(receiptId.(string), 12586061)
				}

				// app notice
				// body := ""
				// for _, con := range content {
				// 	body = body + con
				// }
				// // Send app and push notices
				// notice := map[string]interface{}{
				// 	"type":    "general",
				// 	"title":   title,
				// 	"isRead":  false,
				// 	"body":    body,
				// 	"time":    time.Now().Unix(),
				// 	"vibrate": []int32{100, 50, 100},
				// 	"icon":    "https://app.grayll.io/favicon.ico",
				// 	"data": map[string]interface{}{
				// 		"url": h.apiContext.Config.Host + "/notifications/overview",
				// 	},
				// }

				//ctx := context.Background()
				// log.Println("Start push notice")
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
			}
		} else {
			go func() {
				mail.SaveLoanPaidInfo(userInfo["Name"].(string), userInfo["LName"].(string), userInfo["Email"].(string), "yes", userInfo["CreatedAt"].(int64), orderId)
			}()
		}

		c.Status(200)
	}
}

func genReminderContent(orderId int64) []string {

	content := make([]string, 0)
	if orderId >= 1 && orderId <= 15 {
		//	1) First 30 days from account creation: send " XLM Loan Repayment Notification" every 48 hours.
		content = []string{
			`GRAYLL has lent you 2.1000 XLM (Stellar Lumens) to activate your Stellar Network Account in the GRAYLL App.
			Certain features, functions and algorithmic services may not be available until the XLM loan is settled.`,

			`If the 2.1000 XLM loan is not settled, your account will eventually be closed.
			We will send you periodic notifications to remind you prior to any account closure.
			We recommend depositing at least 2.50 XLM to your GRAYLL Account, you may pay off the loan now.`,
		}
	} else if orderId >= 16 && orderId <= 25 {
		//2) Next 15 days (30 days from account creation): send "XLM Loan Repayment Notification" every 36 hours.
		days := 15 + (orderId-15)*36/24
		content = []string{

			fmt.Sprintf(`%d days ago GRAYLL lent you 2.1000 XLM (Stellar Lumens) to activate your Stellar Network Account in the GRAYLL App.
		Certain features, functions and algorithmic services may not be available until the XLM loan is settled.`, days),

			`If the 2.1000 XLM loan is not settled, your account will eventually be closed.
		We will send you periodic notifications to remind you prior to any account closure.
		We recommend depositing at least 2.50 XLM to your GRAYLL Account, you may pay off the loan now.`,
		}
	} else if orderId >= 26 && orderId < 40 {
		//3) Next 15 days (45 days from account creation): send "XLM Loan Repayment Notification" every 24 hours.
		days := 15 + 10 + (orderId - 25)
		content = []string{
			fmt.Sprintf(`In %d days your GRAYLL account will be closed if the 2.1000 XLM loan is not settled.
			%d days ago GRAYLL lent you 2.1000 XLM (Stellar Lumens) to activate your Stellar Network Account in the GRAYLL App.
			Certain features, functions and algorithmic services may not be available until the XLM loan is settled.`, days, days),

			`Once your account is closed you will need to sign up again to access the GRAYLL App and the algorithmic services.
			We recommend depositing at least 2.50 XLM to your GRAYLL Account, you may pay off the loan now.`,
		}

	} else if orderId >= 40 {
		//4) After 60 days "Account Closure Notification".
		content = []string{
			`Your GRAYLL account has now been closed as the 2.1000 XLM loan was not repaid within 60 days.
		You will need to sign up again to access the GRAYLL App and the algorithmic services.`,
		}
	}

	//{"Sign Up" BUTTON}
	return content

}

func createLoanReminder(uid string, orderId, activatedAt int64) {
	// create task
	projectId := "grayll-app-f3f3f3"
	queueId := "xlm-loan-reminder"
	local := "us-central1"
	url := "https://grayll-app-bqqlgbdjbq-uc.a.run.app/api/v1/accounts/xlmLoanReminder"
	svAccountEmail := "cloud-tasks-grayll-app@grayll-app-f3f3f3.iam.gserviceaccount.com"
	data := map[string]interface{}{
		"uid":         uid,
		"activatedAt": activatedAt,
		"orderId":     orderId,
	}
	json, _ := json.Marshal(data)
	// TEST
	scheduleTime := activatedAt + getScheduleTime(orderId)
	task, err := createHTTPTask(projectId, local, queueId, url, svAccountEmail, json, scheduleTime)
	if err != nil {
		log.Println("createHTTPTask error:", err)
	} else {
		log.Println("createHTTPTask error:", task.GetScheduleTime())
		log.Println("createHTTPTask task name:", task.GetName())
	}
}

func getScheduleTime(orderId int64) int64 {
	if orderId >= 1 && orderId <= 15 {
		return 48 * 60 * 60 * orderId
	} else if orderId >= 1 && orderId <= 25 {
		return (48*60*60*15 + 36*60*60*(orderId-15))
	} else if orderId >= 26 && orderId < 40 {
		return (48*60*60*15 + 36*60*60*10 + 24*60*60*(orderId-25))
	}
	return 0
}
func getScheduleTimeTest(orderId int64) int64 {
	if orderId >= 1 && orderId <= 15 {
		return 15 * 60 * orderId
	} else if orderId >= 1 && orderId <= 25 {
		return (15*60*15 + 10*60*(orderId-15))
	} else if orderId >= 26 && orderId < 40 {
		return (15*60*15 + 10*60*10 + 5*60*(orderId-25))
	}
	return 0
}
func (h UserHandler) SaveReportSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input ReportDataSetting

		err := c.BindJSON(&input)
		if err != nil {
			log.Printf("[ERROR] unable to parse report data setting : %v", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Update postion error")
			return
		}

		setting := structs.Map(input)

		uid := c.GetString("Uid")
		ctx := context.Background()

		userData, _ := h.apiContext.Store.Doc("users/" + uid).Get(ctx)
		if _reportSetting, ok := userData.Data()["ReportSetting"]; ok {
			reportSetting := _reportSetting.(map[string]interface{})
			if freq, ok := reportSetting["Frequency"]; ok {
				h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"ReportSetting": setting}, firestore.MergeAll)
				if freq.(string) != "None" {
					// already schedule task running
					log.Println("task already running")
					GinRespond(c, http.StatusOK, SUCCESS, "")
					return
				}
			}
		} else {
			h.apiContext.Store.Doc("users/"+uid).Set(ctx, map[string]interface{}{"ReportSetting": setting}, firestore.MergeAll)
		}
		if input.Frequency == "None" {
			GinRespond(c, http.StatusOK, SUCCESS, "")
		}
		log.Println("input:", input)

		current, err := TimeIn(time.Now(), input.TimeZone)
		scheduleTime := NewDate(current, input.TimeHour, input.TimeMinute, input.TimeZone)

		// calculate time to send report
		switch input.Frequency {
		case "Daily":
			if current.Unix() > scheduleTime.Unix() {
				scheduleTime = scheduleTime.Add(time.Hour * 24)
			}
			break
		case "Weekly":
			beginWeek, _ := BeginOfNextWeek(current, input.TimeZone)
			scheduleTime = NewDate(beginWeek, input.TimeHour, input.TimeMinute, input.TimeZone)
			break
		case "Monthly":
			beginMonth, _ := BeginOfNextMonth(current, input.TimeZone)
			scheduleTime = NewDate(beginMonth, input.TimeHour, input.TimeMinute, input.TimeZone)
			break
		}
		data := make(map[string]interface{})
		data["UserId"] = uid
		data["Time"] = scheduleTime.Unix()
		log.Println("schedule time:", scheduleTime.Unix(), scheduleTime.Format("2012-11-01T22:08:41+00:00"))
		h.ScheduleTask(data, scheduleTime.Unix(), h.apiContext.Config.DataReportUrl, h.apiContext.Config.DataReportQueueId)

		GinRespond(c, http.StatusOK, SUCCESS, "")

	}
}
func (h UserHandler) ReportData() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := context.Background()
		setting := make(map[string]interface{})
		err := c.BindJSON(&setting)

		if err != nil {
			log.Printf("[ERROR] unable to unmarshal GRZ update position request's body, error : %v", err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Update postion error")
			return
		}

		uid := setting["UserId"].(string)
		log.Println(setting)
		userData, _ := GetUserByField(h.apiContext.Store, UID, uid)
		if err != nil {
			log.Printf("[ERROR] Can not get user info with user id: %s, %v\n", uid, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Invalid user id")
			return
		}
		reportSettingMap := make(map[string]interface{})
		if _reportSettingMap, ok := userData["ReportSetting"]; ok {
			reportSettingMap = _reportSettingMap.(map[string]interface{})
			if freq, ok := reportSettingMap["Frequency"]; ok {
				if freq.(string) == "None" {
					// Not report anymore
					GinRespond(c, http.StatusOK, SUCCESS, "")
					return
				}
			}
		} else {
			GinRespond(c, http.StatusOK, SUCCESS, "")
			return
		}
		var currReportSetting ReportDataSetting
		err = mapstructure.Decode(reportSettingMap, &currReportSetting)

		doc, err := h.apiContext.Store.Doc("users_meta/" + uid).Get(ctx)
		if err != nil {
			log.Printf("[ERROR] Can not get user meta with user id: %s, %v\n", uid, err)
			GinRespond(c, http.StatusOK, INVALID_PARAMS, "Invalid user id")
			return
		}

		grxusd, err := h.apiContext.Cache.GetGRXUsd()
		xlmusd, err1 := h.apiContext.Cache.GetXLMUsd()
		if err != nil || err1 != nil {
			prices, err := h.apiContext.Store.Doc("price_update/794retePzavE19bTcMaH").Get(ctx)
			if err != nil {
				log.Printf("[ERROR] Can not get user meta with user id: %s, %v\n", uid, err)
				GinRespond(c, http.StatusOK, INVALID_PARAMS, "Invalid user id")
				return
			}
			xlmusd = prices.Data()["xlmusd"].(float64)
			grxusd = prices.Data()["grxusd"].(float64)
		}
		timeLocal := time.Now().Unix()
		contents := GenDataReportMail(currReportSetting, doc.Data(), xlmusd, grxusd, timeLocal)

		title := `GRAYLL | Data Summary Report`
		err = mail.SendNoticeMail(userData["Email"].(string), userData["Name"].(string), title, contents)
		if err != nil {
			log.Println("SendNoticeMail error: ", err)
		}
		//push notice
		content := ""
		for i, s := range contents {
			if i == 0 {
				content = s
			} else {
				content = content + ". " + s
			}
		}
		notice := map[string]interface{}{
			"title":  title,
			"body":   content,
			"isRead": false,
			"time":   timeLocal,
		}

		// contents := make([]string, 0)
		// contents = append(contents, []string{timeLocal.Format(`15:04 | 02-01-2006`)
		docRef := h.apiContext.Store.Collection("notices").Doc("general").Collection(uid).NewDoc()
		_, err = docRef.Set(ctx, notice)
		_, err = h.apiContext.Store.Doc("users_meta/"+uid).Update(ctx, []firestore.Update{
			{Path: "UrGeneral", Value: firestore.Increment(1)},
		})

		current, err := TimeIn(time.Now(), currReportSetting.TimeZone)
		scheduleTime := NewDate(current, currReportSetting.TimeHour, currReportSetting.TimeMinute, currReportSetting.TimeZone)
		// calculate time to send report
		switch currReportSetting.Frequency {
		case "Daily":
			scheduleTime = scheduleTime.Add(time.Hour * 24)
			break
		case "Weekly":
			beginWeek, _ := BeginOfNextWeek(current, currReportSetting.TimeZone)
			scheduleTime = NewDate(beginWeek, currReportSetting.TimeHour, currReportSetting.TimeMinute, currReportSetting.TimeZone)
			break
		case "Monthly":
			beginMonth, _ := BeginOfNextMonth(current, currReportSetting.TimeZone)
			scheduleTime = NewDate(beginMonth, currReportSetting.TimeHour, currReportSetting.TimeMinute, currReportSetting.TimeZone)
			break
		}
		data := make(map[string]interface{})
		data["UserId"] = uid
		data["Time"] = scheduleTime.Unix()
		log.Println("schedule time:", scheduleTime.Unix(), time.Unix(scheduleTime.Unix(), 0).Format("2012-11-01T22:08:41+00:00"))
		h.ScheduleTask(setting, scheduleTime.Unix(), h.apiContext.Config.DataReportUrl, h.apiContext.Config.DataReportQueueId)

		GinRespond(c, http.StatusOK, SUCCESS, "")

		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS,
		})

	}
}

// createHTTPTask creates a new task with a HTTP target then adds it to a Queue.
func (h UserHandler) createUpdateTask(message []byte, execTime int64, taskURL, queueId string) (*taskspb.Task, error) {

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	// opt := option.WithCredentialsFile("grayll-grz-arkady-bda9949575fc.json")
	// client, err := cloudtasks.NewClient(ctx, opt)
	// if err != nil {
	// 	return nil, fmt.Errorf("NewClient: %v", err)
	// }

	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", h.apiContext.Config.ProjectId, h.apiContext.Config.LocationId, queueId)
	ts := new(timestamp.Timestamp)
	ts.Seconds = execTime

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        taskURL,
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: h.apiContext.Config.ServiceAccountEmail,
						},
					},
				},
			},
			ScheduleTime: ts,
		},
	}

	// Add a payload message if one is present.
	req.Task.GetHttpRequest().Body = message

	createdTask, err := h.apiContext.CloudTaskClient.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}

// UpdatePositionTask runs every minute to update the
func (h UserHandler) ScheduleTask(data map[string]interface{}, scheduleTime int64, taskURL, queueId string) (*taskspb.Task, error) {
	json, _ := json.Marshal(data)
	task, err := h.createUpdateTask(json, scheduleTime, taskURL, queueId)
	if err != nil {
		log.Println("createHTTPTask error:", err)
	} else {
		log.Println("createHTTPTask schedule time:", task.GetScheduleTime())
		log.Println("createHTTPTask task name:", task.GetName())
	}
	return task, err

}

// createHTTPTask creates a new task with a HTTP target then adds it to a Queue.
func createHTTPTask(projectID, locationID, queueID, url, serviceAccountEmail string, message []byte, execTime int64) (*taskspb.Task, error) {

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %v", err)
	}

	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)
	ts := new(timestamp.Timestamp)
	ts.Seconds = execTime

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        url,
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: serviceAccountEmail,
						},
					},
				},
			},
			ScheduleTime: ts,
		},
	}

	// Add a payload message if one is present.
	req.Task.GetHttpRequest().Body = message

	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
func GenDataReportMail(reportSetting ReportDataSetting, data map[string]interface{}, xlmusd, grxusd float64, reportTime int64) []string {

	// log.Println(reportSetting)
	// log.Println(data)
	timeLocal, _ := TimeIn(time.Unix(reportTime, 0), reportSetting.TimeZone)
	contents := make([]string, 0)
	contents = append(contents, []string{timeLocal.Format(`15:04 | 02-01-2006`), `Please find your scheduled GRAYLL Data Summary Report below.`}...)

	//If Wallet Balance has been selected
	xlmBalance := GetFloatValue(data["XLM"])
	grxBalance := GetFloatValue(data["GRX"])
	walletBalance := xlmBalance*xlmusd + grxBalance*grxusd
	if reportSetting.WalletBalance {
		contents = append(contents, []string{
			fmt.Sprintf(`Total Wallet Balance: $ %.6f`, walletBalance),
			fmt.Sprintf(`Total XLM Balance: %.6f XLM`, xlmBalance),
			fmt.Sprintf(`Total XLM Balance: $ %.6f`, xlmBalance*xlmusd),
			fmt.Sprintf(`Total GRX Balance: %.6f GRX`, grxBalance),
			fmt.Sprintf(`Total GRX Balance: $ %.6f`, grxBalance*grxusd)}...)
	}

	//If Account Value has been selected by the user
	if reportSetting.AccountValue {
		total_grz_current_position_value := float64(0)
		total_gry1_current_position_value := float64(0)
		total_gry2_current_position_value := float64(0)
		total_gry3_current_position_value := float64(0)
		if value, ok := data["total_grz_current_position_value_$"]; ok {
			total_grz_current_position_value = GetFloatValue(value)
		}
		if value, ok := data["total_gry1_current_position_value_$"]; ok {
			total_gry1_current_position_value = GetFloatValue(value)
		}
		if value, ok := data["total_gry2_current_position_value_$"]; ok {
			total_gry2_current_position_value = GetFloatValue(value)
		}
		if value, ok := data["total_gry3_current_position_value_$"]; ok {
			total_gry3_current_position_value = GetFloatValue(value)
		}

		total := total_gry1_current_position_value + total_gry2_current_position_value + total_gry3_current_position_value + total_grz_current_position_value + walletBalance

		contents = append(contents, fmt.Sprintf(`Total Account Value: $ %.6f`, total))
	}

	//If Account Profit has been selected by the user
	if reportSetting.AccountProfit {

		total_gry1_current_position_ROI := float64(0)
		if value, ok := data["total_gry1_current_position_ROI_$"]; ok {
			total_gry1_current_position_ROI = GetFloatValue(value)
		}
		total_gry1_close_position_ROI := float64(0)
		if value, ok := data["total_gry1_close_position_ROI_$"]; ok {
			total_gry1_close_position_ROI = GetFloatValue(value)
		}

		total_gry2_current_position_ROI := float64(0)
		if value, ok := data["total_gry2_current_position_ROI_$"]; ok {
			total_gry2_current_position_ROI = GetFloatValue(value)
		}
		total_gry2_close_position_ROI := float64(0)
		if value, ok := data["total_gry2_close_position_ROI_$"]; ok {
			total_gry2_close_position_ROI = GetFloatValue(value)
		}

		total_gry3_current_position_ROI := float64(0)
		if value, ok := data["total_gry3_current_position_ROI_$"]; ok {
			total_gry3_current_position_ROI = GetFloatValue(value)
		}
		total_gry3_close_position_ROI := float64(0)
		if value, ok := data["total_gry3_close_position_ROI_$"]; ok {
			total_gry3_close_position_ROI = GetFloatValue(value)
		}

		total_grz_current_position_ROI := float64(0)
		if value, ok := data["total_grz_current_position_ROI_$"]; ok {
			total_grz_current_position_ROI = GetFloatValue(value)
		}
		total_grz_close_positions_ROI := float64(0)
		if value, ok := data["total_grz_close_positions_ROI_$"]; ok {
			total_grz_close_positions_ROI = GetFloatValue(value)
		}
		totalProfit := total_gry3_current_position_ROI + total_gry3_close_position_ROI + total_gry2_current_position_ROI +
			total_gry2_close_position_ROI + total_gry1_current_position_ROI + total_gry1_close_position_ROI + total_grz_current_position_ROI + total_grz_close_positions_ROI
		contents = append(contents, fmt.Sprintf(`Total Account Profits: $ %.6f`, totalProfit))
	}

	//If Account Profit has been selected by the user
	if reportSetting.OpenPosition {
		total_gry1_open_positions := int64(0)
		if value, ok := data["total_gry1_open_positions"]; ok {
			total_gry1_open_positions = GetIntValue(value)
		}
		total_gry2_open_positions := int64(0)
		if value, ok := data["total_gry2_open_positions"]; ok {
			total_gry2_open_positions = GetIntValue(value)
		}
		total_gry3_open_positions := int64(0)
		if value, ok := data["total_gry3_open_positions"]; ok {
			total_gry3_open_positions = GetIntValue(value)
		}
		total_grz_open_positions := int64(0)
		if value, ok := data["total_grz_open_positions"]; ok {
			total_grz_open_positions = GetIntValue(value)
		}
		contents = append(contents,
			[]string{fmt.Sprintf(`Total Open Algo Positions: %d`, total_gry1_open_positions+total_gry2_open_positions+total_gry3_open_positions+total_grz_open_positions),
				fmt.Sprintf(`GRY 1 | Open Algo Positions: %d`, total_gry1_open_positions),
				fmt.Sprintf(`GRY 2 | Open Algo Positions: %d`, total_gry2_open_positions),
				fmt.Sprintf(`GRY 3 | Open Algo Positions: %d`, total_gry3_open_positions),
				fmt.Sprintf(`GRZ | Open Algo Positions: %d`, total_grz_open_positions)}...)
	}
	return contents
}
func (h UserHandler) Warmup() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

func (h UserHandler) CheckPw() gin.HandlerFunc {
	return func(c *gin.Context) {
		pw, ok := c.GetQuery("pw")
		if ok && pw == "--06tVlnkrp-78Uxs0cXU3IEx1B4sY" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "fail",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		}

	}
}

func (h UserHandler) GetDashBoardInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Coins string `json:"coins"`
		}
		mt := new(sync.Mutex)
		db := make(map[string]interface{}, 0)
		err := c.BindJSON(&input)
		if err != nil {
			log.Println("BindJSON err:", err)
			GinRespond(c, http.StatusBadRequest, INVALID_PARAMS, "Can not parse json input")
			return
		}
		//fmt.Printf("got input data: %v\n", input)
		coins := strings.Split(input.Coins, ",")
		wg := new(sync.WaitGroup)
		wg.Add(3)
		// Get dashboard data
		for _, pair := range coins {
			go func(pairstr string) {
				dbData := GetPairDashBoardData(h.apiContext.Store, pairstr, "frame_01m")
				mt.Lock()
				db[pairstr] = dbData
				mt.Unlock()
				wg.Done()
			}(pair)
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"status": "success", "db": db,
		})
	}
}
func (h UserHandler) GetDashBoardInfoGet() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Coins string `json:"coins"`
		}

		input.Coins = c.Param("coins")
		mt := new(sync.Mutex)
		db := make(map[string]interface{}, 0)

		coins := strings.Split(input.Coins, ",")
		wg := new(sync.WaitGroup)
		wg.Add(len(coins))
		// Get dashboard data
		for _, pair := range coins {
			go func(pairstr string) {
				dbData := GetPairDashBoardData(h.apiContext.Store, pairstr, "frame_01m")
				mt.Lock()
				db[pairstr] = dbData
				mt.Unlock()
				wg.Done()
			}(pair)
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"status": "success", "db": db,
		})
	}
}

// func (h UserHandler) GetAlgoRoi() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		//start := time.Now()
// 		//uid := c.GetString(UID)
// 		gry1s := make([]float64, 0)
// 		gry2s := make([]float64, 0)
// 		gry3s := make([]float64, 0)
// 		grzs := make([]float64, 0)
// 		wg := new(sync.WaitGroup)
// 		wg.Add(4)
// 		go func() {
// 			gry1s = h.apiContext.Cache.GetRois("gry1")
// 			wg.Done()
// 		}()
// 		//elapse := time.Since(start).Seconds()
// 		go func() {
// 			gry2s = h.apiContext.Cache.GetRois("gry2")
// 			wg.Done()
// 		}()
// 		go func() {
// 			gry3s = h.apiContext.Cache.GetRois("gry3")
// 			wg.Done()
// 		}()
// 		go func() {
// 			grzs = h.apiContext.Cache.GetRois("grz")
// 			wg.Done()
// 		}()
// 		wg.Wait()
// 		// log.Println("gry1", gry1s, elapse)
// 		// log.Println("gry2s", gry2s)
// 		// log.Println("gry3s", gry3s)
// 		// log.Println("grzs", grzs)
// 		c.JSON(http.StatusOK, gin.H{
// 			"errCode": SUCCESS, "gry1s": gry1s, "gry2s": gry2s, "gry3s": gry3s, "grzs": grzs,
// 		})

// 	}
// }

func (h UserHandler) GetAlgoRoi() gin.HandlerFunc {
	return func(c *gin.Context) {
		//start := time.Now()
		uid := c.GetString(UID)
		gry1s := make([]float64, 0)
		gry2s := make([]float64, 0)
		gry3s := make([]float64, 0)
		grzs := make([]float64, 0)
		wg := new(sync.WaitGroup)
		wg.Add(4)
		go func() {
			gry1s = h.apiContext.Cache.GetRoisNew(uid, "gry1")
			wg.Done()
		}()
		//elapse := time.Since(start).Seconds()
		go func() {
			gry2s = h.apiContext.Cache.GetRoisNew(uid, "gry2")
			wg.Done()
		}()
		go func() {
			gry3s = h.apiContext.Cache.GetRoisNew(uid, "gry3")
			wg.Done()
		}()
		go func() {
			grzs = h.apiContext.Cache.GetRoisNew(uid, "grz")
			wg.Done()
		}()
		wg.Wait()
		// log.Println("gry1", gry1s, elapse)
		// log.Println("gry2s", gry2s)
		// log.Println("gry3s", gry3s)
		// log.Println("grzs", grzs)
		c.JSON(http.StatusOK, gin.H{
			"errCode": SUCCESS, "gry1s": gry1s, "gry2s": gry2s, "gry3s": gry3s, "grzs": grzs,
		})

	}
}
func GetPairDashBoardData(client *firestore.Client, coin, frame string) map[string]interface{} {
	curPrice := GetDashBoardData(client, coin, frame, 0)
	prev1DayPrice := GetDashBoardData(client, coin, frame, 1)
	prev7DayPrice := GetDashBoardData(client, coin, frame, 7)
	db := make(map[string]interface{})

	if coin == "grzusd" {
		db["roi"] = (curPrice.Price - 0.014833677768075555) * 100 / 0.014833677768075555
	} else {
		db["roi"] = (curPrice.Price - 0.01) * 100 / 0.01
	}
	db["price"] = curPrice.Price

	if curPrice.Price > 0 && prev1DayPrice.Price > 0 {
		db["day"] = (curPrice.Price - prev1DayPrice.Price) * 100 / prev1DayPrice.Price
	} else {
		db["day"] = 0.2
	}

	if curPrice.Price > 0 && prev7DayPrice.Price > 0 {
		db["sevendays"] = (curPrice.Price - prev7DayPrice.Price) * 100 / prev7DayPrice.Price
	} else {
		db["sevendays"] = 1.3
	}

	//log.Println("GetPairDashBoardData:", coin, db)
	return db
}

func GetDashBoardData(client *firestore.Client, coin, frame string, days int64) PriceData {
	docPath := fmt.Sprintf("asset_algo_values/%s/%s", coin, frame)
	//fmt.Printf("QueryFrameDataWithTs - coin %s, frame %s docpath %s\n", coin, frame, docPath)
	ctx := context.Background()
	ts := time.Now().Unix() - days*24*60*60
	var it *firestore.DocumentIterator
	if days == 0 {
		it = client.Collection(docPath).OrderBy(UNIX_timestamp, firestore.Desc).Limit(1).Documents(ctx)
	} else {
		it = client.Collection(docPath).Where(UNIX_timestamp, ">=", ts).OrderBy(UNIX_timestamp, firestore.Asc).Limit(1).Documents(ctx)
	}
	p := PriceData{}

	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("err reading db: ", err)
		}

		if doc.Data() != nil {

			p.Ts = doc.Data()[UNIX_timestamp].(int64)
			p.Price = doc.Data()[price].(float64)
			if (ts == 0) || (p.Price == 0) {
				continue
			}
			break
		}
	}
	if p.Price == 0 {
		ts := time.Now().Unix() - (days+1)*24*60*60
		var it *firestore.DocumentIterator
		if days == 0 {
			it = client.Collection(docPath).OrderBy(UNIX_timestamp, firestore.Desc).Limit(1).Documents(ctx)
		} else {
			it = client.Collection(docPath).Where(UNIX_timestamp, ">=", ts).OrderBy(UNIX_timestamp, firestore.Asc).Limit(1).Documents(ctx)
		}

		for {

			doc, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Println("err reading db: ", err)
			}
			if doc.Data() != nil {
				p.Ts = doc.Data()[UNIX_timestamp].(int64)
				p.Price = doc.Data()[price].(float64)
				if (ts == 0) || (p.Price == 0) || math.IsNaN(p.Price) {
					continue
				}
				break
			}
		}
	}
	return p
}

// QueryFrameDataWithTs queries document from subcollection in firestore.
// with limit number of documents and timestamp greater than fromTimeStamp parameter
// grz_price_frames/grzusd/frame_01d/
func QueryFrameDataWithTs(client *firestore.Client, limit int, coin, frame string) []PriceData {
	prices := make([]PriceData, 0)

	var ts int64 = 0
	switch frame {
	case frame05m:
		ts = 1440
		break
	case frame15m:
		ts = 4320
		break
	case frame30m:
		ts = 8640
		break
	case frame01h:
		ts = 17280
		break
	case frame04h:
		ts = 69120
		break
	case frame01d:
		ts = 414720
		break
	default:
		return prices
	}

	//defer client.Close()
	docPath := fmt.Sprintf("asset_algo_values/%s/%s", coin, frame)
	//fmt.Printf("QueryFrameDataWithTs - coin %s, frame %s docpath %s\n", coin, frame, docPath)
	ctx := context.Background()
	ts = time.Now().Unix() - ts*60
	// Change on request start times 1556989920
	// newStartTs := int64(1572519421)
	// if ts < newStartTs {
	// 	ts = newStartTs
	// }
	it := client.Collection(docPath).Where(UNIX_timestamp, ">=", ts).OrderBy(UNIX_timestamp, firestore.Asc).Documents(ctx)

	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("err reading db: ", err)
		}
		if doc == nil {
			log.Println("doc is nil")
			return prices
		}
		ts := doc.Data()[UNIX_timestamp].(int64)

		p := doc.Data()[price].(float64)
		if (ts == 0) || (p == 0) || math.IsNaN(p) {
			continue
		}

		price := PriceData{ts, p}
		prices = append(prices, price)

	}

	return prices
}

// QueryFrameData queries document from subcollection in firestore.
// with limit number of documents and timestamp greater than fromTimeStamp parameter
// pair_frames/gryusd/frame_01d/
func QueryFrameData(client *firestore.Client, limit int, coin, frame string) []PriceData {
	prices := make([]PriceData, 0)
	docPath := fmt.Sprintf("pair_frames/%s/%s", coin, frame)
	//fmt.Printf("QueryFrameData - coin %s, frame %s docpath %s\n", coin, frame, docPath)
	ctx := context.Background()
	it := client.Collection(docPath).OrderBy(UNIX_timestamp, firestore.Desc).Limit(limit).Documents(ctx)

	//defer client.Close()
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("err reading db: ", err)
		}
		if doc == nil {
			log.Println("doc is nil")
			return prices
		}
		ts := doc.Data()[UNIX_timestamp].(int64)

		p := doc.Data()[price].(float64)
		if (ts == 0) || (p == 0) {
			continue
		}

		price := PriceData{ts, p}
		prices = append(prices, price)
		//fmt.Printf("price: %v\n", price)
		//fmt.Printf("Doc Id: %s - timestamp: %d - price: %f\n", doc.Ref.ID, doc.Data()[tsField], doc.Data()[priceField])
	}
	return prices
}
