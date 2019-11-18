package api

import (
	"context"
	//"log"

	//"bitbucket.org/grayll/user-app-backend/models"
	"cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
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
