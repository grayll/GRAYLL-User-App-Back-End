package main

import (
	"context"
	"fmt"

	"log"
	"os"
	"strconv"

	//"strings"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"

	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/huyntsgs/cors"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/huyntsgs/stellar-service/assets"
	"google.golang.org/api/option"

	libredis "github.com/go-redis/redis/v7"
	ccm "github.com/orcaman/concurrent-map"
	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func main() {
	jwt, err := jwttool.NewJwtFromRsaKey("key/key.pem")
	if err != nil {
		log.Fatal("Error loading rsa key")
	}
	jwtAdmin, err := jwttool.NewJwtFromRsaKey("key/keyAdmin.pem")
	if err != nil {
		log.Fatal("Error loading rsa key")
	}
	var store *firestore.Client
	srv := os.Getenv("SERVER")
	var config *api.Config
	log.Println(srv)
	if srv == "prod" {
		config = parseConfig("config1.json")
	} else if srv == "dev" {
		config = parseConfig("config1-dev.json")
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
		log.Println("IsMainNet")
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
		log.Println("here")
		config.IsMainNet = false
		store, err = GetFsClient(true)
		if err != nil {
			log.Fatalln("main: GetFsClient error: ", err)
		}
		opt1 := option.WithCredentialsFile("./grayll-grz-arkady-528f3c71b2da.json")
		cloudTaskClient, err = cloudtasks.NewClient(ctx, opt1)
	}

	stellar.SetupParam(float64(1000), config.IsMainNet, config.HorizonUrl)

	// connect redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost != "" {
		config.RedisHost = redisHost
	}
	ttl, _ := time.ParseDuration("12h")
	cache, err := api.NewRedisCache(ttl, config)
	if err != nil {
		log.Fatalln("ERROR - main - unable connect to redis", err)
	} else {
		log.Println("main- connected to redis", config.RedisHost)
	}
	client := search.NewClient("BXFJWGU0RM", "ef746e2d654d89f2a32f82fd9ffebf9e")
	algoliaOrderIndex := client.InitIndex("orders-ua")

	appContext := &api.ApiContext{Store: store, Jwt: jwt, JwtAdmin: jwtAdmin, Cache: cache, Config: config, Asset: asset, OrderIndex: algoliaOrderIndex}
	if cloudTaskClient == nil {
		log.Println("cloudTaskClient is nil")
	}
	appContext.CloudTaskClient = cloudTaskClient

	// Concurrent map for blocking IP
	ipmap := ccm.New()
	GetBlockedIPs(appContext, ipmap)
	appContext.BlockIPs = ipmap

	router := SetupRouter(appContext, srv)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
func GetBlockedIPs(appContext *api.ApiContext, ipmap ccm.ConcurrentMap) {
	doc, err := appContext.Store.Doc("blocked_ips/ips").Get(context.Background())
	if err != nil {
		log.Println("ERROR unable get blocked ips")
	}

	list := doc.Data()["arrs"].([]interface{})
	for _, ip := range list {
		ipmap.Set(ip.(string), "")
	}
}
func SetupRouter(appContext *api.ApiContext, srv string) *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	rateFormat := "20-M"
	rateFormatEnv := os.Getenv("RATE_LIMIT")
	if rateFormatEnv != "" {
		rateFormat = rateFormatEnv
	}
	rate, err := limiter.NewRateFromFormatted(rateFormat)
	if err != nil {
		log.Fatal(err)
	}
	client := libredis.NewClient(&libredis.Options{
		Addr:     fmt.Sprintf("%s:%d", appContext.Config.RedisHost, appContext.Config.RedisPort),
		Password: appContext.Config.RedisPass})

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "app_limit",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}
	// Create a new middleware with the limiter instance.
	middleware := mgin.NewMiddleware(limiter.New(store, rate))

	router := gin.New()
	//router.Use(gin.Logger())
	if srv == "prod" {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"https://app.grayll.io", "https://admin.grayll.io", "https://grayll-app-test.web.app"},
			AllowMethods:     []string{"POST, GET, OPTIONS, PUT, DELETE"},
			AllowHeaders:     []string{"Authorization", "Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			AllowURLKeywords: []string{"warmup", "federation", "reportData"},
			MaxAge:           24 * time.Hour,
		}))
	} else {
		router.Use(cors.New(cors.Config{
			//AllowOrigins: []string{"https://app.grayll.io"},
			AllowOrigins:     []string{"http://127.0.0.1:4200", "http://127.0.0.1:8081", "https://admin.grayll.io", "https://grayll-app-test.web.app"},
			AllowMethods:     []string{"POST, GET, OPTIONS, PUT, DELETE"},
			AllowHeaders:     []string{"Authorization", "Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			AllowURLKeywords: []string{"warmup", "federation", "reportData"},
			MaxAge:           24 * time.Hour,
		}))
	}

	router.Use(gin.Recovery())
	router.Use(middleware)

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s | %s | %s | %d | %s  | %s | %s | %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Keys["Uid"],
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	}))

	userHandler := api.NewUserHandler(appContext)
	phones := api.NewPhoneHandler(appContext)

	// Always has versioning for api
	// Default(initial) is v1
	v1 := router.Group("/api/v1")
	v1.POST("/accounts/register", userHandler.Register())
	v1.GET("/accounts/validatecode", userHandler.ValidateCode())
	v1.POST("/accounts/login", userHandler.Login())
	v1.POST("/accounts/loginadmin", userHandler.LoginAdmin())
	v1.POST("/accounts/resendemail", userHandler.ResendEmailConfirm())
	v1.POST("/accounts/mailresetpassword", userHandler.SendEmailResetPwd())
	v1.POST("/accounts/resetpassword", userHandler.ResetPassword())

	v1.GET("/accounts/federation", userHandler.Federation())
	v1.POST("/accounts/xlmLoanReminder", userHandler.XlmLoanReminder())

	v1.GET("/warmup", userHandler.Warmup())
	v1.POST("/reportData", userHandler.ReportData())

	v1.GET("/checkpw", userHandler.CheckPw())
	v1.GET("/verifyemail/:email", userHandler.VerifyEmail())
	v1.POST("/verifyrecapchatoken/:email/:action", userHandler.VerifyRecapchaToken())

	v1admin := router.Group("/api/admin/v1")
	v1admin.POST("/accounts/loginadmin", userHandler.LoginAdmin())

	v1admin.Use(api.Authorize(appContext.Jwt))
	{
		v1admin.GET("/users/getusersmeta/:cursor", userHandler.GetUsersMeta())
		v1admin.POST("/users/setstatus", userHandler.SetStatus())
		v1admin.GET("/users/getuserdata/:searchStr", userHandler.GetUserData())
		v1admin.POST("/users/firebaseauth", userHandler.AdminFirebaseAuth())

		v1admin.POST("/users/verifykycdoc", userHandler.VerifyKycDoc())
		v1admin.POST("/users/finalauditkyc", userHandler.FinalAuditKYC())
	}

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
		v1.POST("/users/txbuygrx", userHandler.TxBuyGrx())
		v1.POST("/users/notices", userHandler.GetNotices())
		v1.POST("/users/updateReadNotices", userHandler.UpdateReadNotices())
		v1.POST("/users/updateAllAsRead/:noticeType", userHandler.UpdateAllAsRead())
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
		v1.GET("/users/getalgoroi", userHandler.GetAlgoRoi())

		v1.POST("/users/saveReportSetting", userHandler.SaveReportSetting())

		v1.POST("/users/MakeTransaction", userHandler.MakeTransaction())
		v1.POST("/users/PayLoan", userHandler.PayLoan())

		v1.POST("/phones/sendcode", phones.SendSms())
		v1.POST("/phones/verifycode", phones.VerifyCode())
		v1.POST("/users/ChangePassword", userHandler.ChangePassword())

		v1.POST("/users/invite", userHandler.Invite())
		v1.POST("/users/reinvite/:docId", userHandler.ReInvite())
		v1.POST("/users/delinvite/:docId", userHandler.DelInvite())
		v1.POST("/users/removeReferral/:referralId", userHandler.RemveReferral())
		v1.POST("/users/removeReferer/:refererId", userHandler.RemveReferer())
		v1.POST("/users/editreferral", userHandler.EditReferral())

		v1.POST("/users/reportclosing", userHandler.ReportClosing())
		v1.POST("/users/updateHomeDomain", userHandler.UpdateHomeDomain())

		v1.POST("/users/updatekyc", userHandler.UpdateKyc())
		v1.POST("/users/updatekyccom", userHandler.UpdateKycCom())
		v1.POST("/users/updatekycdoc", userHandler.UpdateKycDoc())
		v1.POST("/users/firebaseauth", userHandler.FirebaseAuth())

	}
	return router
}
