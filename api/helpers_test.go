package api

import (
	"log"
	"testing"
)

func TestParsePayment(t *testing.T) {
	em, err := ParseLedgerData("https://horizon.stellar.org/ledgers/26871047/payments")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(em)
	}

	if len(em.Embed.Records) >= 2 {
		record := em.Embed.Records[1]

		if from, ok := record["from"]; ok {
			to, _ := record["to"]
			amount, _ := record["amount"]

			log.Println(from)
			log.Println(to)
			log.Println(amount)
			//return from, to, amount, nil
		} else {
			log.Println("Can not find key from")
			//return "", "", "", errors.New("Invalid ledger Id")
		}
	}
}

func TestParsePriceGRX(t *testing.T) {
	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
	n, d, err := GetPrice(url)
	log.Println(n, d, err, d/n)

	url = "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
	n, d, err = GetPrice(url)
	log.Println(n, d, err, n/d)
}
