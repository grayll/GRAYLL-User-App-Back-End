package main

import (
	"context"
	"encoding/json"
	"flag"

	"fmt"
	"log"
	"strconv"

	"strings"

	"os"
	"os/signal"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	"cloud.google.com/go/firestore"

	//"github.com/SherClockHolmes/webpush-go"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"

	//"github.com/antigloss/go/logger"
	// "net/http"
	// _ "net/http/pprof"

	//"cloud.google.com/go/profiler"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/operations"
	//"gopkg.in/natefinch/lumberjack.v2"
)

type (
	ReserveSupply struct {
		// Hot Wallet 1 Account:
		// GDVHBPPOKDFGRW5ZJAFZ7V6JYVTZEK5CJG22YJU2VV6LNRN2WBWQGHLQ
		Hotwallet1 float64 `json:"hotwallet1"`

		// Hot Wallet 2 Account:
		// GBDLGL5BOMQ3DLKDXXOCQZYCDCAZRYRVTDR2GRQU4WVFUBFMHVZDRPS2
		Hotwallet2 float64 `json:"hotwallet2"`

		// GRAYLL Super Admin Account:
		// GDVYGHTLKVYAPMX76AY24FLNNFV3SJRODMYE3PIUFLQHZ67XBOT7FPVE
		SuperAdmin float64 `json:"superAdminAddress"`

		// GRAYLL System Reserve (Liquidity) Account 1:
		// GDPFZB33CRJOZKKH3HADRIHZHBBNEFLAZ6QUMEMS7F7B34XI45EXEANA
		SystemReserve1 float64 `json:"systemReserve1"`

		// GRAYLL System Reserve (Liquidity) Account 2:
		// GCOWVK3SPNVJ3FXSKGUTDA5ESK3FE4ONITZPHM5VYICTNHHGC4GMOMLQ
		SystemReserve2         float64 `json:"systemReserve2"`
		CurrentAvailableSupply float64 `json:"availableSupply"`
		TotalAvailableSupply   float64 `json:"totalAvailableSupply"`
		TotalSupply            float64 `json:"totalSupply"`
		Las                    float64 `json:"las"`
	}

	Config struct {
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
		Hotwallet1        string  `json:"hotwallet1"`
		Hotwallet2        string  `json:"hotwallet2"`
		SystemReserve1    string  `json:"systemReserve1"`
		SystemReserve2    string  `json:"systemReserve2"`
		SuperAdminSeed    string  `json:"superAdminSeed"`
		SellingPrice      float64 `json:"sellingPrice"`
		SellingPercent    int     `json:"sellingPercent"`
	}
)

