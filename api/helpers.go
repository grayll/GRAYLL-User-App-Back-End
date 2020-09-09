package api

import (
	"crypto/rand"
	"errors"
	"log"
	"strconv"
	"strings"

	//"bytes"
	//"fmt"
	"io/ioutil"
	"net/http"

	"encoding/json"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/NeverBounce/NeverBounceApi-Go"
	"github.com/NeverBounce/NeverBounceApi-Go/models"
	"github.com/gin-gonic/gin"
	stellar "github.com/huyntsgs/stellar-service"
	"github.com/jinzhu/now"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/operations"
	build "github.com/stellar/go/txnbuild"
)

func VerifyEmailNeverBounce(neverBounceApiKey, email string) error {
	client := neverbounce.New(neverBounceApiKey)
	client.SetAPIVersion("v4.1")
	singleResults, err := client.Single.Check(&nbModels.SingleCheckRequestModel{
		Email:          email,
		AddressInfo:    true,
		CreditInfo:     true,
		Timeout:        10,
		HistoricalData: nbModels.HistoricalDataModel{RequestMetaData: 0},
	})
	if err != nil {
		return err
	}

	if singleResults.Result != "valid" {
		return errors.New("Invalid email address")
	}
	return nil
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
func GetIntValue(input interface{}) int64 {
	switch input.(type) {
	case int64:
		return input.(int64)
	case float64:
		return int64(input.(float64))
	}
	return 0
}
func Hash(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return string(h.Sum(nil))
}

// TimeIn returns the time in UTC if the name is "" or "UTC".
// It returns the local time if the name is "Local".
// Otherwise, the name is taken to be a location name in
// the IANA Time Zone database, such as "Africa/Lagos".
func TimeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}
func NewDate(t time.Time, h, m int, local string) time.Time {
	log.Println("NewDate-local:", local)
	loc, _ := time.LoadLocation(local)
	log.Println("NewDate-loc:", loc)
	return time.Date(t.Year(), t.Month(), t.Day(), h, m, 0, 0, loc)
}

func BeginOfWeek(t time.Time, local string) (time.Time, error) {
	location, _ := time.LoadLocation(local)

	myConfig := &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
		TimeFormats:  []string{"2006-01-02 15:04:05"},
	}

	//t := time.Date(2013, 11, 18, 17, 51, 49, 123456789, time.Now().Location()) // // 2013-11-18 17:51:49.123456789 Mon
	return myConfig.With(t).BeginningOfWeek(), nil
}

func BeginOfMonth(t time.Time, local string) (time.Time, error) {
	location, _ := time.LoadLocation(local)

	myConfig := &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
		TimeFormats:  []string{"2006-01-02 15:04:05"},
	}

	//t := time.Date(2013, 11, 18, 17, 51, 49, 123456789, time.Now().Location()) // // 2013-11-18 17:51:49.123456789 Mon
	return myConfig.With(t).BeginningOfMonth(), nil
}
func BeginOfNextWeek(t time.Time, local string) (time.Time, error) {
	location, _ := time.LoadLocation(local)

	myConfig := &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
		TimeFormats:  []string{"2006-01-02 15:04:05"},
	}

	//t := time.Date(2013, 11, 18, 17, 51, 49, 123456789, time.Now().Location()) // // 2013-11-18 17:51:49.123456789 Mon
	return myConfig.With(t).EndOfWeek().Add(time.Hour * 24 * 1), nil
}

func BeginOfNextMonth(t time.Time, local string) (time.Time, error) {
	location, _ := time.LoadLocation(local)

	myConfig := &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
		TimeFormats:  []string{"2006-01-02 15:04:05"},
	}

	//t := time.Date(2013, 11, 18, 17, 51, 49, 123456789, time.Now().Location()) // // 2013-11-18 17:51:49.123456789 Mon
	return myConfig.With(t).EndOfMonth().Add(time.Hour * 24 * 1), nil
}

func Hmac(secret, key string) string {
	hmc := hmac.New(sha256.New, []byte(secret))
	hmc.Write([]byte(key))
	enstr := hex.EncodeToString(hmc.Sum(nil))
	return enstr
}
func GinRespond(c *gin.Context, status int, errCode, msg string) {
	c.JSON(status, gin.H{
		"errCode": errCode, "msg": msg,
	})
	c.Abort()
}
func ExtractToken(r *http.Request) (string, error) {
	tokenEncrypted := r.Header.Get("Authorization")

	if !strings.Contains(tokenEncrypted, "Bearer ") {
		return "", errors.New("Authorization header not contain Bearer")
	}
	tokenEncrypted = tokenEncrypted[7:]
	return tokenEncrypted, nil
}

