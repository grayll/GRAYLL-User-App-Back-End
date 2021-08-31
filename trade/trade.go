package main

import (
	"context"
	"flag"

	//"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	//"encoding/json"
	//"fmt"
	"log"
	//"strings"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	"cloud.google.com/go/firestore"

	//"github.com/SherClockHolmes/webpush-go"
	//"net/http"
	//_ "net/http/pprof"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	//"github.com/antigloss/go/logger"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	PRICE_PATH    = "price_update/794retePzavE19bTcMaH"
	ADMIN_SETTING = "admin/8efngc9fgm12nbcxeq"
)

type Config struct {
	ProjectId         string `json:"projectId"`
	DataReportQueueId string `json:"queueId"`

	LocationId string `json:"locationId"`

	DataReportUrl string `json:"dataReportUrl"`

	ServiceAccountEmail string `json:"serviceAccountEmail"`

	IsMainNet         bool    `json:"isMainNet"`
	AssetCode         string  `json:"assetCode"`
	IssuerAddress     string  `json:"issuerAddress"`
	XlmLoanerSeed     string  `json:"xlmLoanerSeed"`
	XlmLoanerAddress  string  `json:"xlmLoanerAddress"`
	RedisHost         string  `json:"redisHost"`
	RedisPort         int     `json:"redisPort"`
	RedisPass         string  `json:"redisPass"`
	HorizonUrl        string  `json:"horizonUrl"`
	Host              string  `json:"host"`
	Numberify         string  `json:"numberify"`
	SuperAdminAddress string  `json:"superAdminAddress"`
	SuperAdminSeed    string  `json:"superAdminSeed"`
	SellingPrice      float64 `json:"sellingPrice"`
	SellingPercent    int     `json:"sellingPercent"`
	NeverBounceApiKey string  `json:"neverBounceApiKey"`

	PauseTimeFrame    int64 `json:"pauseTimeFrame"`    // time frame 15 minutes
	PausePeriod       int64 `json:"pausePeriod"`       // will pause closing 120 minutes
	PauseTimeFrameExt int64 `json:"pauseTimeFrameExt"` // If in 10 minutes from 120 minutes pause increase than PauseIncPerExt
	PausePeriodExt    int64 `json:"pausePeriodExt"`    // Will pause more 30 minutes

	PauseIncPer    float64 `json:"pauseIncPer"`    // increase 15 %
	PauseIncPerExt float64 `json:"pauseIncPerExt"` // Extend 10%
}

