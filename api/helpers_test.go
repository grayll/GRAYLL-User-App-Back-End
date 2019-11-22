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
