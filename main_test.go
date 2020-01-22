package main

import (
	"fmt"
	"time"

	"testing"

	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	// gray price
	UNIX_timestamp = "UNIX_timestamp"

//price          = "price"
)

func QueryGrz() {
	//var startTs int64 = 1569949920 + int64(1440*60)
	//var startTs int64 = 1563469200
	var startTs int64 = 1572519421

	ctx := context.Background()
	opt := option.WithCredentialsFile("../grayll-mvp-firebase-adminsdk-jhq9x-cd9d4774ad.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	//fsclient := FireStoreClient{client}
	var curPrice float64
	var curTime int64
	//var pair string
	docPath := "grz_price_frames/grzusd/frame_30m"
	it := client.Collection(docPath).Where(UNIX_timestamp, ">", startTs).OrderBy(UNIX_timestamp, firestore.Asc).Limit(10).Documents(context.Background())
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			return
		}

		if err != nil {
			log.Println("Error get latest price: ", err)
			return
		}

		curTime = doc.Data()[UNIX_timestamp].(int64)
		//pair = doc.Data()["pair"].(string)
		curPrice = doc.Data()["price"].(float64)

		fmt.Printf("time: %d - price: %f \n", curTime, curPrice)
		priceTime := time.Unix(curTime, 0).Format("2006-01-02 15:04:05")
		fmt.Println("priceTime:", priceTime)

		// Add missing data for next day
		// client.Collection("pair_frames/gryusd/frame_01d").Add(context.Background(), map[string]interface{}{
		// 	UNIX_timestamp: curTime,
		// 	price:          curPrice,
		// })

	}

}

