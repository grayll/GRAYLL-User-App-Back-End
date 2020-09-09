package main

import (
	"context"
	"time"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
	stellar "github.com/huyntsgs/stellar-service"

	"cloud.google.com/go/firestore"

	//"cloud.google.com/go/firestore"

	// "fmt"
	// "time"

	"testing"

	// "context"
	"log"

	//"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	// gray price
	UNIX_timestamp = "UNIX_timestamp"
)

func TestVerify(t *testing.T) {
	//QueryAlgoPosition()
	//checkExistUserMeta()
	CheckUser("", "sanchezbuenoelromeral@gmail.com")
	//DelUsers()
	//DelUser("andsoft88@gmail.com")
	//MergeAccount("GC5TQRTXZHXIOSKI4SXRVVIRALFZQ6SV2D7WCFUGP2M2TRN3UFRKCOD2")
}

func GetFloatValue(input interface{}) float64 {
	switch input.(type) {
	case int64:
		return float64(input.(int64))
	case float64:
		return input.(float64)
	}
	return 0
}
func GetClient() (*firestore.Client, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile("./grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	return app.Firestore(ctx)

}
func QueryAlgoPosition() {

	client, err := GetClient()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}
	ctx := context.Background()

	//users := make([]map[string]interface{}, 0)
	//var it *firestore.DocumentIterator
	//it = h.apiContext.Store.Collection("users_meta").Limit(limit).StartAfter(cursor).OrderBy("time", firestore.Desc).Documents(context.Background())
	usersmeta := client.Collection("users_meta")
	firstPage, err := usersmeta.Limit(20).Documents(ctx).GetAll()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}

	for _, doc := range firstPage {
		log.Println("doc:", doc.Data())
	}

}
func DelUsers() {
	client, err := GetClient()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}
	ctx := context.Background()

	config := parseConfig("config.json")
	cache, err := api.NewRedisCache(time.Hour*24, config)
	settingFields := []string{"IpConfirm", "MulSignature", "AppGeneral", "AppWallet", "AppAlgo", "MailGeneral", "MailWallet", "MailAlgo"}

	stellar.SetupParams(float64(1000), true)
	accounts := []string{"GBAKJG5LHVUR5SMHQH5XCO4MYFK3NLB7PRI3VWN73DQPV6DX5CJRIQS3", "GCZVYLRZ4BRXEWEXQBKLZKQK6GW7IHSHZGYMNAL3FDIV43RZ37NA5DWR", "GAGOZC3I5R6EUN2CAL62JXSNURDNSLCJ7MEA3UCC6QOSBR3VLUV52VK6",
		"GBCWPT2WEMOEC5WINH4NBEIZNG6H32WFZIHMMDLDTGWQHFFNBGGZWLQF", "GCM3LTXFZUZKHZCHGC266JHGCNJPUFUN2PHW5QI65RRD5QG2UHYUWGWC", "GCLOFLUZZJOHS7IHJBYTSGUBJJRI32Y73ZIFIQKCYBI72QTKURAJNMJI"}
	//"GALS6XF2FNME4XZGRH7PLZPXUZQHB6TZ6MS5VNIAYIYWXS3ATBVH5FWM", "GA6ZQM3WGGLEC53NKOVRXTNFSGTKW7YD4JJL2OTLWNID7HNVIBDUKPKO", "GBVQ2AHQL6A7FCNL7KM3XZWKIWU7F6FXNYIGMHS5UL5YAEZZ36SQN3VO", "GDBEAEKAJBMZDTI4JHJ52DFPXVI5OIJVMDMRGTNZKMNVZTH5D7NC5QWJ"}
	for _, acc := range accounts {
		err := api.MergeAccount(acc, "SATORSIMUQSQRV6H2TJRE7DO5YLES36JUHBGNQENSLXOAVBGHVI7K64B")
		log.Println(acc, err)
		time.Sleep(15 * time.Second)
		//cache.DelPublicKey()
		cache.DelNotice("", settingFields...)
	}

	// accounts := []string{"ZiWnX4Vo33DCNHp4mkxmo77CLDeWT6sPdzdLoCKkTbk", "KtLZAbvdqvZlfZzYkWE7n-z7cBGBfCvNz7GmDGmZayQ", "I5EOhOCmOJZpTFdwdoGtZJGutMHabN6cg1kj5SIWyWo",
	// 	"NRXQZnZM5L4nEfChBU2LpuJvkto-aaL8JdMel0eIX_Y", "mx9N-98No5v63_S8dqWNmlj92nSL8lmHJ9JXeT-VEHM", "Gxs4Ol6ZMkB2q_k-X054W8jo7D2xWOiLyeEwA3YXRxA"}
	//"oqA8g_cFJVOkNG0AjP3Ag2mlEzy1xwMy3w62nuX_uZE", "0nIWsBzoBCLfe8g7udwMWrqAcsV2xMM1PTw-ibdiWbw", "gPdMYAAnQIW2G1WpDfDoCVIOoLsyS2_eAs2thJ9nGf4",
	//"DxO7wOx4ua2VHuXhkdI1WlmE_vLpdeymgS2j6lvwhP8", "QkEfJeZVfx-185XB1Cyn4B1rO-ImMC3M9298_6cGUZE"}

	batch := client.Batch()
	for _, acc := range accounts {
		doc := client.Doc("users/" + acc)
		batch.Delete(doc)
		doc1 := client.Doc("users_meta/" + acc)
		batch.Delete(doc1)
	}
	_, err = batch.Commit(ctx)
	log.Println("delusers commit ", err)
}
func DelUser(pk string) {
	//1595938305
	client, err := GetClient()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}
	ctx := context.Background()

	config := parseConfig("config.json")
	cache, err := api.NewRedisCache(time.Hour*24, config)
	settingFields := []string{"IpConfirm", "MulSignature", "AppGeneral", "AppWallet", "AppAlgo", "MailGeneral", "MailWallet", "MailAlgo"}

	if pk != "" {
		doc, _ := client.Collection("users").Where("PublicKey", "==", pk).Documents(ctx).GetAll()

		if len(doc) == 0 {
			doc, _ = client.Collection("users").Where("Email", "==", pk).Documents(ctx).GetAll()
		}
		if len(doc) > 0 {
			log.Println("remove from users", doc[0].Data()["Email"])
			doc[0].Ref.Delete(ctx)

			publicKey := doc[0].Data()["PublicKey"].(string)
			api.MergeAccount(publicKey, "SATORSIMUQSQRV6H2TJRE7DO5YLES36JUHBGNQENSLXOAVBGHVI7K64B")
			cache.DelNotice(doc[0].Ref.ID, settingFields...)
			cache.DelPublicKey(doc[0].Ref.ID, publicKey)
		}

		// users_meta
		doc, _ = client.Collection("users_meta").Where("PublicKey", "==", pk).Documents(ctx).GetAll()

		if len(doc) == 0 {
			doc, _ = client.Collection("users_meta").Where("Email", "==", pk).Documents(ctx).GetAll()
		}

		if len(doc) > 0 {
			log.Println("remove from users_meta", doc[0].Data()["Email"])
			doc[0].Ref.Delete(ctx)
			publicKey := doc[0].Data()["PublicKey"].(string)
			api.MergeAccount(publicKey, "SATORSIMUQSQRV6H2TJRE7DO5YLES36JUHBGNQENSLXOAVBGHVI7K64B")
			cache.DelNotice(doc[0].Ref.ID, settingFields...)
			cache.DelPublicKey(doc[0].Ref.ID, publicKey)
		}
		return
	}

}
func CheckUser(uid, pk string) {
	//1595938305
	client, err := GetClient()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}
	ctx := context.Background()

	//var it *firestore.DocumentIterator
	//it = client.Collection("users_meta").Documents(ctx)
	//batch := client.Batch()

	if pk != "" {
		doc, _ := client.Collection("users").Where("PublicKey", "==", pk).Documents(ctx).GetAll()

		if len(doc) == 0 {
			doc, _ = client.Collection("users").Where("Email", "==", pk).Documents(ctx).GetAll()
		}
		log.Println(pk, doc[0].Ref.ID, doc[0].Data())
		return
	}

	docs, err := client.Collection("users").Where("CreatedAt", ">=", 1595938305).OrderBy("CreatedAt", firestore.Desc).Documents(ctx).GetAll()
	cnt := 0
	for _, doc := range docs {
		cnt++
		log.Println("User Info:", doc.Ref.ID, doc.Data()["Email"], doc.Data()["PublicKey"], doc.Data()["Ip"], doc.Data()["Name"], doc.Data()["LName"], time.Unix(doc.Data()["CreatedAt"].(int64), 0).Format("2006-01-02 15:04:05"))
	}
	log.Println("User Info:", cnt)
	// for {
	// 	doc, err := it.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}

	// 	//check whether exist in users collection
	// 	user, err := client.Doc("users/" + doc.Ref.ID).Get(ctx)
	// 	if err != nil {
	// 		log.Println("user not exist:", doc.Ref.ID)
	// 		//batch.Delete(doc.Ref)
	// 	} else {
	// 		log.Println(user.Ref.ID, "email:", user.Data()["Email"], GetFloatValue(doc.Data()["GRX"]))
	// 		activatedAt := int64(0)
	// 		if val, ok := user.Data()["ActivatedAt"]; ok {
	// 			activatedAt = val.(int64)
	// 		}
	// 		PublicKey := ""
	// 		if val, ok := user.Data()["PublicKey"]; ok {
	// 			PublicKey = val.(string)
	// 		}
	// 		batch.Set(doc.Ref, map[string]interface{}{
	// 			"Name":        user.Data()["Name"].(string),
	// 			"LName":       user.Data()["LName"].(string),
	// 			"Email":       user.Data()["Email"].(string),
	// 			"UserId":      user.Ref.ID,
	// 			"CreatedAt":   user.Data()["CreatedAt"].(int64),
	// 			"ActivatedAt": activatedAt,
	// 			"PublicKey":   PublicKey,
	// 		}, firestore.MergeAll)

	// 	}
	// }
	// _, err = batch.Commit(ctx)
	// if err != nil {
	// 	log.Println(err)
	// }

}

