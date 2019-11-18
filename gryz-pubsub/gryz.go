package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

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
