package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	//"encoding/json"

	"log"
	"net/http"
	"time"

	//"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	"cloud.google.com/go/firestore"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/timestamp"
	stellar "github.com/huyntsgs/stellar-service"
	"google.golang.org/api/iterator"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
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

		//log.Println("res", res)

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
			txrs, err, txcode := stellar.ParseLoanXDR(input.XDR, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerSeed, h.apiContext.Config.XlmLoanerAddress, float64(2.0001))
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"errCode": txcode,
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
					"errCode": txcode, "xdrResult": txrs.Result,
				})
			}

		} else {
			txrs, err, txcode := stellar.ParseXDR(input.XDR, userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerSeed)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"errCode": txcode,
				})
			} else {
				// tr := &xdr.TransactionResult{}
				// err := tr.Scan(txrs.Result)
				// log.Println("code:", tr.Result.Code.String())
				// rs, _ := tr.Result.MustResults()[0].MustTr().GetManageBuyOfferResult()
				// log.Println("err:", err, rs.Success.Offer.MustOffer().Amount)

				c.JSON(http.StatusOK, gin.H{
					"errCode": txcode, "xdrResult": txrs.Result,
				})
			}
		}
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
					"type":    "general",
					"title":   title,
					"isRead":  false,
					"url":     url,
					"body":    body,
					"time":    time.Now().Unix(),
					"vibrate": []int32{100, 50, 100},
					"icon":    "https://app.grayll.io/favicon.ico",
					"data": map[string]interface{}{
						"url": h.apiContext.Config.Host + "/notifications/overview",
					},
				}

				//ctx := context.Background()
				log.Println("Start push notice")
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
				ctx := context.Background()
				docRef := h.apiContext.Store.Collection("notices").Doc("wallet").Collection(uid).NewDoc()
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
				go func() {
					createLoanReminder(uid, orderId+1, activatedAt)
				}()
				go func() {
					mail.SaveLoanPaidInfo(userInfo["Name"].(string), userInfo["LName"].(string), userInfo["Email"].(string), "no", userInfo["CreatedAt"].(int64), orderId)
				}()
			} else {
				content := genReminderContent(orderId)
				title := "GRAYLL | Account Closure Reminder"
				mail.SendLoanReminder(userInfo["Email"].(string), userInfo["Name"].(string), title, h.apiContext.Config.Host, content, false)

				// merge account
				err := stellar.MergeAccount(userInfo["PublicKey"].(string), h.apiContext.Config.XlmLoanerAddress,
					h.apiContext.Config.XlmLoanerSeed, h.apiContext.Config.IssuerAddress)
				if err != nil {
					log.Println("Can not MergeAccount error:", err)
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

				// app notice
				body := ""
				for _, con := range content {
					body = body + con
				}
				// Send app and push notices
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

				//ctx := context.Background()
				log.Println("Start push notice")
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
	//svAccountEmail := "service-622069026410@gcp-sa-cloudtasks.iam.gserviceaccount.com"
	svAccountEmail := "cloud-tasks-grayll-app@grayll-app-f3f3f3.iam.gserviceaccount.com"
	//service-622069026410@gcp-sa-cloudtasks.iam.gserviceaccount.com
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
		fmt.Printf("got input data: %v\n", input)
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

		fmt.Printf("got input data: %v\n", input)
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
				if (ts == 0) || (p.Price == 0) {
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
	log.Println("docPath:", docPath)
	//it := client.Collection(docPath).OrderBy(UNIX_timestamp, firestore.Desc).Limit(limit).Documents(ctx)

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
		//fmt.Printf("Doc Id: %s - timestamp: %d - price: %f\n", doc.Ref.ID, doc.Data()["UNIX_timestamp"], doc.Data()["price"])
	}

	return prices
}

// QueryFrameData queries document from subcollection in firestore.
// with limit number of documents and timestamp greater than fromTimeStamp parameter
// pair_frames/gryusd/frame_01d/
func QueryFrameData(client *firestore.Client, limit int, coin, frame string) []PriceData {
	prices := make([]PriceData, 0)
	docPath := fmt.Sprintf("pair_frames/%s/%s", coin, frame)
	fmt.Printf("QueryFrameData - coin %s, frame %s docpath %s\n", coin, frame, docPath)
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
