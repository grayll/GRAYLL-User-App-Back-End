package main

import (
	"context"
	"flag"
	"math"
	"os"
	"os/signal"

	//"strconv"
	//"strings"
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

	//"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/antigloss/go/logger"
	stellar "github.com/huyntsgs/stellar-service"
	//"github.com/stellar/go/clients/horizonclient"
	//"github.com/stellar/go/protocols/horizon"
	//"gopkg.in/natefinch/lumberjack.v2"
)

const (
	PRICE_PATH    = "price_update/794retePzavE19bTcMaH"
	ADMIN_SETTING = "admin/8efngc9fgm12nbcxeq"
)

func main() {

	var xlmusd, xlmgrx, grxusd float64
	// cache and firestore and webpush
	var store, grzStore *firestore.Client
	var err error

	xlmgrxChan := make(chan float64)

	isMainNet := flag.Bool("mainnet", false, "run on mainnet or testnet")
	isProd := flag.Bool("prod", false, "run on prod or local")
	flag.Parse()
	var config *api.Config
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

	// store, err = GetFsClient(true, configPath, "grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
	// if err != nil {
	// 	log.Fatalln("main: GetFsClient error: ", err)
	// }
	store, err = ReconnectFireStore("grayll-app-f3f3f3", 60)
	grzStore, err = GetFsClient(true, configPath, "grayll-grz-arkady-firebase-adminsdk-9q3s2-3fb5715c06.json")
	if err != nil {
		log.Fatalln("main: GetFsClient error: ", err)
	}

	log.Println("config.HorizonUrl:", config.HorizonUrl)

	ctx, cancel := context.WithCancel(context.Background())

	stellar.SetupParam(float64(1000), config.IsMainNet, config.HorizonUrl)
	ttl, _ := time.ParseDuration("12h")
	cache, err := api.NewRedisCache(ttl, config)
	cnt := 0
	if err != nil {
		for {
			cnt++
			time.Sleep(1 * time.Second)
			cache, err = api.NewRedisCache(ttl, config)
			if err == nil {
				break
			}
			if cnt > 120 {
				log.Fatalln("Can not connect to redis", err)
			}

		}
	}
	// log.SetOutput(&lumberjack.Logger{
	// 	Filename:   configPath + "log/trade-log1.txt",
	// 	MaxSize:    10, // megabytes
	// 	MaxBackups: 2,
	// 	MaxAge:     10,   //days
	// 	Compress:   true, // disabled by default
	// })

	isPriceValid := true

	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
	n, d, err := api.GetPrice(url)
	if err == nil {
		xlmusd = n / d
	} else {
		isPriceValid = false
	}

	url = "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
	n, d, err = api.GetPrice(url)
	if err == nil {
		xlmgrx = d / n
	} else {
		isPriceValid = false
	}

	//client := search.NewClient("BXFJWGU0RM", "ef746e2d654d89f2a32f82fd9ffebf9e")
	//algoliaOrderIndex := client.InitIndex("orders-ua")

	// Set price
	if isPriceValid {
		grxusd = xlmgrx * xlmusd
		cache.SetXLMPrice(xlmusd)
		cache.SetGRXPrice(xlmgrx)
		cache.SetGRXUsd(grxusd)

		_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
		if err != nil {
			log.Println("Update grxp error: ", err)
			store, _ = api.ReconnectFireStore("grayll-app-f3f3f3", 120)
			_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
			if err != nil {
				log.Println("Update grxp error after retry: ", err)

				return
			}
		}

		_, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
		if err != nil {
			log.Println("Update grxp error: ", err)
			//return
		}
	}

	logger.Init(configPath+"log", // specify the directory to save the logfiles
		40,    // maximum logfiles allowed under the specified log directory
		10,    // number of logfiles to delete when number of logfiles exceeds the configured limit
		10,    // maximum size of a logfile in MB
		false) // whether logs with Trace level are written down

	log.Println("main net: %v", *isMainNet)

	// orderBookReq := horizonclient.OrderBookRequest{
	// 	SellingAssetType:   horizonclient.AssetTypeNative,
	// 	SellingAssetCode:   "XLM",
	// 	SellingAssetIssuer: "",
	// 	BuyingAssetCode:    "GRX",
	// 	BuyingAssetIssuer:  "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333",
	// 	BuyingAssetType:    horizonclient.AssetType4,
	// 	Limit:              1,
	// }

	// var askPrice, bidPrice float64

	// orderBookHandler := func(orderbook horizon.OrderBookSummary) {
	// 	defer func() {
	// 		if v := recover(); v != nil {
	// 			log.Println("capture a panic:", v)
	// 			log.Println("avoid crashing the program")
	// 		}
	// 	}()

	// 	log.Println("Asks:")
	// 	var askf float64 = 0
	// 	for _, ask := range orderbook.Asks {
	// 		askf = float64(ask.PriceR.D) / float64(ask.PriceR.N)
	// 		log.Println(ask.Amount, askf)
	// 	}
	// 	log.Println("End Asks:")

	// 	log.Println("Bids:")
	// 	var bidf float64 = 0
	// 	for _, bid := range orderbook.Bids {
	// 		bidf = float64(bid.PriceR.D) / float64(bid.PriceR.N)
	// 		log.Println(bid.Amount, bidf)
	// 	}
	// 	if askPrice > 0 && bidPrice > 0 && askPrice == askf && bidPrice == bidf {
	// 		return
	// 	}
	// 	askbid := make(map[string]interface{})
	// 	if askf > 0 && bidf > 0 {
	// 		askPrice = askf
	// 		bidPrice = bidf
	// 		askbid = map[string]interface{}{"xlmgrx_ask": askf, "xlmgrx_bid": bidf}
	// 	} else if askf > 0 {
	// 		askPrice = askf
	// 		askbid = map[string]interface{}{"xlmgrx_ask": askf}
	// 	} else if bidf > 0 {
	// 		bidPrice = bidf
	// 		askbid = map[string]interface{}{"xlmgrx_bid": bidf}
	// 	}
	// 	_, err = store.Doc(PRICE_PATH).Set(ctx, askbid, firestore.MergeAll)
	// 	if err != nil {
	// 		store, _ = api.ReconnectFireStore("grayll-app-f3f3f3", 120)
	// 		log.Println("Update xlmp error: ", err)
	// 		_, err = store.Doc(PRICE_PATH).Set(ctx, askbid, firestore.MergeAll)
	// 		if err != nil {
	// 			log.Println("Update xlmp error after retry: ", err)
	// 			//return
	// 		}
	// 	}
	// 	_, err = grzStore.Doc(PRICE_PATH).Set(ctx, askbid, firestore.MergeAll)
	// 	if err != nil {
	// 		log.Println("Update xlmp error: ", err)
	// 		//return
	// 	}
	// 	log.Println("BEnd ids:")
	// }

	// go func() {
	// 	err = stellar.HorizonClient.StreamOrderBooks(ctx, orderBookReq, orderBookHandler)
	// 	if err != nil {
	// 		log.Println("StreamOrderBooks", err)

	// 	}
	// }()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:8091", nil))
	// }()
	go func(ctx context.Context) {
		for {
			time.Sleep(90 * time.Second)
			xlmgrx = xlmgrx + xlmgrx*0.18
			xlmgrxChan <- xlmgrx
		}
	}(ctx)

	// tradeRequest := horizonclient.TradeRequest{Cursor: "now"}

	// tradeHandler := func(trade horizon.Trade) {
	// 	defer func() {
	// 		if v := recover(); v != nil {
	// 			log.Println("capture a panic:", v)
	// 			log.Println("avoid crashing the program")
	// 		}
	// 	}()

	// 	// identify base account - counteraccount is belong to grayll
	// 	// base is buyer, counter is seller
	// 	//get pair grx/xlm

	// 	//asset := "GRX"
	// 	if trade.BaseAssetType == "native" && strings.Contains(trade.CounterAssetCode, "GRX") {

	// 		baseAccount := trade.BaseAccount
	// 		counterAcc := trade.CounterAccount

	// 		// amount := trade.CounterAmount
	// 		// totalxlm := trade.BaseAmount
	// 		// time := trade.LedgerCloseTime.Unix()

	// 		xlmgrx = float64(trade.Price.D) / float64(trade.Price.N)
	// 		grxusd = xlmgrx * xlmusd

	// 		xlmgrxChan <- xlmgrx

	// 		// _, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
	// 		// if err != nil {
	// 		// 	store, _ = api.ReconnectFireStore("grayll-app-f3f3f3", 120)
	// 		// 	log.Println("Update grxp error: ", err)
	// 		// 	_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
	// 		// 	if err != nil {
	// 		// 		log.Println("Update grxp error after retry: ", err)
	// 		// 		//return
	// 		// 	}
	// 		// }
	// 		// _, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmgrx": xlmgrx, "grxusd": grxusd}, firestore.MergeAll)
	// 		// if err != nil {
	// 		// 	log.Println("Update grxp error: ", err)
	// 		// 	//return
	// 		// }

	// 		//priceusd := xlmGrxPrice * xlmUsdPrice

	// 		//baseAmount, _ := strconv.ParseFloat(trade.BaseAmount, 64)
	// 		//totalusd := xlmusd * baseAmount

	// 		// txId := trade.ID
	// 		// offerId := trade.OfferID

	// 		log.Println("counterAcc:", counterAcc)

	// 		//offerId := trade.ID[:strings.Index(trade.ID, "-")]
	// 		cache.SetGRXPrice(xlmgrx)
	// 		cache.SetGRXUsd(grxusd)

	// 		uidBaseAcc, err := cache.GetUidFromPublicKey(baseAccount)
	// 		log.Println("uidBaseAcc:", uidBaseAcc)
	// 		if err == nil && uidBaseAcc != "" {
	// 			// data := map[string]interface{}{
	// 			// 	"time":     time,
	// 			// 	"type":     "BUY",
	// 			// 	"asset":    asset,
	// 			// 	"amount":   amount,
	// 			// 	"xlmp":     xlmgrx,
	// 			// 	"totalxlm": totalxlm,
	// 			// 	"priceusd": grxusd,
	// 			// 	"totalusd": totalusd,
	// 			// 	"offerId":  offerId,
	// 			// }

	// 			// docRef := store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
	// 			// _, err = docRef.Set(ctx, data)
	// 			// if err != nil {
	// 			// 	log.Println("SaveNotice error: ", uidBaseAcc, err)
	// 			// 	store, err = api.ReconnectFireStore("grayll-app-f3f3f3", 60)
	// 			// 	docRef = store.Collection("trades").Doc("users").Collection(uidBaseAcc).NewDoc()
	// 			// 	_, err = docRef.Set(ctx, data)
	// 			// 	if err != nil {
	// 			// 		log.Println("SaveNotice retried: ", uidBaseAcc, err)
	// 			// 		return
	// 			// 	}
	// 			// }

	// 			// data["id"] = docRef.ID
	// 			// data["uid"] = uidBaseAcc
	// 			// algoliaOrderIndex.SaveObject(data)
	// 		}
	// 		uidCounter, err := cache.GetUidFromPublicKey(counterAcc)
	// 		log.Println("uidCounter:", uidCounter)
	// 		if err == nil && uidCounter != "" {
	// 			// data := map[string]interface{}{
	// 			// 	"time":     time,
	// 			// 	"type":     "SELL",
	// 			// 	"asset":    asset,
	// 			// 	"amount":   amount,
	// 			// 	"xlmp":     xlmgrx,
	// 			// 	"totalxlm": totalxlm,
	// 			// 	"priceusd": grxusd,
	// 			// 	"totalusd": totalusd,
	// 			// 	"offerId":  offerId,
	// 			// }

	// 			// docRef := store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
	// 			// _, err = docRef.Set(ctx, data)
	// 			// if err != nil {
	// 			// 	store, err = api.ReconnectFireStore("grayll-app-f3f3f3", 60)
	// 			// 	log.Println("SaveNotice error: ", uidCounter, err)
	// 			// 	docRef = store.Collection("trades").Doc("users").Collection(uidCounter).NewDoc()
	// 			// 	_, err = docRef.Set(ctx, data)
	// 			// 	if err != nil {
	// 			// 		log.Println("SaveNotice retried: ", uidCounter, err)
	// 			// 		return
	// 			// 	}
	// 			// }
	// 			// data["id"] = docRef.ID
	// 			// data["uid"] = uidCounter
	// 			// algoliaOrderIndex.SaveObject(data)
	// 		}
	// 	}
	// 	//get pair xlm/usd
	// 	if trade.BaseAssetType == "native" && trade.CounterAssetCode == "USD" {
	// 		//log.Println("xlmUsdPrice:", xlmUsdPrice)
	// 		xlmusd = float64(trade.Price.N) / float64(trade.Price.D)
	// 		grxusd = xlmgrx * xlmusd
	// 		// _, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
	// 		// if err != nil {
	// 		// 	store, err = api.ReconnectFireStore("grayll-app-f3f3f3", 60)
	// 		// 	log.Println("Update xlmp error: ", err)
	// 		// 	_, err = store.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
	// 		// 	if err != nil {
	// 		// 		log.Println("ERROR Update xlmp: ", err)
	// 		// 	}
	// 		// }
	// 		// _, err = grzStore.Doc(PRICE_PATH).Set(ctx, map[string]interface{}{"xlmusd": xlmusd, "grxusd": grxusd}, firestore.MergeAll)
	// 		// if err != nil {
	// 		// 	log.Println("Update xlmp error: ", err)
	// 		// 	//return
	// 		// }
	// 		// cache.SetXLMPrice(xlmusd)
	// 		// cache.SetGRXUsd(grxusd)
	// 	}

	// }

	go func(ctx context.Context, xlmgrxChan chan float64) {
		ticker := time.NewTicker(1 * time.Minute)

		//isPause := false
		pauseUntil := int64(0)
		contxt := context.Background()

		doc, err := store.Doc(ADMIN_SETTING).Get(contxt)
		if err == nil {
			//isPause = doc.Data()["isPauseClosing"].(bool)
			pauseUntil = doc.Data()["pauseUntil"].(int64)
			log.Println("ispause", pauseUntil)
			cache.Client.HDel("pauseClosing", "isPause")
			cache.Client.HSet("pauseClosing", "pauseUntil", pauseUntil)
		}

		previousPrice := xlmgrx
		currentPrice := xlmgrx

		for {
			select {
			case <-ticker.C:
				log.Println("tick. xlmgrx", xlmgrx, previousPrice)
				if xlmgrx != previousPrice {
					docs, err := store.Collection("asset_algo_values/grxxlm/frame_01m").
						Where("UNIX_timestamp", ">=", time.Now().Unix()-15*60).OrderBy("UNIX_timestamp", firestore.Asc).Limit(1).Documents(ctx).GetAll()
					if err == nil && len(docs) > 0 {
						// check if price change in 15 minute
						lastPrice := docs[0].Data()["price"].(float64)
						log.Println("tick.lastPrice:", lastPrice, "currentPrice:", xlmgrx)
						if math.Abs(100*(xlmgrx-lastPrice))/lastPrice > 15 && time.Now().Unix() > pauseUntil {
							// pause close algo 60 minutes
							//isPause = true
							pauseUntil = time.Now().Unix() + int64(60*60)
							log.Println("tick.price change over 15% within 15 minute, pause until:", pauseUntil)
							store.Doc(ADMIN_SETTING).Set(contxt, map[string]interface{}{"pauseUntil": pauseUntil}, firestore.MergeAll)
							//cache.Client.HSet("pauseClosing", "isPause", true)
							cache.Client.HSet("pauseClosing", "pauseUntil", pauseUntil)
							// set to redis cache for gry1,..grz can read
						}
					}
				} else {
					log.Println("tick.Price is not change. lastPrice:", previousPrice, "currentPrice:", xlmgrx)
				}
			case newPrice := <-xlmgrxChan:

				previousPrice = currentPrice
				currentPrice = newPrice
				log.Println("pricechange.currentPrice:", currentPrice, "previousPrice", previousPrice)

				per := math.Abs(currentPrice-previousPrice) * 100 / previousPrice
				if per > 10 && time.Now().Unix() < pauseUntil {
					pauseUntil += int64(15 * 60)
					log.Println("pricechange.price change over 10%, pause until:", pauseUntil)
					store.Doc(ADMIN_SETTING).Set(contxt, map[string]interface{}{"pauseUntil": pauseUntil}, firestore.MergeAll)
					cache.Client.HSet("pauseClosing", "pauseUntil", pauseUntil)
				} else if per > 15 && time.Now().Unix() > pauseUntil {
					pauseUntil = time.Now().Unix() + int64(60*60)
					store.Doc(ADMIN_SETTING).Set(contxt, map[string]interface{}{"pauseUntil": pauseUntil}, firestore.MergeAll)
					cache.Client.HSet("pauseClosing", "pauseUntil", pauseUntil)
					log.Println("pricechange.price change over 15%, pause until:", pauseUntil)
				} else {
					log.Println("pricechange.price is not change over 10%. lastPrice:", previousPrice, "currentPrice:", currentPrice, "percent:", per)
				}

			case <-ctx.Done():
				return
			}

		}
	}(ctx, xlmgrxChan)

	// go func() {

	// 	err = stellar.HorizonClient.StreamTrades(ctx, tradeRequest, tradeHandler)
	// 	if err != nil {
	// 		log.Println("StreamTrades", err)

	// 	}
	// }()

	// Wait for SIGINT.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

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