func main() {

	var xlmusd, xlmgrx, grxusd float64
	// cache and firestore and webpush
	var store, grzStore *firestore.Client
	var err error

	xlmgrxChan := make(chan float64)

	isMainNet := flag.Bool("mainnet", false, "run on mainnet or testnet")
	isProd := flag.Bool("prod", false, "run on prod or local")
	flag.Parse()
	var config *Config
	configPath := ""
	if *isProd {
		configPath = "/home/huykbc/"
		config = parseConfig(configPath + "config2.json")
	} else {
		configPath = "/home/bc/go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/trade/"
		if *isMainNet {
			config = parseConfig(configPath + "config2.json")
		} else {
			config = parseConfig(configPath + "config.json")
		}
	}

	store, err = GetFsClient(true, configPath, "grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
	if err != nil {
		log.Fatalln("main: GetFsClient error: ", err)
	}
	grzStore, err = GetFsClient(true, configPath, "grayll-grz-arkady-firebase-adminsdk-9q3s2-3fb5715c06.json")
	if err != nil {
		log.Fatalln("main: GetFsClient error: ", err)
	}

	log.Println("config:", config)

	ctx, cancel := context.WithCancel(context.Background())

	stellar.SetupParam(float64(1000), config.IsMainNet, config.HorizonUrl)
	ttl, _ := time.ParseDuration("12h")
	cache, err := api.NewRedisCacheHost(ttl, config.RedisHost, config.RedisPass, config.RedisPort)

	log.SetOutput(&lumberjack.Logger{
		Filename:   configPath + "log/trade-log.txt",
		MaxSize:    10, // megabytes
		MaxBackups: 2,
		MaxAge:     10,   //days
		Compress:   true, // disabled by default
	})

	isPriceValid := true
	HORIZOL_URL := "https://horizon.stellar.org/"
	url := HORIZOL_URL + "trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
	n, d, err := api.GetPrice(url)
	if err == nil {
		xlmusd = n / d
	} else {
		isPriceValid = false
	}

	url = HORIZOL_URL + "trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
	n, d, err = api.GetPrice(url)
	if err == nil {
		xlmgrx = d / n
	} else {
		isPriceValid = false
	}

	client := search.NewClient("BXFJWGU0RM", "ef746e2d654d89f2a32f82fd9ffebf9e")
	algoliaOrderIndex := client.InitIndex("orders-ua")

	// Set price
	if isPriceValid {
		grxusd = xlmgrx * xlmusd
		cache.SetXLMPrice(xlmusd)
		cache.SetGRXPrice(xlmgrx)
		cache.SetGRXUsd(grxusd)

		//ctx := context.Background()

		_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
		if err != nil {
			log.Println("Update grxp error: ", err)
			store, _ = ReconnectFireStore("grayll-app-f3f3f3", 120)
			_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				log.Println("Update grxp error after retry: ", err)
			}
		}

		_, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
		if err != nil {
			log.Println("Update grxp error: ", err)
			//return
		}
	}

	// logger.Init(configPath+"log", // specify the directory to save the logfiles
	// 	40,    // maximum logfiles allowed under the specified log directory
	// 	10,    // number of logfiles to delete when number of logfiles exceeds the configured limit
	// 	10,    // maximum size of a logfile in MB
	// 	false) // whether logs with Trace level are written down

	log.Println("main net: %v", *isMainNet)

	//dexAsset.StreamOrderBook(ctx, "XLM", "", horizonclient.AssetTypeNative, horizonclient.AssetType4)
	//dexAsset.StreamOrderBook(ctx, "USDC", "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN", horizonclient.AssetType4, horizonclient.AssetType4)
	//dexAsset.StreamOrderBook(ctx, "USDC", "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN", horizonclient.AssetType4, horizonclient.AssetType4)
	// dexAsset := DexAsset{assetCode: "GRX", assetIssuer: "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333", store: store}
	// dexAsset.StreamOrderBook(ctx, "XLM", "", horizonclient.AssetTypeNative, horizonclient.AssetType4)
	dexAsset1 := DexAsset{assetCode: "USDC", assetIssuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN", store: store}
	dexAsset1.StreamOrderBook(ctx, "XLM", "", horizonclient.AssetTypeNative, horizonclient.AssetType4)

	// dexAsset1 := DexAsset{assetCode: "XLM", assetIssuer: "", store: store}
	// dexAsset1.StreamOrderBook(ctx, "USDC", "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN", horizonclient.AssetType4, horizonclient.AssetTypeNative)

	tradeRequest := horizonclient.TradeRequest{Cursor: "now"}
	tradeHandler := func(trade horizon.Trade) {
		defer func() {
			if v := recover(); v != nil {
				log.Println("capture a panic:", v)
				log.Println("avoid crashing the program")
			}
		}()

		asset := "GRX"
		if trade.BaseAssetType == "native" && strings.Contains(trade.CounterAssetCode, "GRX") {

			baseAccount := trade.BaseAccount
			counterAcc := trade.CounterAccount

			amount := trade.CounterAmount
			totalxlm := trade.BaseAmount
			time := trade.LedgerCloseTime.Unix()

			xlmgrx = float64(trade.Price.D) / float64(trade.Price.N)
			grxusd = xlmgrx * xlmusd

			// Send to function check volatile
			xlmgrxChan <- xlmgrx

			_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				store, _ = ReconnectFireStore("grayll-app-f3f3f3", 120)
				log.Println("Update grxp error: ", err)
				_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
				if err != nil {
					log.Println("Update grxp error after retry: ", err)
					//return
				}
			}
			_, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				log.Println("Update grxp error: ", err)
				//return
			}

			baseAmount, _ := strconv.ParseFloat(trade.BaseAmount, 64)
			totalusd := xlmusd * baseAmount

			log.Println("counterAcc:", counterAcc)

			offerId := trade.ID[:strings.Index(trade.ID, "-")]
			cache.SetGRXPrice(xlmgrx)
			cache.SetGRXUsd(grxusd)

			uidBaseAcc, err := cache.GetUidFromPublicKey(baseAccount)
			log.Println("uidBaseAcc:", uidBaseAcc)
			if err == nil && uidBaseAcc != "" {

				data := map[string]interface{}{
					"time":     time,
					"type":     "BUY",
					"asset":    asset,
					"amount":   amount,
					"xlmp":     xlmgrx,
					"totalxlm": totalxlm,
					"priceusd": grxusd,
					"totalusd": totalusd,
					"offerId":  offerId,
				}

				docRef := store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
				_, err = docRef.Set(ctx, data)
				if err != nil {
					log.Println("SaveNotice error: ", uidBaseAcc, err)
					store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
					docRef = store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
					_, err = docRef.Set(ctx, data)
					if err != nil {
						log.Println("SaveNotice retried: ", uidBaseAcc, err)
						return
					}
				}

				data["id"] = docRef.ID
				data["uid"] = uidBaseAcc
				algoliaOrderIndex.SaveObject(data)
			}
			uidCounter, err := cache.GetUidFromPublicKey(counterAcc)
			log.Println("uidCounter:", uidCounter)
			if err == nil && uidCounter != "" {
				data := map[string]interface{}{
					"time":     time,
					"type":     "SELL",
					"asset":    asset,
					"amount":   amount,
					"xlmp":     xlmgrx,
					"totalxlm": totalxlm,
					"priceusd": grxusd,
					"totalusd": totalusd,
					"offerId":  offerId,
				}

				docRef := store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
				_, err = docRef.Set(ctx, data)
				if err != nil {
					store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
					log.Println("SaveNotice error: ", uidCounter, err)
					docRef = store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
					_, err = docRef.Set(ctx, data)
					if err != nil {
						log.Println("SaveNotice retried: ", uidCounter, err)
						return
					}
				}
				data["id"] = docRef.ID
				data["uid"] = uidCounter
				algoliaOrderIndex.SaveObject(data)
			}
		}

		if trade.BaseAssetType == "USDC" && strings.Contains(trade.CounterAssetCode, "XLM") {

			baseAccount := trade.BaseAccount
			counterAcc := trade.CounterAccount

			amount := trade.CounterAmount
			totalusdc := trade.BaseAmount
			time := trade.LedgerCloseTime.Unix()

			grxusdc := float64(trade.Price.D) / float64(trade.Price.N)
			//grxusd = xlmgrx * xlmusd

			_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"grxusdc": grxusdc}, firestore.MergeAll)
			if err != nil {
				store, _ = ReconnectFireStore("grayll-app-f3f3f3", 120)
				log.Println("Update grxp error: ", err)
				_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"grxusdc": grxusdc}, firestore.MergeAll)
				if err != nil {
					log.Println("Update grxp error after retry: ", err)
					//return
				}
			}

			log.Println("counterAcc:", counterAcc)

			offerId := trade.ID[:strings.Index(trade.ID, "-")]
			// cache.SetGRXPrice(xlmgrx)
			// cache.SetGRXUsd(grxusd)

			uidBaseAcc, err := cache.GetUidFromPublicKey(baseAccount)
			log.Println("uidBaseAcc:", uidBaseAcc)
			if err == nil && uidBaseAcc != "" {

				data := map[string]interface{}{
					"time":   time,
					"type":   "BUY",
					"asset":  asset,
					"amount": amount,
					//"xlmp":     xlmgrx,
					//"totalxlm": totalxlm,
					"priceusdc": grxusdc,
					"totalusdc": totalusdc,
					"offerId":   offerId,
				}

				docRef := store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
				_, err = docRef.Set(ctx, data)
				if err != nil {
					log.Println("SaveNotice error: ", uidBaseAcc, err)
					store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
					docRef = store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
					_, err = docRef.Set(ctx, data)
					if err != nil {
						log.Println("SaveNotice retried: ", uidBaseAcc, err)
						return
					}
				}

				data["id"] = docRef.ID
				data["uid"] = uidBaseAcc
				algoliaOrderIndex.SaveObject(data)
			}
			uidCounter, err := cache.GetUidFromPublicKey(counterAcc)
			log.Println("uidCounter:", uidCounter)
			if err == nil && uidCounter != "" {
				data := map[string]interface{}{
					"time":   time,
					"type":   "SELL",
					"asset":  asset,
					"amount": amount,
					// "xlmp":     xlmgrx,
					// "totalxlm": totalxlm,
					"priceusdc": grxusdc,
					"totalusdc": totalusdc,
					"offerId":   offerId,
				}

				docRef := store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
				_, err = docRef.Set(ctx, data)
				if err != nil {
					store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
					log.Println("SaveNotice error: ", uidCounter, err)
					docRef = store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
					_, err = docRef.Set(ctx, data)
					if err != nil {
						log.Println("SaveNotice retried: ", uidCounter, err)
						return
					}
				}
				data["id"] = docRef.ID
				data["uid"] = uidCounter
				algoliaOrderIndex.SaveObject(data)
			}
		}

		//get pair xlm/usd
		if trade.BaseAssetType == "native" && trade.CounterAssetCode == "USDC" {

			xlmusd = float64(trade.Price.N) / float64(trade.Price.D)
			xlmusdc := float64(float64(trade.Price.D / trade.Price.N))
			grxusd = xlmgrx * xlmusd
			log.Println("xlmusd:", xlmusd)
			_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "xlmusdc": xlmusdc, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
				log.Println("Update xlmp error: ", err)
				_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "xlmusdc": xlmusdc, "grxusd": grxusd}, firestore.MergeAll)
				if err != nil {
					log.Println("ERROR Update xlmp: ", err)
				}
			}
			_, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "xlmusdc": xlmusdc, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				log.Println("Update xlmp error: ", err)
				//return
			}
			cache.SetXLMPrice(xlmusd)
			cache.SetGRXUsd(grxusd)
		}

	}

	go func() {

		err = stellar.HorizonClient.StreamTrades(ctx, tradeRequest, tradeHandler)
		if err != nil {
			log.Println("StreamTrades", err)

		}
	}()

	// Wait for SIGINT.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	//cursor := tradeRequest.Cursor

	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()

}

