package main

import (
	"bytes"
	//"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	//"bitbucket.org/grayll/user-app-backend/utils"

	"log"
	"net/http"

	//"bitbucket.org/grayll/user-app-backend/mail"

	"time"
	//firebase "firebase.google.com/go"
)

type Setting struct {
	IpConfirm    bool `json:"ipConfirm,omitempty"`
	MulSignature bool `json:"mulSignature,omitempty"`
}
type Tfa struct {
	BackupCode string `json:"BackupCode"`
	DataURL    string `json:"DataURL"`
	Enable     bool   `json:"Enable"`
	TempSecret string `json:"TempSecret"`
}

type UserInfo struct {
	Uid   string `json:"Uid"`
	Name  string `json:"Name"`
	Email string `json:"Email"`
	Token string `json:"Token,omitempty"`
	//Tfa                Tfa     `json:"Tfa"`
	UserSetting        Setting `json:"UserSetting"`
	Ip                 string  `json:"Ip,omitempty"`
	CreatedAt          int64   `json:"CreatedAt,omitempty"`
	PublicKey          string  `json:"PublicKey"`
	EncryptedSecretKey string  `json:"EncryptedSecretKey"`
	SecretKeySalt      string  `json:"SecretKeySalt"`
	Federation         string  `json:"Federation"`
}

type geoIPData struct {
	Country string
	Region  string
	City    string
}

//https://medium.com/@shangyilim/verifying-phone-numbers-with-firebase-phone-authentication-on-your-backend-for-free-7a9bef326d02
func GeoIP(w http.ResponseWriter, req *http.Request) {
	var gd geoIPData
	gd.Country = req.Header.Get("X-AppEngine-Country")
	gd.Region = req.Header.Get("X-AppEngine-Region")
	gd.City = req.Header.Get("X-AppEngine-City")

	j, _ := json.Marshal(gd)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func ExtractToken(r *http.Request) (string, error) {
	tokenEncrypted := r.Header.Get("Authorization")
	fmt.Println("ExtractToken: tokenEncrypted: ", tokenEncrypted)
	if !strings.Contains(tokenEncrypted, "Bearer ") {
		return "", errors.New("Authorization header not contain Bearer")
	}
	tokenEncrypted = tokenEncrypted[7:]
	return tokenEncrypted, nil
}

func VerifyRecapchaToken(w http.ResponseWriter, r *http.Request) {
	SetupCORS(&w)
	var respApi = make(map[string]string)
	if (*r).Method == "OPTIONS" {
		return
	}
	var respData struct {
		Success      bool      `json:"success"`      // whether this request was a valid reCAPTCHA token for your site
		Score        float64   `json:"score"`        // the score for this request (0.0 - 1.0)
		Action       string    `json:"action"`       // the action name for this request (important to verify)
		Challenge_ts time.Time `json:"challenge_ts"` // timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
		Hostname     string    `json:"hostname"`     // the hostname of the site where the reCAPTCHA was solved
		//Errors       string    `json:"error-codes"`
	}

	token, err := ExtractToken(r)
	if err != nil {
		fmt.Printf("VerifyRecapchaToken: Authorization header does not contain Bearer\n", err)
		respData.Success = false
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(respData)
		return
	}

	url := "https://www.google.com/recaptcha/api/siteverify"
	secret := "6LfYI7EUAAAAAKGxMquwzN5EsJHlp-0_bfspQhGI"
	url = fmt.Sprintf("%s?secret=%s&response=%s", url, secret, token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		log.Println("verifyRecapchaToken: Can not create new req")
		respApi["status"] = "fail"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&respApi)
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("verifyRecapchaToken: call client.Do() error %v\n", err)
		respApi["status"] = "fail"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&respApi)
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("verifyRecapchaToken: call resp.Body error %v\n", err)
		respApi["status"] = "fail"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&respApi)
		return
	}

	if respData.Score <= 0.5 {
		// Can not process this request
		respApi["status"] = "fail"
	} else {
		respApi["status"] = "success"
	}

	err = json.NewEncoder(w).Encode(&respApi)
	if err != nil {
		log.Printf("verifyRecapchaToken: call Encode respApi error %v\n", err)
	}
}

func SetupCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Origin, x-requested-with, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
