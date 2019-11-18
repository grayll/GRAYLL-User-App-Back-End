package main

import (
	//"fmt"
	//"net/http"
	//"strings"
	//"sync"
	//"time"

	"context"
	"log"

	//"encoding/json"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	//"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirebaseAdminConfig struct {
	Typ             string `json:"type"`
	Project_id      string `json:"project_id"`
	Private_key_id  string `json:"private_key_id"`
	Private_key     string `json:"private_key"`
	Client_email    string `json:"client_email"`
	Client_id       string `json:"client_id"`
	Auth_uri        string `json: "auth_uri"`
	Token_uri       string `json:"token_uri"`
	Auth_provide    string `json:"auth_provider_x509_cert_url"`
	Client_cert_url string `json:"client_x509_cert_url"`
}

func GetFsClient1(isLocal bool) (*firestore.Client, error) {
	var client *firestore.Client
	var err error

	ctx := context.Background()
	if isLocal {
		// add content of firebase-admin json here
		bs := ``

		opt := option.WithCredentialsJSON([]byte(bs))
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalln("Error create new gray user app:", err)
		}

		client, err = app.Firestore(ctx)
	} else {
		projectID := "grayll-app-f3f3f3"
		conf := &firebase.Config{ProjectID: projectID}
		app, err := firebase.NewApp(ctx, conf)
		if err != nil {
			log.Fatalln(err)
		}

		client, err = app.Firestore(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return client, err
}

// GetFsClient gets firestore database instance of grayll user app
func GetFsClient(isLocal bool) (*firestore.Client, error) {
	var client *firestore.Client
	var err error

	ctx := context.Background()
	if isLocal {
		opt := option.WithCredentialsFile("grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json")
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalln("Error create new gray user app:", err)
		}
		client, err = app.Firestore(ctx)
	} else {
		projectID := "grayll-app-f3f3f3"
		conf := &firebase.Config{ProjectID: projectID}
		app, err := firebase.NewApp(ctx, conf)
		if err != nil {
			log.Fatalln(err)
		}

		client, err = app.Firestore(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return client, err
}

// GetGrxyzClient gets firestore database instance of grayll mvp app
func GetGrxyzClient(isLocal bool) (*firestore.Client, error) {
	var client *firestore.Client
	var err error

	ctx := context.Background()
	if isLocal {
		opt := option.WithCredentialsFile("grayll-mvp-firebase-adminsdk-jhq9x-cd9d4774ad.json")
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalln("Error create mvp firebase app:", err)
		}
		client, err = app.Firestore(ctx)
	} else {
		projectID := "grayll-mvp"
		conf := &firebase.Config{ProjectID: projectID}
		app, err := firebase.NewApp(ctx, conf)
		if err != nil {
			log.Fatalln(err)
		}

		client, err = app.Firestore(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return client, err
}
