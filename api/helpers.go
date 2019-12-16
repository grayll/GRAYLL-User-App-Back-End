package api

import (
	"crypto/rand"
	"errors"
	"log"
	"strconv"

	//"bytes"
	//"fmt"
	"io/ioutil"
	"net/http"

	//"os"
	"encoding/json"

	"github.com/gin-gonic/gin"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/utils"
)

func GinRespond(c *gin.Context, status int, errCode, msg string) {
	c.JSON(status, gin.H{
		"errCode": errCode, "msg": msg,
	})
	c.Abort()
}

type Price struct {
	N string `json:"n"`
	D string `json:"d"`
}
type Embedded struct {
	Records []map[string]interface{} `json:"records"`
}
type LedgerPayment struct {
	Embed Embedded `json:"_embedded"`
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

func GetLedgerInfo(url, publicKey, xlmLoaner string) (string, string, float64, error) {
	//"https://horizon-testnet.stellar.org/ledgers/1072717/payments"
	em, err := ParseLedgerData(url)
	if err != nil {
		log.Println("ParseLedgerData err:", err)
		return "", "", 0, err
	}

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

func GetPrice(url string) (float64, float64, error) {
	embs, err := ParseLedgerData(url)
	if err != nil {
		return 0, 0, err
	}

	if len(embs.Embed.Records) > 0 {
		if price, ok := embs.Embed.Records[0]["price"]; ok {
			log.Println("price:", price)
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
