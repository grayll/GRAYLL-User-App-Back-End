package main

import (
	"log"
	"os"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"cloud.google.com/go/firestore"
	"github.com/davecgh/go-spew/spew"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/huyntsgs/stellar-service/assets"
)

func main() {
	jwt, err := jwttool.NewJwtFromRsaKey("key/key.pem")
	if err != nil {
		log.Fatal("Error loading rsa key")
	}
	var store *firestore.Client

	config := parseConfig("config.json")
	asset := assets.Asset{Code: config.AssetCode, IssuerAddress: config.IssuerAddress}

	spew.Dump(config)

	if config.IsMainNet {
		config.IsMainNet = true
		store, err = GetFsClient(false)
		if err != nil {
			log.Fatalln("main: GetFsClient error: ", err)
		}
	} else {
		config.IsMainNet = false
		store, err = GetFsClient(true)
		if err != nil {
			log.Fatalln("main: GetFsClient error: ", err)
		}
	}
	stellar.SetupParams(float64(1000), config.IsMainNet)
	ttl, _ := time.ParseDuration("12h")
	cache := api.NewRedisCache(ttl, config)

	appContext := &api.ApiContext{Store: store, Jwt: jwt, Cache: cache, Config: config, Asset: asset}

	router := SetupRouter(appContext)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func SetupRouter(appContext *api.ApiContext) *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "app.grayll.io"},
		AllowMethods:     []string{"POST, GET, OPTIONS, PUT, DELETE"},
		AllowHeaders:     []string{"Authorization", "Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "https://github.com"
		// },
		MaxAge: 12 * time.Hour,
	}))
	//router.Use(cors.Default())
	//router.Use(gin.Recovery())

	//productHandler := api.NewProductHandle(store)
	userHandler := api.NewUserHandler(appContext)
	phones := api.NewPhoneHandler(appContext)

	// Always has versioning for api
	// Default(initial) is v1
	v1 := router.Group("/api/v1")

	// v1.POST("/users/register", userHandler.Register())
	// v1.GET("/users/validatecode", userHandler.ValidateCode())
	// v1.POST("/users/login", userHandler.Login())
	// v1.POST("/users/resendemail", userHandler.ResendEmailConfirm())
	// v1.POST("/users/mailresetpassword", userHandler.SendEmailResetPwd())
	// v1.POST("/users/resetpassword", userHandler.ResetPassword())

	v1.POST("/accounts/register", userHandler.Register())
	v1.GET("/accounts/validatecode", userHandler.ValidateCode())
	v1.POST("/accounts/login", userHandler.Login())
	v1.POST("/accounts/resendemail", userHandler.ResendEmailConfirm())
	v1.POST("/accounts/mailresetpassword", userHandler.SendEmailResetPwd())
	v1.POST("/accounts/resetpassword", userHandler.ResetPassword())

	v1.GET("/accounts/federation", userHandler.Federation())

	// apis needs to authenticate
	v1.Use(api.Authorize(appContext.Jwt))
	{
		v1.POST("/users/updatetfa", userHandler.UpdateTfa())
		v1.POST("/users/setuptfa", userHandler.SetupTfa())
		v1.POST("/users/verifytoken", userHandler.VerifyToken())
		v1.POST("/users/updatesetting", userHandler.UpdateSetting())
		v1.POST("/users/changeemail", userHandler.ChangeEmail())
		v1.POST("/users/updateprofile", userHandler.UpdateProfile())
		v1.POST("/users/editfederation", userHandler.EditFederation())
		v1.POST("/users/validateaccount", userHandler.ValidateAccount())
		v1.POST("/users/savesubcriber", userHandler.SaveSubcriber())
		v1.POST("/users/txverify", userHandler.TxVerify())
		v1.POST("/users/notices", userHandler.GetNotices())
		v1.POST("/users/updateReadNotices", userHandler.UpdateReadNotices())
		v1.POST("/users/getFieldInfo", userHandler.GetFieldInfo())
		v1.POST("/users/verifyRevealSecretToken", userHandler.VerifyRevealSecretToken())
		v1.POST("/users/sendRevealSecretToken", userHandler.SendRevealSecretToken())
		v1.POST("/users/validatePhone", userHandler.ValidatePhone())

		v1.POST("/phones/sendcode", phones.SendSms())
		v1.POST("/phones/verifycode", phones.VerifyCode())

	}
	return router
}
