package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"context"
	"fmt"

	//"io"
	//"errors"
	"cloud.google.com/go/firestore"
	// firebase "firebase.google.com/go"
	// "google.golang.org/api/iterator"
	"runtime"

	"github.com/antigloss/go/logger"
	"google.golang.org/api/option"

	"cloud.google.com/go/pubsub"
)

type AppContext struct {
}
type PriceH struct {
	Price  float64 `json:"price"`
	UnixTs int64   `json:"unixts"`
	GRType string  `json:"grtype"`
	Frame  string  `json:"frame"`
}

var client *firestore.Client
var pubsubClient *pubsub.Client

// For pull subscription
func main() {

	ctx, cancel := context.WithCancel(context.Background())
	var err error
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "grayll-app-f3f3f3"
		log.Println("GOOGLE_CLOUD_PROJECT is not set. Will use default: grayll-app-f3f3f3")
	}
	isProd := flag.Bool("prod", false, "run on prod or local")
	flag.Parse()
	configPath := ""
	if *isProd {
		configPath = "/home/huykbc/"
	} else {
		configPath = "/home/bc/go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/pullsubs/"
	}
	logger.Init(configPath+"log", // specify the directory to save the logfiles
		40,    // maximum logfiles allowed under the specified log directory
		10,    // number of logfiles to delete when number of logfiles exceeds the configured limit
		10,    // maximum size of a logfile in MB
		false) // whether logs with Trace level are written down
	if client == nil {
		if *isProd {
			client, err = firestore.NewClient(ctx, projectID)
			if err != nil {
				logger.Error("firestore.NewClient error: ", err)
				return
			}
		} else {
			opt := option.WithCredentialsFile("./grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
			client, err = firestore.NewClient(ctx, projectID, opt)
			if err != nil {
				logger.Error("firestore.NewClient error: ", err)
				return
			}
		}
	}
	if pubsubClient == nil {
		if *isProd {
			pubsubClient, err = pubsub.NewClient(ctx, projectID)
			if err != nil {
				logger.Error("pubsub.NewClient error:", err)
				return
			}
		} else {
			opt := option.WithCredentialsFile("./grayll-app-f3f3f3-0dbded8d153a.json")
			pubsubClient, err = pubsub.NewClient(ctx, projectID, opt)
			if err != nil {
				logger.Error("pubsub.NewClient error:", err)
				return
			}
		}
	}

	//TEST
	openData := map[string]interface{}{
		"user_id":                     "2Ar1licxmSTJgzqHYqNk_bNaHgoh8ZR4zj8nJcYaB18",
		"grayll_transaction_id":       8,
		"open_stellar_transaction_id": 525867,
		"algorithm_type":              "GRZ",
		"open_position_timestamp":     1583155967,
		"open_position_fee_$":         0.3,
		"open_position_fee_GRX":       13.8057,
		"open_position_value_$":       99.7,
		"open_position_total_GRX":     4601.9,
		"open_position_value_GRZ":     3120.9864752,
		"open_position_value_GRX":     4588.0943,
		"duration":                    0,
		"status":                      "OPEN",
		"action":                      "open",
		"current_value_GRX":           0.021,
		"current_value_GRZ":           0.03207,
		"current_position_value_$":    99.7,
		"current_position_value_GRX":  4588.0943,
		"current_position_ROI_$":      0,
		"current_position_ROI_%":      0,
	}
	openDataBytes, err := json.Marshal(openData)
	PublishMessage("algo_position_data", openDataBytes)

	// close_stellar_transaction_id
	// close_position_timestamp
	// close_position_total_$
	// close_position_total_GRX
	// close_position_value_GRZ

	closeData := map[string]interface{}{
		"user_id":                      "2Ar1licxmSTJgzqHYqNk_bNaHgoh8ZR4zj8nJcYaB18",
		"grayll_transaction_id":        6,
		"close_stellar_transaction_id": 525867,
		"algorithm_type":               "GRZ",
		"close_position_timestamp":     1583165967,
		"close_position_total_$":       0.3,

		"close_position_total_GRX": 7601.9,
		"close_position_value_GRZ": 5120.9864752,

		"duration":                  344545,
		"status":                    "CLOSE",
		"action":                    "close",
		"close_position_fee_$":      0.3,
		"close_performance_fee_$":   18,
		"close_performance_fee_GRX": 78,
		"close_position_value_$":    199.7,
		"close_position_value_GRX":  7522,
		"close_position_ROI_$":      99,
		"close_position_ROI_%":      80,
	}
	openDataBytes, err = json.Marshal(closeData)
	PublishMessage("algo_position_data", openDataBytes)
	//END TEST

	// go func() {
	// 	pullMsgsConcurrenyControl(ctx, projectID, "pull_mvp_price")
	// }()

	go func() {
		pullMsgsConcurrenyControl(ctx, projectID, "algo_position_data")
	}()

	// Wait for SIGINT.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()

}

