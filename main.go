package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/huyntsgs/stellar-service/assets"
	"google.golang.org/api/option"
)

func main() {
	jwt, err := jwttool.NewJwtFromRsaKey("key/key.pem")
	if err != nil {
		log.Fatal("Error loading rsa key")
	}
	var store *firestore.Client
	prod := os.Getenv("PROD")
	var config *api.Config
	if prod == "1" {
		config = parseConfig("config1.json")
	} else {
		config = parseConfig("config.json")
	}
	asset := assets.Asset{Code: config.AssetCode, IssuerAddress: config.IssuerAddress}
	var cloudTaskClient *cloudtasks.Client

	//spew.Dump(config)
	superAdminAddress := os.Getenv("SUPER_ADMIN_ADDRESS")
	superAdminSeed := os.Getenv("SUPER_ADMIN_SEED")
	sellingPrice := os.Getenv("SELLING_PRICE")
	sellingPercent := os.Getenv("SELLING_PERCENT")
	ctx := context.Background()
	if config.IsMainNet {
		config.IsMainNet = true
		store, err = GetFsClient(false)
		if err != nil {
			log.Fatalln("main: GetFsClient error: ", err)
		}
		if superAdminAddress != "" {
			config.SuperAdminAddress = superAdminAddress
		}

		if superAdminSeed != "" {
			config.SuperAdminSeed = superAdminSeed
		}
		sellingPriceF, _ := strconv.ParseFloat(sellingPrice, 64)
		if sellingPriceF > 0 {
			config.SellingPrice = sellingPriceF
		}
		sellingPercentF, _ := strconv.Atoi(sellingPercent)
		if sellingPercentF > 0 && sellingPercentF <= 100 {
			config.SellingPercent = sellingPercentF
		}
		log.Println("ENV:", config.SuperAdminAddress, config.SellingPrice, config.SellingPercent)
		cloudTaskClient, err = cloudtasks.NewClient(ctx)
	} else {
		config.IsMainNet = false
		store, err = GetFsClient(true)
		if err != nil {
			log.Fatalln("main: GetFsClient error: ", err)
		}
		opt1 := option.WithCredentialsFile("./grayll-grz-arkady-528f3c71b2da.json")
		cloudTaskClient, err = cloudtasks.NewClient(ctx, opt1)
	}

	stellar.SetupParams(float64(1000), config.IsMainNet)
	ttl, _ := time.ParseDuration("12h")
	cache := api.NewRedisCache(ttl, config)
	client := search.NewClient("BXFJWGU0RM", "ef746e2d654d89f2a32f82fd9ffebf9e")
	algoliaOrderIndex := client.InitIndex("orders-ua")

	appContext := &api.ApiContext{Store: store, Jwt: jwt, Cache: cache, Config: config, Asset: asset, OrderIndex: algoliaOrderIndex}
	if cloudTaskClient == nil {
		log.Println("cloudTaskClient is nil")
	}
	appContext.CloudTaskClient = cloudTaskClient

	router := SetupRouter(appContext)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
func CheckAllowedCorApi(reqURL string) bool {
	// if strings.Contains(reqURL, "/users/Renew") || strings.Contains(reqURL, "/users/ReportData") || strings.Contains(reqURL, "/warmup") {
	// 	return true
	// }
	log.Println("reqURL:", reqURL)
	if strings.Contains(reqURL, "/users/ReportData") || strings.Contains(reqURL, "/warmup") {
		return true
	}
	return false
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CheckAllowedCorApi(c.Request.URL.String()) {
			log.Println("allow *")
			c.Writer.Header().Set("Access-Control-Allow-Original", "https://app.grayll.io")
		} else {
			log.Println("allow app.grayll.io", c.Request.Host)
			//c.Writer.Header().Set("Access-Control-Allow-Original", "https://app.grayll.io")
			// if strings.Contains(c.Request.Host, "app.grayll.io") {
			// 	c.Writer.Header().Set("Access-Control-Allow-Original", "https://app.grayll.io")
			// } else {
			// 	c.Writer.Header().Set("Access-Control-Allow-Original", "http://127.0.0.1:8080")
			// }
			c.Writer.Header().Set("Access-Control-Allow-Original", "http://127.0.0.1:4200")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
func SetupRouter(appContext *api.ApiContext) *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://app.grayll.io", "http://127.0.0.1:8081", "http://127.0.0.1:4200"},
		AllowMethods:     []string{"POST, GET, OPTIONS, PUT, DELETE"},
		AllowHeaders:     []string{"Authorization", "Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			log.Println("origin", origin)
			return origin == "https://github.com"
		},
		MaxAge: 24 * time.Hour,
	}))
	//router.Use(CORSMiddleware())
	//router.Use(cors.Default())
	router.Use(gin.Recovery())

	userHandler := api.NewUserHandler(appContext)
	phones := api.NewPhoneHandler(appContext)

	// Always has versioning for api
	// Default(initial) is v1
	v1 := router.Group("/api/v1")
	v1.POST("/accounts/register", userHandler.Register())
	v1.GET("/accounts/validatecode", userHandler.ValidateCode())
	v1.POST("/accounts/login", userHandler.Login())
	v1.POST("/accounts/resendemail", userHandler.ResendEmailConfirm())
	v1.POST("/accounts/mailresetpassword", userHandler.SendEmailResetPwd())
	v1.POST("/accounts/resetpassword", userHandler.ResetPassword())

	v1.GET("/accounts/federation", userHandler.Federation())
	v1.POST("/accounts/xlmLoanReminder", userHandler.XlmLoanReminder())

	v1.GET("/warmup", userHandler.Warmup())
	v1.POST("/reportData", userHandler.ReportData())

	v1.GET("/checkpw", userHandler.CheckPw())

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
		//v1.POST("/users/saveUserData", userHandler.SaveUserData())
		v1.POST("/users/saveUserMetaData", userHandler.SaveUserMetaData())
		v1.POST("/users/saveEnSecretKeyData", userHandler.SaveEnSecretKeyData())

		v1.POST("/users/getUserInfo", userHandler.GetUserInfo())
		v1.POST("/users/GetFramesData", userHandler.GetFramesData())
		v1.GET("/users/GetFramesDataGet/:limit/:coins/:frame", userHandler.GetFramesDataGet())

		v1.POST("/users/GetDashBoardInfo", userHandler.GetDashBoardInfo())
		v1.GET("/users/GetDashBoardInfoGet/:coins", userHandler.GetDashBoardInfoGet())
		v1.POST("/users/Renew", userHandler.Renew())

		v1.POST("/users/saveReportSetting", userHandler.SaveReportSetting())

		v1.POST("/users/MakeTransaction", userHandler.MakeTransaction())

		v1.POST("/phones/sendcode", phones.SendSms())
		v1.POST("/phones/verifycode", phones.VerifyCode())
		v1.POST("/users/ChangePassword", userHandler.ChangePassword())

	}
	return router
}
