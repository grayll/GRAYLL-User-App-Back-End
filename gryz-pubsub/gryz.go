package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"context"
	"fmt"
	"io"
	"runtime"

	"cloud.google.com/go/pubsub"
)

type AppContext struct {
}

// // PubSubMessage is the payload of a Pub/Sub event.
// type PubSubMessage struct {
// 	Message struct {
// 		Data []byte `json:"data,omitempty"`
// 		ID   string `json:"id"`
// 	} `json:"message"`
// 	Subscription string `json:"subscription"`
// }

func main() {
	appContext := &AppContext{}
	router := SetupRouter(appContext)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func pullMsgsConcurrenyControl(w io.Writer, projectID, subID string) error {
	// projectID := "my-project-id"
	// subID := "my-sub"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)
	// Must set ReceiveSettings.Synchronous to false (or leave as default) to enable
	// concurrency settings. Otherwise, NumGoroutines will be set to 1.
	sub.ReceiveSettings.Synchronous = false
	// NumGoroutines is the number of goroutines sub.Receive will spawn to pull messages concurrently.
	sub.ReceiveSettings.NumGoroutines = runtime.NumCPU()

	// Receive messages for 10 seconds.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create a channel to handle messages to as they come in.
	cm := make(chan *pubsub.Message)
	// Handle individual messages in a goroutine.
	go func() {
		for {
			select {
			case msg := <-cm:
				fmt.Fprintf(w, "Got message :%q\n", string(msg.Data))
				msg.Ack()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		cm <- msg
	})
	if err != nil {
		return fmt.Errorf("Receive: %v", err)
	}
	close(cm)

	return nil
}
func SetupRouter(appContext *AppContext) *gin.Engine {
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
	// Default(initial) is v1
	v1 := router.Group("/api/v1")
	{
		v1.POST("/price/gry", GryPrice())
		v1.POST("/price/grz", GrzPrice())
	}

	return router
}

func GryPrice() gin.HandlerFunc {
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
func GrzPrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		msg := &pubsub.Message{}
		err := c.BindJSON(msg)
		if err != nil {
			log.Println("error bind json: ", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		log.Printf("grz price %s - time %s", msg.Attributes["price"], msg.Attributes["time"])

		c.Status(http.StatusOK)

	}
}