type Price struct {
	N string `json:"n"`
	D string `json:"d"`
}
type PriceTime struct {
	P   float64 `json:"p"`
	UTS int64   `json:"uts"`
}
type Embedded struct {
	Records []map[string]interface{} `json:"records"`
}
type LedgerPayment struct {
	Embed Embedded `json:"_embedded"`
}

func MergeAccount(mergedAccount, loanSeed string) error {
	_, _, err := stellar.MergeAccount(mergedAccount, loanSeed, build.CreditAsset{Code: "GRX", Issuer: "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333"})
	return err
}
func ParseLedgerData(url string) (*LedgerPayment, error) {
	ledger := LedgerPayment{}

	res, err := http.Get(url)
	if err != nil {
		log.Println("http.Get "+url+" error:", err)
		return nil, err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	//log.Println("data:", string(data))

	err = json.Unmarshal(data, &ledger)
	if err != nil {
		return nil, err
	}
	return &ledger, nil

}
func ParsePaymentFromTxHash(txHash string, client *horizonclient.Client) ([]operations.Payment, error) {
	opRequest := horizonclient.OperationRequest{ForTransaction: txHash}
	ops, err := client.Operations(opRequest)

	payments := make([]operations.Payment, 0)
	if err != nil {
		return payments, err
	}
	for _, record := range ops.Embedded.Records {
		log.Println(record)
		payment, ok := record.(operations.Payment)
		if ok {
			payments = append(payments, payment)
		}

	}
	return payments, nil
}
func GetLedgerInfo(url, publicKey, xlmLoaner string) (string, string, float64, error) {
	//"https://horizon-testnet.stellar.org/ledgers/1072717/payments"
	em, err := ParseLedgerData(url)
	if err != nil {
		log.Println("ParseLedgerData err:", err)
		return "", "", 0, err
	}
	log.Println("GetLedgerInfo-em", em)

	for _, record := range em.Embed.Records {
		if from, ok := record["from"]; ok && from.(string) == publicKey {
			to, _ := record["to"]
			amount, _ := record["amount"]

			log.Println("from:", from)
			log.Println(to)
			log.Println(amount)

			a, err := strconv.ParseFloat(amount.(string), 64)
			if err != nil {
				return "", "", 0, errors.New("Invalid ledger Id")
			}
			if to.(string) == xlmLoaner {
				return from.(string), to.(string), a, nil
			}
			return from.(string), to.(string), a, nil
		}
	}
	return "", "", 0, errors.New("Invalid ledger Id")
}

//
// func GetPriceData(record map[string]interface{}, asset string) (PriceTime, error) {
// 	var n, d float64
// 	var err error
// 	if price, ok := record["price"]; ok {
// 		log.Println("price:", price)
// 		prices := price.(map[string]interface{})
// 		n1, ok1 := prices["n"]
// 		d1, ok2 := prices["d"]
// 		if ok1 && ok2 {
// 			n = n1.(float64)
// 			d = d1.(float64)
// 		}
// 	}
// 	var uts time.Time
// 	if closeTime, ok := record["ledger_close_time"]; ok {
// 		ts := closeTime.(string)
// 		uts, err = time.Parse("2006-01-02T15:04:05-0700", ts)
// 	}
// 	if asset == "xlm" {
// 		return PriceTime{P: n / d, UTS: uts.Unix()}, nil
// 	} else {
// 		return PriceTime{P: d / n, UTS: uts.Unix()}, nil
// 	}
//
// }

func GetPrice(url string) (float64, float64, error) {
	embs, err := ParseLedgerData(url)
	if err != nil {
		return 0, 0, err
	}

	if len(embs.Embed.Records) > 0 {
		if price, ok := embs.Embed.Records[0]["price"]; ok {
			//log.Println("price:", price)
			prices := price.(map[string]interface{})
			n, ok1 := prices["n"]
			d, ok2 := prices["d"]
			if ok1 && ok2 {
				return n.(float64), d.(float64), nil
			}
		}
	}
	return 0, 0, errors.New("price not found")
}

func GetPrices() {
	//grx xlm
	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
	n, d, err := GetPrice(url)
	log.Println(n, d, err, d/n)

	//xlm usd
	url = "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
	n, d, err = GetPrice(url)
	log.Println(n, d, err, n/d)
}

func randStr(strSize int, randType string) string {

	var dictionary string
	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	if randType == "number" {
		dictionary = "0123456789"
	}
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}
