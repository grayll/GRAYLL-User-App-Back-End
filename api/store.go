package api

import (
	"context"
	//"log"

	//"bitbucket.org/grayll/grayll.io-user-app-back-end/models"
	"cloud.google.com/go/firestore"
)

func GetUserByField(client *firestore.Client, field, value string) (map[string]interface{}, string) {
	ctx := context.Background()
	if field == "Uid" {
		docSnap, err := client.Doc("users/" + value).Get(ctx)
		if err != nil {
			return nil, ""
		}
		return docSnap.Data(), docSnap.Ref.ID
	}
	docs, err := client.Collection("users").Where(field, "==", value).Limit(1).Documents(ctx).GetAll()
	if err == nil && len(docs) > 0 {
		return docs[0].Data(), docs[0].Ref.ID
	}
	return nil, ""
}