func pullMsgsConcurrenyControl(ctx context.Context, projectID, subID string) error {

	sub := pubsubClient.Subscription(subID)
	// Must set ReceiveSettings.Synchronous to false (or leave as default) to enable
	// concurrency settings. Otherwise, NumGoroutines will be set to 1.
	sub.ReceiveSettings.Synchronous = false
	// NumGoroutines is the number of goroutines sub.Receive will spawn to pull messages concurrently.
	sub.ReceiveSettings.NumGoroutines = runtime.NumCPU()

	// Receive messages for 10 seconds.
	//ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a channel to handle messages to as they come in.
	cm := make(chan *pubsub.Message)
	// Handle individual messages in a goroutine.
	go func() {
		for {
			select {
			case msg := <-cm:
				logger.Info("Received data:", string(msg.Data))
				// Store gry,grz data
				if subID == "pull_mvp_price" {
					go func() {
						ProcessMvpData(ctx, msg)
					}()
				} else if subID == "algo_position_data" {
					go func() {
						ProcessAlgoData(ctx, msg)
					}()
				}

				msg.Ack()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Receive blocks until the context is cancelled or an error occurs.
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		cm <- msg
	})
	if err != nil {
		return fmt.Errorf("Receive: %v", err)
	}
	close(cm)

	return nil
}
func ProcessAlgoData(ctx context.Context, msg *pubsub.Message) error {

	algoData := make(map[string]interface{})
	err := json.Unmarshal(msg.Data, &algoData)
	if err != nil {
		logger.Error("json.Unmarshal error: ", err)
		return err
	}
	logger.Info("algoData:", algoData)
	//action: open,close,update
	action := algoData["action"].(string)

	user_id := algoData["user_id"].(string)
	grayll_tx_id := algoData["grayll_transaction_id"].(float64)
	if user_id == "" {
		logger.Error("Invalid user_id empty")
		return errors.New("invalid userid or grayll tx id")
	}
	if action == "open" {
		docPath := fmt.Sprintf("algo_positions/users/%s/%d", user_id, int64(grayll_tx_id))
		_, err = client.Doc(docPath).Set(ctx, algoData)
		if err != nil {
			logger.Error("Save algo position data error: ", err)
			return err
		}
	} else {
		docPath := fmt.Sprintf("algo_positions/users/%s/%d", user_id, int64(grayll_tx_id))
		_, err = client.Doc(docPath).Set(ctx, algoData, firestore.MergeAll)
		if err != nil {
			logger.Error("Save algo position data error: ", err)
			return err
		}
	}

	return nil
}
func ProcessMvpData(ctx context.Context, msg *pubsub.Message) error {

	priceh := PriceH{}
	err := json.Unmarshal(msg.Data, &priceh)
	if err != nil {
		logger.Error("json.Unmarshal error: ", err)
		return err
	}

	batch := client.Batch()
	docPath := fmt.Sprintf("asset_algo_values/%s/%s", priceh.GRType, priceh.Frame)
	docRef := client.Collection(docPath).NewDoc()
	batch.Create(docRef, map[string]interface{}{
		"price":          priceh.Price,
		"UNIX_timestamp": priceh.UnixTs,
	})
	logger.Info("ADD ASSET-ALGO PRICE: ", priceh.GRType, priceh.Frame, time.Unix(priceh.UnixTs, 0).Format("2006-01-02 15:04:05"), priceh.Price)
	if priceh.Frame == "frame_01m" && (priceh.GRType == "gryusd" || priceh.GRType == "grzusd") {
		logger.Info("PUBLISH PRICE: ", priceh.GRType, priceh.Frame, time.Unix(priceh.UnixTs, 0).Format("2006-01-02 15:04:05"))
		priceStr := priceh.GRType[:3] + "p"
		priceStrNew := priceh.GRType[:3] + "usd"
		docRefPrice := client.Doc("prices/794retePzavE19bTcMaH")
		batch.Set(docRefPrice, map[string]interface{}{
			priceStr:    priceh.Price,
			priceStrNew: priceh.Price,
		}, firestore.MergeAll)
	}
	// _, err = batch.Commit(ctx)
	// if err != nil {
	// 	logger.Error("Batch commit error: ", err)
	// 	return err
	// }
	return nil
}

func PublishMessage(topicID string, msg []byte) error {

	ctx := context.Background()
	t := pubsubClient.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{Data: msg})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		log.Printf("Published error: %v\n", err)
		return fmt.Errorf("Get: %v", err)
	}
	log.Printf("Published message ID: %v\n", id)
	return nil
}