func DelGrz() {
	//var startTs int64 = 1569949920 + int64(1440*60)
	//var startTs int64 = 1563469200

	var startTs int64 = 1572519421

	ctx := context.Background()
	opt := option.WithCredentialsFile("../grayll-mvp-firebase-adminsdk-jhq9x-cd9d4774ad.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	//fsclient := FireStoreClient{client}
	var curPrice float64
	var curTime int64
	//var pair string
	docPaths := []string{"grz_price_frames/grzusd/frame_01m", "grz_price_frames/grzusd/frame_05m", "grz_price_frames/grzusd/frame_15m", "grz_price_frames/grzusd/frame_30m", "grz_price_frames/grzusd/frame_01h",
		"grz_price_frames/grzusd/frame_04h", "grz_price_frames/grzusd/frame_01d"}

	for _, docPath := range docPaths {
		it := client.Collection(docPath).Where(UNIX_timestamp, "<", startTs).OrderBy(UNIX_timestamp, firestore.Desc).Limit(500).Documents(context.Background())
		for {
			doc, err := it.Next()
			if err == iterator.Done {
				return
			}

			if err != nil {
				log.Println("Error get latest price: ", err)
				return
			}

			curTime = doc.Data()[UNIX_timestamp].(int64)
			//pair = doc.Data()["pair"].(string)
			curPrice = doc.Data()["price"].(float64)

			fmt.Printf("time: %d - price: %f \n", curTime, curPrice)
			priceTime := time.Unix(curTime, 0).Format("2006-01-02 15:04:05")
			fmt.Println("priceTime:", priceTime, doc.Ref.ID)

			//client.Collection(docPath).Doc(doc.Ref.ID).Delete(ctx)

			// Add missing data for next day
			// client.Collection("pair_frames/gryusd/frame_01d").Add(context.Background(), map[string]interface{}{
			// 	UNIX_timestamp: curTime,
			// 	price:          curPrice,
			// })

		}
	}

}

func TestVerify(t *testing.T) {
	//QueryOriginalGry()
}

func QueryOriginalGrz() {
	//var startTs int64 = 1569949920 + int64(1440*60)
	//var startTs int64 = 1563469200

	var startTs int64 = 1573636800 - 2*60*60

	ctx := context.Background()
	opt := option.WithCredentialsFile("../grayll-mvp-firebase-adminsdk-jhq9x-cd9d4774ad.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	//fsclient := FireStoreClient{client}
	var curPrice float64
	var curTime int64
	//var pair string
	docPath := "grz_price"
	it := client.Collection(docPath).Where("UNIX_timestamp", ">=", startTs).OrderBy(UNIX_timestamp, firestore.Desc).Documents(context.Background())
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			return
		}

		if err != nil {
			log.Println("Error get latest price: ", err)
			return
		}

		curTime = doc.Data()[UNIX_timestamp].(int64)
		//pair = doc.Data()["pair"].(string)
		curPrice = doc.Data()["price"].(float64)

		priceTime := time.Unix(curTime, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("time: %d (%s) - price: %f \n", curTime, priceTime, curPrice)

		// Add missing data for next day
		// client.Collection("pair_frames/gryusd/frame_01d").Add(context.Background(), map[string]interface{}{
		// 	UNIX_timestamp: curTime,
		// 	price:          curPrice,
		// })

	}

}

func QueryOriginalGry() {
	//var startTs int64 = 1569949920 + int64(1440*60)
	//var startTs int64 = 1563469200

	//	var startTs int64 = 1575684300

	ctx := context.Background()
	opt := option.WithCredentialsFile("./grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	//fsclient := FireStoreClient{client}
	var curPrice float64
	var curTime int64
	//var pair string
	docPath := "asset_algo_values/gryusd/frame_01m"
	it := client.Collection(docPath).OrderBy("UNIX_timestamp", firestore.Desc).Limit(20).Documents(context.Background())
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			return
		}

		if err != nil {
			log.Println("Error get latest price: ", err)
			return
		}

		curTime = doc.Data()[UNIX_timestamp].(int64)
		//pair = doc.Data()["pair"].(string)
		curPrice = doc.Data()["price"].(float64)

		priceTime := time.Unix(curTime, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("time: %d (%s) - price: %f \n", curTime, priceTime, curPrice)
		//fmt.Println("priceTime:", priceTime)

		// Add missing data for next day
		// client.Collection("pair_frames/gryusd/frame_01d").Add(context.Background(), map[string]interface{}{
		// 	UNIX_timestamp: curTime,
		// 	price:          curPrice,
		// })

	}

}

// func TestQueryDB(t *testing.T) {
// 	var startTs int64 = 1556989920 - 5

// 	ctx := context.Background()
// 	opt := option.WithCredentialsFile("./grayll-mvp-firebase-adminsdk-jhq9x-cd9d4774ad.json")
// 	app, err := firebase.NewApp(ctx, nil, opt)
// 	if err != nil {
// 		log.Fatalln("Error create new firebase app:", err)
// 	}

// 	client, err := app.Firestore(ctx)
// 	if err != nil {
// 		log.Fatalln("Error create new firebase app:", err)
// 	}

// 	//fsclient := FireStoreClient{client}
// 	var curPrice float64
// 	var curTime int64
// 	var pair string
// 	it := client.Collection("medianclose").Where(UNIX_timestamp, ">=", startTs).OrderBy(UNIX_timestamp, firestore.Asc).Limit(8).Documents(context.Background())
// 	for {
// 		doc, err := it.Next()
// 		if err == iterator.Done {
// 			return
// 		}

// 		if err != nil {
// 			log.Println("Error get latest price: ", err)
// 			return
// 		}

// 		curTime = doc.Data()[UNIX_timestamp].(int64)
// 		pair = doc.Data()["pair"].(string)
// 		curPrice = doc.Data()[medain_close].(float64)

// 		fmt.Printf("coin: %s - time: %d - price: %f \n", pair, curTime, curPrice)
// 		priceTime := time.Unix(curTime, 0).Format("2006-01-02 15:04:05")
// 		fmt.Println("priceTime:", priceTime)

// 	}

// }
