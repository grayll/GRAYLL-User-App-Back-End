package main

import (
	"context"
	"time"

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
	CheckUser("", "")
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
		log.Println(pk, doc[0].Data(), doc[0].Ref.ID)
		return
	}

	docs, err := client.Collection("users").Where("CreatedAt", ">=", 1595938305).OrderBy("CreatedAt", firestore.Desc).Documents(ctx).GetAll()
	cnt := 0
	for _, doc := range docs {
		cnt++
		log.Println("User Info:", doc.Data()["Email"], doc.Data()["PublicKey"], doc.Data()["Ip"], doc.Data()["Name"], doc.Data()["LName"], doc.Ref.ID, time.Unix(doc.Data()["CreatedAt"].(int64), 0).Format("2006-01-02 15:04:05"))
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