func main() {

	// cache and firestore and webpush
	var store *firestore.Client
	var gry1Client *firestore.Client
	var err error
	var configPath string
	var config *Config
	reserveSupply := new(ReserveSupply)

	isMainNet := flag.Bool("mainnet", false, "run on mainnet or testnet")
	isProd := flag.Bool("prod", false, "run on prod or local")
	flag.Parse()
	projectId := "grayll-app-f3f3f3"

	if *isProd {
		configPath = "/home/huykbc/"
		config = parseConfig(configPath + "config1.json")
	} else {
		configPath = "/home/bc/go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/streaming/"
		if *isMainNet {
			config = parseConfig(configPath + "config1.json")
		} else {
			config = parseConfig(configPath + "config.json")
		}
	}

	// log.SetOutput(&lumberjack.Logger{
	// 	Filename:   configPath + "log/payment-log.txt",
	// 	MaxSize:    5, // megabytes
	// 	MaxBackups: 2,
	// 	MaxAge:     10,   //days
	// 	Compress:   true, // disabled by default
	// })

	store, err = GetFsClient(true, configPath)
	if err != nil {
		log.Fatalln("main: GetFsClient error: ", err)
	}

	gry1Client, err = GetGryClient(configPath + "grayll-gry-1-balthazar-fb-admin.json")
	if err != nil {
		log.Fatalln("main: gry1Client error: ", err)
	}

	stellar.SetupParam(float64(1000), config.IsMainNet, config.HorizonUrl)

	ttl, _ := time.ParseDuration("12h")
	cache, err := api.NewRedisCacheHost(ttl, config.RedisHost, config.RedisPass, config.RedisPort)
	// cnt := 0
	// if err != nil {
	// 	for {
	// 		cnt++
	// 		time.Sleep(1 * time.Second)
	// 		cache, err = api.NewRedisCacheHost(ttl, config.Host, config.RedisPass, config.RedisPort)
	// 		if err == nil {
	// 			break
	// 		}
	// 		if cnt > 120 {
	// 			log.Fatalln("Can not connect to redis", err)
	// 		}

	// 	}
	// }

	client := search.NewClient("BXFJWGU0RM", "ef746e2d654d89f2a32f82fd9ffebf9e")
	algoliaTransferIndex := client.InitIndex("transfers-ua")

	//getCurrentAvailbleSupply(config, reserveSupply)

	ctx := context.Background()

	// _, err = gry1Client.Doc("gry_1_algo_input/function_2_input").Set(ctx, map[string]interface{}{
	// 	"Hotwallet1":             reserveSupply.Hotwallet1,
	// 	"Hotwallet2":             reserveSupply.Hotwallet2,
	// 	"SystemReserveAccount1":  reserveSupply.SystemReserve1,
	// 	"SystemReserveAccount2":  reserveSupply.SystemReserve2,
	// 	"SuperAdminAccount":      reserveSupply.SuperAdmin,
	// 	"CurrentAvailableSupply": reserveSupply.CurrentAvailableSupply,
	// 	"TotalSupply":            reserveSupply.TotalSupply,
	// 	"TotalAvailableSupply":   reserveSupply.TotalAvailableSupply,
	// 	"Las":                    reserveSupply.Las,
	// }, firestore.MergeAll)
	// if err != nil {
	// 	log.Fatalln("[ERROR]: set reserve supply error: ", err)
	// }
	// cache.SetFunc2LAS(reserveSupply.Las)

	log.Println("main net:", config.RedisHost, config.RedisPort)

	opRequest := horizonclient.OperationRequest{Cursor: "now", IncludeFailed: false}
	ctx, cancel := context.WithCancel(context.Background())
	operationHandler := func(op operations.Operation) {
		defer func() {
			if v := recover(); v != nil {
				log.Println("capture a panic:", v)
				log.Println("avoid crashing the program")
			}
		}()
		log.Println("op type:", op.GetType())
		if op.GetType() == "set_option" {
			bytes, err := json.Marshal(op)
			if err != nil {
				return
			}

			var setoption operations.SetOptions
			err = json.Unmarshal(bytes, &setoption)
			if err != nil {
				log.Println(err)
				return
			}

			//setoption.ID
			log.Println("op id:", setoption.ID, setoption.HomeDomain)
			if setoption.HomeDomain != "" && setoption.HomeDomain != "grayll.io" {
				log.Println("SourceAccount:", setoption.SourceAccount)
				uid, err := cache.GetUidFromPublicKey(setoption.SourceAccount)
				if err == nil && uid != "" {
					log.Println("user belongs to grayll but set to other home domain:", uid, setoption.SourceAccount)
					// user belongs to grayll but set to other home domain.
				}
			}

		} else if op.GetType() == "payment" {
			bytes, err := json.Marshal(op)
			if err != nil {
				return
			}

			var payment operations.Payment
			err = json.Unmarshal(bytes, &payment)
			if err != nil {
				log.Println(err)
				return
			}

			if payment.Asset.Type == "native" || payment.Asset.Code == config.AssetCode {
				code := payment.Code
				if code == "" {
					code = "XLM"
				}
				log.Printf("Account %s send to %s with %v %s\n", payment.From, payment.To, payment.Amount, code)

				// Notification for user account
				// Check whether account belongs to grayll system
				var uid string
				//isIncomingPayment := true
				uid, err = cache.GetUidFromPublicKey(payment.To)
				if err != nil {
					cache, err = api.NewRedisCacheHost(ttl, config.RedisHost, config.RedisPass, config.RedisPort)
					uid, err = cache.GetUidFromPublicKey(payment.To)
					if err != nil {
						return
					}
				}
				uidFrom, err := cache.GetUidFromPublicKey(payment.From)
				if err != nil {
					cache, err = api.NewRedisCacheHost(ttl, config.RedisHost, config.RedisPass, config.RedisPort)
					uidFrom, err = cache.GetUidFromPublicKey(payment.From)
					if err != nil {
						return
					}
				}

				updateBlFn := func(isIncome bool, idUser string) {
					amount, _ := strconv.ParseFloat(payment.Amount, 64)
					if !isIncome {
						amount = -amount
					}
					fieldPath := ""
					value := float64(0)
					if idUser == config.Hotwallet1 {
						fieldPath = "Hotwallet1"
						reserveSupply.Hotwallet1 += amount
						value = reserveSupply.Hotwallet1
					} else if idUser == config.Hotwallet2 {
						fieldPath = "Hotwallet2"
						reserveSupply.Hotwallet2 += amount
						value = reserveSupply.Hotwallet2
					} else if idUser == config.SuperAdminAddress {
						fieldPath = "SuperAdminAccount"
						reserveSupply.SuperAdmin += amount
						value = reserveSupply.SuperAdmin
					} else if idUser == config.SystemReserve1 {
						reserveSupply.SystemReserve1 += amount
						fieldPath = "SystemReserveAccount1"
						value = reserveSupply.SystemReserve1
					} else if idUser == config.SystemReserve2 {
						fieldPath = "SystemReserveAccount2"
						reserveSupply.SystemReserve2 += amount
						value = reserveSupply.SystemReserve2
					}
					if fieldPath != "" {
						updateLas(reserveSupply)
						gry1Client.Doc("gry_1_algo_input/function_2_input").Set(ctx, map[string]interface{}{
							fieldPath: value, "Las": reserveSupply.Las}, firestore.MergeAll)
						cache.SetFunc2LAS(reserveSupply.Las)
					}
				}

				f := func(isIncome bool, idUser string) {
					// Get subscription and push
					// log.Println("idUser: ", idUser)
					// log.Printf("Account %s send to %s with %v %s\n", payment.From, payment.To, payment.Amount, code)
					trans := "https://stellar.expert/explorer/public/search?term=" + payment.ID
					var title, body, amount, counter, issuer string

					if isIncome {
						title = "GRAYLL | New Incoming Payment"
						body = fmt.Sprintf("You have a new incoming payment of %s %s from %s.", payment.Amount, code, payment.From)
						counter = payment.From
						amount = "+" + string(payment.Amount)
					} else {
						title = "GRAYLL | New Outgoing Payment"
						body = fmt.Sprintf("You have a new outgoing payment of %s %s to %s.", payment.Amount, code, payment.To)
						counter = payment.To
						amount = "-" + string(payment.Amount)
					}

					if strings.Contains(code, "GRX") {
						issuer = "Grayll"
						// asset is GRX, check if account related to GRX balance and update
						updateBlFn(isIncome, idUser)
					} else {
						issuer = "Stellar"
					}

					notice := map[string]interface{}{
						"type":   "wallet",
						"title":  title,
						"isRead": false,
						"body":   body,
						"time":   payment.LedgerCloseTime.Unix(),
						"txId":   payment.ID,
						// "vibrate": []int32{100, 50, 100},
						// "icon":    "https://app.grayll.io/favicon.ico",
						// "data": map[string]interface{}{
						// 	"url": config.Host + "/notifications/overview",
						// },
						"counter": counter,
						"amount":  amount,
						"asset":   code,
						"issuer":  issuer,
					}
					algoliaNotice := map[string]interface{}{
						"type":    "wallet",
						"title":   title,
						"body":    body,
						"time":    payment.LedgerCloseTime.Unix(),
						"txId":    payment.ID,
						"counter": counter,
						"amount":  amount,
						"asset":   code,
						"issuer":  issuer,
						"uid":     idUser,
					}

					ctx := context.Background()

					// Save to firestore
					if code == "GRXT" {
						code = "GRX"
					}
					docRef := store.Collection("notices").Doc("wallet").Collection(idUser).NewDoc()
					_, err = docRef.Set(ctx, notice)
					if err != nil {
						log.Println("SaveNotice error: ", idUser, err)
						store, err = api.ReconnectFireStore(projectId, 60)
						docRef = store.Collection("notices").Doc("wallet").Collection(idUser).NewDoc()
						_, err = docRef.Set(ctx, notice)
						if err != nil {
							log.Println("ERROR SaveNotice retried: ", idUser, err)
							return
						}
					}

					_, err = store.Doc("users_meta/"+idUser).Update(ctx, []firestore.Update{
						{Path: "UrWallet", Value: firestore.Increment(1)},
					})
					if err != nil {
						log.Println("SaveNotice update error: ", err)
						//return
					}
					amountF, _ := strconv.ParseFloat(payment.Amount, 64)
					if isIncome {
						_, err = store.Doc("users_meta/"+idUser).Update(ctx, []firestore.Update{
							{Path: code, Value: firestore.Increment(float64(amountF))},
						})
						if err != nil {
							log.Println("Update users-meta error: ", err)
							store, err = api.ReconnectFireStore(projectId, 60)
							_, err = store.Doc("users_meta/"+idUser).Update(ctx, []firestore.Update{
								{Path: code, Value: firestore.Increment(float64(amountF))},
							})
							log.Println("Update users-meta retried error: ", err)
						}
					} else {
						_, err = store.Doc("users_meta/"+idUser).Update(ctx, []firestore.Update{
							{Path: code, Value: firestore.Increment(float64(-amountF))},
						})
						if err != nil {

							log.Println("Update update error: ", err)
							store, err = api.ReconnectFireStore(projectId, 60)
							_, err = store.Doc("users_meta/"+idUser).Update(ctx, []firestore.Update{
								{Path: code, Value: firestore.Increment(float64(-amountF))},
							})
							if err != nil {
								log.Println("Update update retried error: ", err)
							}
						}
					}
					// Mail wallet
					go func() {
						mailWallet, err := cache.GetNotice(idUser, "MailWallet")
						if err != nil {
							log.Println("Can not get MailWallet setting from cache:", err)
						} else {
							// check setting and send mail
							if mailWallet == "1" {
								userRef, err := store.Doc("users/" + idUser).Get(ctx)
								if userRef != nil {
									err = mail.SendNoticeMail(userRef.Data()["Email"].(string), userRef.Data()["Name"].(string), title, []string{body, "Stellar | Transaction ID | " + trans})
									if err != nil {
										log.Println("SendNoticeMail error: ", err)
									}
								}
							}
						}
					}()

					// index algolianotice
					algoliaNotice["id"] = docRef.ID
					_, err = algoliaTransferIndex.SaveObject(algoliaNotice)
					if err != nil {
						log.Println("algoliaTransferIndex error:", err)
					}
				}

				// From normal user to normal user
				if uidFrom != config.SuperAdminAddress && uid != config.SuperAdminAddress {
					if uid != "" {
						f(true, uid)
					}
					if uidFrom != "" {
						f(false, uidFrom)
					}
				} else if uidFrom != config.SuperAdminAddress && uid == config.SuperAdminAddress {
					f(true, uid)
				} else if uidFrom == config.SuperAdminAddress && uid != config.SuperAdminAddress {
					f(false, uidFrom)
				}

				// if uidFrom == config.SuperAdminAddress {
				// 	f(false, uidFrom)
				// } else if uid == config.SuperAdminAddress {
				// 	f(true, uid)
				// } else {
				// 	if uid != "" {
				// 		f(true, uid)
				// 	}
				// 	if uidFrom != "" {
				// 		f(false, uidFrom)
				// 	}
				// }

				if uidFrom != "" || uid != "" {
					log.Printf("Account %s send to %s with %v %s\n", payment.From, payment.To, payment.Amount, code)
				}
			}
		}
	}

	go func() {
		err = stellar.HorizonClient.StreamPayments(ctx, opRequest, operationHandler)
		if err != nil {
			log.Println("StreamPayments error:", err)
		}
	}()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:8090", nil))
	// }()

	// Wait for SIGINT.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()
}