///==================///

func main1() {
	ctx, _ := context.WithCancel(context.Background())
	var err error
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "grayll-app-f3f3f3"
		log.Println("GOOGLE_CLOUD_PROJECT is not set. Will use default: grayll-app-f3f3f3")
	}
	if client == nil {
		//opt := option.WithCredentialsFile("./grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
		client, err = firestore.NewClient(ctx, projectID)
		if err != nil {
			log.Println("firestore.NewClient error: ", err)
			return
		}
	}

	router := SetupRouter(client, ctx)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)

}

func SetupRouter(client *firestore.Client, ctx context.Context) *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "app.grayll.io"},
		AllowMethods:     []string{"POST, OPTIONS, PUT"},
		AllowHeaders:     []string{"Authorization", "Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	//router.Use(cors.Default())
	//router.Use(gin.Recovery())

	// Always has versioning for api
	v1 := router.Group("/api/v1")
	{
		v1.POST("/price/mvp", MvpPrice())
		v1.POST("/postion/open", GrzPrice(ctx))
		v1.POST("/postion/close", GrzPrice(ctx))
	}

	return router
}

func MvpPrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		msg := &pubsub.Message{}
		err := c.BindJSON(msg)
		if err != nil {
			log.Println("error bind json: ", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		log.Printf("gry price %s - time %s", msg.Attributes["price"], msg.Attributes["time"])
		c.Status(http.StatusOK)

	}
}
func GrzPrice(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		priceh := &PriceH{}
		err := c.BindJSON(priceh)
		if err != nil {
			log.Println("error bind json: ", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		batch := client.Batch()
		docPath := fmt.Sprintf("asset_algo_values/%s/%s", priceh.GRType, priceh.Frame)
		docRef := client.Collection(docPath).NewDoc()
		batch.Create(docRef, map[string]interface{}{
			"price":          priceh.Price,
			"UNIX_timestamp": priceh.UnixTs,
		})
		log.Println("ADD ASSET-ALGO PRICE: ", priceh.GRType, priceh.Frame, time.Unix(priceh.UnixTs, 0).Format("2006-01-02 15:04:05"), priceh.Price)
		if priceh.Frame == "frame_01m" && (priceh.GRType == "gryusd" || priceh.GRType == "grzusd") {
			log.Println("PUBLISH PRICE: ", priceh.GRType, priceh.Frame, time.Unix(priceh.UnixTs, 0).Format("2006-01-02 15:04:05"))
			priceStr := priceh.GRType[:3] + "p"
			priceStrNew := priceh.GRType[:3] + "usd"
			docRefPrice := client.Doc("prices/794retePzavE19bTcMaH")
			batch.Set(docRefPrice, map[string]interface{}{
				priceStr:    priceh.Price,
				priceStrNew: priceh.Price,
			}, firestore.MergeAll)
		}
		_, err = batch.Commit(ctx)
		if err != nil {
			log.Println("Batch commit error: ", err)

		}

		c.Status(http.StatusOK)

	}
}
