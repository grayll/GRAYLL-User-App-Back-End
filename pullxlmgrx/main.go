package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

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

type Ledger struct {
	Embed struct {
		Records []map[string]interface{} `json:"records"`
	} `json:"_embedded"`
	Links struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
	} `json:"_links"`
}

func ParseLedgerData(url string) (*Ledger, error) {
	ledger := Ledger{}

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

func GetPriceData(record map[string]interface{}, asset string) (PriceTime, error) {
	var n, d float64
	var err error
	if p, ok := record["price"]; ok {

		prices := p.(map[string]interface{})
		n1, ok1 := prices["n"]
		d1, ok2 := prices["d"]
		if ok1 && ok2 {
			n = n1.(float64)
			d = d1.(float64)
		}
	}
	var uts time.Time
	if closeTime, ok := record["ledger_close_time"]; ok {
		ts := closeTime.(string)
		log.Println(ts)
		uts, err = time.Parse("2006-01-02T15:04:05Z", ts)
		if err != nil {
			log.Println(err)
			return PriceTime{}, err
		}
	}
	if asset == "xlm" {
		return PriceTime{P: n / d, UTS: uts.Unix()}, nil
	} else if asset == "grx" {
		return PriceTime{P: d / n, UTS: uts.Unix()}, nil
	}
	log.Println("end")
	return PriceTime{}, err
}

// func GetPrice(url string) (float64, float64, error) {
// 	embs, err := ParseLedgerData(url)
// 	if err != nil {
// 		return 0, 0, err
// 	}
//
// 	if len(embs.Embed.Records) > 0 {
// 		if price, ok := embs.Embed.Records[0]["price"]; ok {
// 			log.Println("price:", price)
// 			prices := price.(map[string]interface{})
// 			n, ok1 := prices["n"]
// 			d, ok2 := prices["d"]
// 			if ok1 && ok2 {
// 				return n.(float64), d.(float64), nil
// 			}
// 		}
// 	}
// 	return 0, 0, errors.New("price not found")
// }

// func GetPrices() {
// 	//grx xlm
// 	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
// 	n, d, err := GetPrice(url)
// 	log.Println(n, d, err, d/n)
//
// 	//xlm usd
// 	url = "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
// 	n, d, err = GetPrice(url)
// 	log.Println(n, d, err, n/d)
// }

func main() {
	//url := "https://horizon.stellar.org/trades?base_asset_type=native&base_asset_code=XLM&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&limit=200&order=desc"
	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=200"
	ledger, err := ParseLedgerData(url)
	if err != nil {
		log.Println("ParseLedgerData err:", err)
		//return "", "", 0, err
	}
	asset := "xlm"
	if strings.Contains(url, "GRX") {
		asset = "grx"
	}

	for _, record := range ledger.Embed.Records {
		price, err := GetPriceData(record, asset)
		if err != nil {
			continue
		}
		log.Println("priceTime:", price)
	}
}