func getAccountBalance(accountId string) float64 {
	bl := float64(0)

	accRequest := horizonclient.AccountRequest{AccountID: accountId}
	accDetail, err := stellar.HorizonClient.AccountDetail(accRequest)
	if err != nil {
		log.Println("[ERROR] get account detail failure", accountId, err)
		return bl
	}

	for _, blance := range accDetail.Balances {
		if blance.Asset.Code == "GRX" {
			bl, _ = strconv.ParseFloat(blance.Balance, 64)
			break
		}
	}
	return bl
}

func getAssetSupply() float64 {
	assetReq := horizonclient.AssetRequest{ForAssetCode: "GRX", ForAssetIssuer: "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333"}
	asset, err := stellar.HorizonClient.Assets(assetReq)
	if err != nil {
		log.Println("[ERROR]-Error get asset detail", err)
		return 0
	} else {
		if len(asset.Embedded.Records) > 0 {
			amount, _ := strconv.ParseFloat(asset.Embedded.Records[0].Amount, 0)
			return amount
		} else {
			return 0
		}
	}
}

func getCurrentAvailbleSupply(config *Config, reserveSupply *ReserveSupply) float64 {
	// get balance of accounts
	reserveSupply.Hotwallet1 = getAccountBalance(config.Hotwallet1)
	reserveSupply.Hotwallet2 = getAccountBalance(config.Hotwallet2)
	reserveSupply.SuperAdmin = getAccountBalance(config.SuperAdminAddress)
	reserveSupply.SystemReserve1 = getAccountBalance(config.SystemReserve1)
	reserveSupply.SystemReserve2 = getAccountBalance(config.SystemReserve2)
	reserveSupply.TotalSupply = getAssetSupply()
	reserveSupply.CurrentAvailableSupply = reserveSupply.TotalSupply -
		(reserveSupply.Hotwallet1 + reserveSupply.Hotwallet2 + reserveSupply.SuperAdmin + reserveSupply.SystemReserve1 + reserveSupply.SystemReserve2)
	reserveSupply.TotalAvailableSupply = 0.8 * reserveSupply.TotalSupply
	reserveSupply.Las = reserveSupply.CurrentAvailableSupply * 100 / reserveSupply.TotalAvailableSupply
	return reserveSupply.CurrentAvailableSupply
}

func updateLas(reserveSupply *ReserveSupply) {
	//reserveSupply.TotalSupply = getAssetSupply()
	reserveSupply.CurrentAvailableSupply = reserveSupply.TotalSupply -
		(reserveSupply.Hotwallet1 + reserveSupply.Hotwallet2 + reserveSupply.SuperAdmin + reserveSupply.SystemReserve1 + reserveSupply.SystemReserve2)

	reserveSupply.Las = reserveSupply.CurrentAvailableSupply * 100 / (0.8 * reserveSupply.TotalSupply)
}

func ReconnectFireStore(projectId string, timeout int) (*firestore.Client, error) {
	cnt := 0
	var client *firestore.Client
	var err error
	ctx := context.Background()
	for {
		cnt++
		time.Sleep(1 * time.Second)
		client, err = firestore.NewClient(ctx, projectId)
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
