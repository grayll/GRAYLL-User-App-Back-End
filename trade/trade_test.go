package main

import (
	"fmt"
	"testing"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
)

func TestParseOrderBook(t *testing.T) {
	url := `https://horizon.grayll.io/order_book?selling_asset_type=native&selling_asset_code=XLM&buying_asset_type=credit_alphanum4&buying_asset_code=GRX&buying_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&limit=1`

	orderBook, err := api.ParseOrderBookData(url)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(orderBook)
}