func ReconnectFireStore(projectId string, timeout int) (*firestore.Client, error) {
	cnt := 0
	var client *firestore.Client
	var err error
	//ctx := context.Background()
	for {
		cnt++
		time.Sleep(1 * time.Second)
		configPath := "/home/huykbc/"
		client, err = GetFsClient(true, configPath, "grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
		if err == nil {
			break
		}
		if cnt > timeout {
			log.Println("[ERROR] Can not connect to firestore after retry times", cnt, projectId, err)
			break
		}

	}
	return client, err
}

type DexAsset struct {
	assetCode   string
	assetIssuer string
	store       *firestore.Client
}

func (asset DexAsset) NewOrderBookReq(sellingAssetCode, sellingAssetIssuer string, sellingAssetType, buyingAssetType horizonclient.AssetType) horizonclient.OrderBookRequest {

	return horizonclient.OrderBookRequest{
		SellingAssetType:   sellingAssetType,
		SellingAssetCode:   sellingAssetCode,
		SellingAssetIssuer: sellingAssetIssuer,
		//GRX
		BuyingAssetCode:   asset.assetCode,
		BuyingAssetIssuer: asset.assetIssuer,
		BuyingAssetType:   buyingAssetType,
		Limit:             1,
	}
}

func (asset DexAsset) StreamOrderBook(ctx context.Context, sellingAssetCode, sellingAssetIssuer string, sellingAssetType, buyingAssetType horizonclient.AssetType) {
	var askPrice, bidPrice float64

	orderBookReq := asset.NewOrderBookReq(sellingAssetCode, sellingAssetIssuer, sellingAssetType, buyingAssetType)

	prefixPrice := strings.ToLower(sellingAssetCode) + strings.ToLower(asset.assetCode)

	orderBookHandler := func(orderbook horizon.OrderBookSummary) {
		defer func() {
			if v := recover(); v != nil {
				log.Println("capture a panic:", v)
				log.Println("avoid crashing the program")
			}
		}()

		var askf float64 = 0
		for _, ask := range orderbook.Asks {
			askf = float64(ask.PriceR.D) / float64(ask.PriceR.N)
			log.Println("Ask:", sellingAssetCode, ask.Amount, askf)
		}

		var bidf float64 = 0
		for _, bid := range orderbook.Bids {
			bidf = float64(bid.PriceR.D) / float64(bid.PriceR.N)
			log.Println("Bid:", sellingAssetCode, bid.Amount, bidf)
		}
		if askPrice > 0 && bidPrice > 0 && askPrice == askf && bidPrice == bidf {
			return
		}
		askbid := make(map[string]interface{})
		if askf > 0 && bidf > 0 {
			askPrice = askf
			bidPrice = bidf
			askbid = map[string]interface{}{prefixPrice + "_ask": askf, prefixPrice + "_bid": bidf}
		} else if askf > 0 {
			askPrice = askf
			askbid = map[string]interface{}{prefixPrice + "_ask": askf}
		} else if bidf > 0 {
			bidPrice = bidf
			askbid = map[string]interface{}{prefixPrice + "_bid": bidf}
		}
		_, err := asset.store.Doc(PRICE_PATH).Set(ctx, askbid, firestore.MergeAll)
		if err != nil {
			asset.store, _ = ReconnectFireStore("grayll-app-f3f3f3", 120)
			log.Println("Update xlmp error: ", err)
			_, err = asset.store.Doc(PRICE_PATH).Set(ctx, askbid, firestore.MergeAll)
			if err != nil {
				log.Println("Update xlmp error after retry: ", err)
				//return
			}
		}

	}

	go func() {
		err := stellar.HorizonClient.StreamOrderBooks(ctx, orderBookReq, orderBookHandler)
		if err != nil {
			log.Println("StreamOrderBooks", err)

		}
	}()
}
