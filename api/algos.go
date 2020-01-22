package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	//"encoding/json"

	"log"
	"net/http"
	"time"

	//"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/timestamp"
	stellar "github.com/huyntsgs/stellar-service"
	"google.golang.org/api/iterator"

	//"github.com/go-redis/redis"

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

// Creates new UserHandler.
// UserHandler accepts interface UserStore.
// Any data store implements UserStore could be the input of the handle.
// func NewUserHandler(apiContext *ApiContext) UserHandler {
// 	return UserHandler{apiContext: apiContext}
// }

func (h UserHandler) GetFramesData() gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Limit int    `json:"limit"`
			Coins string `json:"coins"`
			Frame string `json:"frame"`
		}
		mt := new(sync.Mutex)
		res := make(map[string][]*PriceData, 0)

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
		for _, pair := range coins {
			if input.Limit == 1 {
				prices := QueryFrameData(h.apiContext.Store, input.Limit, pair, input.Frame)
				res[pair] = prices
			} else {
				go func(pairstr string) {
					prices := QueryFrameDataWithTs(h.apiContext.Store, input.Limit, pairstr, input.Frame)
					mt.Lock()
					res[pairstr] = prices
					mt.Unlock()
					wg.Done()
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
				log.Println("set isloan:")
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

func (h UserHandler) HandleXlmLoanReminder() gin.HandlerFunc {
	return func(c *gin.Context) {
		//X-CloudTasks-TaskETA X-CloudTasks-TaskName X-CloudTasks-TaskRetryCount
		taskName := c.GetHeader("X-Appengine-Taskname")
		if len(taskName) == 0 {
			// You may use the presence of the X-Appengine-Taskname header to validate
			// the request comes from Cloud Tasks.
			log.Println("Invalid Task: No X-Appengine-Taskname request header found")
			//http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
			return
		}

		// Pull useful headers from Task request.
		queueName := c.GetHeader("X-Appengine-Queuename")

		// Extract the request body for further task details.
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("ReadAll: %v", err)
			//http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// Log & output details of the task.
		output := fmt.Sprintf("Completed task: task queue(%s), task name(%s), payload(%s)",
			queueName,
			taskName,
			string(body),
		)
		log.Println(output)
	}
}

// createHTTPTask creates a new task with a HTTP target then adds it to a Queue.
func createHTTPTask(projectID, locationID, queueID, url, message string) (*taskspb.Task, error) {

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %v", err)
	}

	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)
	execTime := new(timestamp.Timestamp)

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
				},
			},
			ScheduleTime: execTime,
		},
	}

	// Add a payload message if one is present.
	req.Task.GetHttpRequest().Body = []byte(message)

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

func GetPairDashBoardData(client *firestore.Client, coin, frame string) map[string]interface{} {
	curPrice := GetDashBoardData(client, coin, frame, 0)
	prev1DayPrice := GetDashBoardData(client, coin, frame, 1)
	prev7DayPrice := GetDashBoardData(client, coin, frame, 7)
	db := make(map[string]interface{})
	if curPrice.Price > 0 && prev1DayPrice.Price > 0 && prev7DayPrice.Price > 0 {
		db["day"] = (curPrice.Price - prev1DayPrice.Price) * 100 / prev1DayPrice.Price
		db["sevendays"] = (curPrice.Price - prev7DayPrice.Price) * 100 / prev7DayPrice.Price
		if coin == "grzusd" {
			db["roi"] = (curPrice.Price - 0.014833677768075555) * 100 / 0.014833677768075555
		} else {
			db["roi"] = (curPrice.Price - 0.01) * 100 / 0.01
		}
		db["price"] = curPrice.Price
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
		if doc == nil {
			log.Println("doc is nil")
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
	return p
}

// QueryFrameDataWithTs queries document from subcollection in firestore.
// with limit number of documents and timestamp greater than fromTimeStamp parameter
// grz_price_frames/grzusd/frame_01d/
func QueryFrameDataWithTs(client *firestore.Client, limit int, coin, frame string) []*PriceData {
	prices := make([]*PriceData, 0)

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
	// Change on request start times
	newStartTs := int64(1572519421)
	if ts < newStartTs {
		ts = newStartTs
	}
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
		if (ts == 0) || (p == 0) {
			continue
		}

		price := &PriceData{ts, p}
		prices = append(prices, price)
		//fmt.Printf("price: %v\n", price)
		//fmt.Printf("Doc Id: %s - timestamp: %d - price: %f\n", doc.Ref.ID, doc.Data()[tsField], doc.Data()[priceField])
	}
	return prices
}

// QueryFrameData queries document from subcollection in firestore.
// with limit number of documents and timestamp greater than fromTimeStamp parameter
// pair_frames/gryusd/frame_01d/
func QueryFrameData(client *firestore.Client, limit int, coin, frame string) []*PriceData {
	prices := make([]*PriceData, 0)
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

		price := &PriceData{ts, p}
		prices = append(prices, price)
		//fmt.Printf("price: %v\n", price)
		//fmt.Printf("Doc Id: %s - timestamp: %d - price: %f\n", doc.Ref.ID, doc.Data()[tsField], doc.Data()[priceField])
	}
	return prices
}