func checkExistUserMeta() {
	client, err := GetClient()
	if err != nil {
		log.Fatalln("Error create new firebase app:", err)
	}
	ctx := context.Background()

	var it *firestore.DocumentIterator
	it = client.Collection("users_meta").Documents(ctx)
	batch := client.Batch()

	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}

		//check whether exist in users collection
		user, err := client.Doc("users/" + doc.Ref.ID).Get(ctx)
		if err != nil {
			log.Println("user not exist:", doc.Ref.ID)
			//batch.Delete(doc.Ref)
		} else {
			log.Println(user.Ref.ID, "email:", user.Data()["Email"], GetFloatValue(doc.Data()["GRX"]))
			activatedAt := int64(0)
			if val, ok := user.Data()["ActivatedAt"]; ok {
				activatedAt = val.(int64)
			}
			PublicKey := ""
			if val, ok := user.Data()["PublicKey"]; ok {
				PublicKey = val.(string)
			}
			batch.Set(doc.Ref, map[string]interface{}{
				"Name":        user.Data()["Name"].(string),
				"LName":       user.Data()["LName"].(string),
				"Email":       user.Data()["Email"].(string),
				"UserId":      user.Ref.ID,
				"CreatedAt":   user.Data()["CreatedAt"].(int64),
				"ActivatedAt": activatedAt,
				"PublicKey":   PublicKey,
			}, firestore.MergeAll)

		}
	}
	_, err = batch.Commit(ctx)
	if err != nil {
		log.Println(err)
	}

}
