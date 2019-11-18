package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"

	"encoding/json"

	"github.com/asaskevich/govalidator"
	"google.golang.org/api/iterator"
)

// func main() {
// 	http.HandleFunc("/federation", federation)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
// https://us-central1-grayll-app-f3f3f3.cloudfunctions.net/Federation?type=name&q=huynt580@gmail.com*grayll.io
func Query(w http.ResponseWriter, r *http.Request) {
	type Output struct {
		StellarAddress string `json:"stellar_address"`
		AccountId      string `json:"account_id"`
		MemoType       string `json:"memo_type,omitempty"`
		Memo           string `json:"memo,omitempty"`
	}
	var output Output
	setupCORS(&w)
	store, err := GetFsClient1(true)
	if err != nil {
		log.Fatal("Can not get fs client:", err)
		writeResponse(http.StatusInternalServerError, &output, &w)
		return
	}

	typeQ := r.FormValue("type")
	q := r.FormValue("q")

	if govalidator.IsNull(typeQ) || govalidator.IsNull(q) {
		writeResponse(http.StatusBadRequest, &output, &w)
		return
	}

	if err != nil {
		log.Fatalln("Federation: GetFsClient error: ", err)
	}

	switch typeQ {
	case "name":
		userInfo, _ := GetUserByField(store, "Federation", q)
		if userInfo == nil {
			writeResponse(http.StatusNotFound, &output, &w)
			return
		}

		output.StellarAddress = userInfo["Federation"].(string)
		output.AccountId = userInfo["PublicKey"].(string)

		err := writeResponse(http.StatusOK, &output, &w)
		if err != nil {
			fmt.Printf("error:%v\n", err)
		}
	case "id":
		userInfo, _ := GetUserByField(store, "PublicKey", q)
		if userInfo == nil {

			err := writeResponse(http.StatusNotFound, &output, &w)
			if err != nil {
				fmt.Printf("error:%v\n", err)
			}
			return
		}
		output.StellarAddress = userInfo["Federation"].(string)
		output.AccountId = userInfo["PublicKey"].(string)
		err := writeResponse(http.StatusOK, &output, &w)
		if err != nil {
			fmt.Printf("error:%v\n", err)
		}
	default:
		writeResponse(http.StatusBadRequest, &output, &w)
	}
}
func setupCORS(w *http.ResponseWriter) {
	//https://app.grayll.io
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
func GetUserByField(client *firestore.Client, field, value string) (map[string]interface{}, string) {
	ctx := context.Background()
	if field == "Uid" {
		docSnap, err := client.Doc("users/" + value).Get(ctx)
		if err != nil {
			return nil, ""
		}
		return docSnap.Data(), docSnap.Ref.ID
	}
	it := client.Collection("users").Where(field, "==", value).Limit(1).Documents(ctx)
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			return nil, ""
		}
		if doc == nil {
			return nil, ""
		}
		return doc.Data(), doc.Ref.ID
	}
	return nil, ""
}
func writeResponse(status int, data interface{}, w *http.ResponseWriter) error {
	(*w).WriteHeader(status)
	return json.NewEncoder(*w).Encode(data)
}
