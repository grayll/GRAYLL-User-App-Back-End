package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"testing"

	"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
)

func TestVerifyEmail(t *testing.T) {
	//VerifyEmailNeverBounce("private_f1db1a9eeccd5dd76347bf58596becf4", "huykbc@gmail.com")
	//VerifyEmailNeverBounce("private_f1db1a9eeccd5dd76347bf58596becf4", "huykbc1@gmail.com")
	err := mail.SendMailRegistrationInvite("huykbc@gmail.com", "huy", "Sign up invite", "https://app.grayll.io", []string{"inviate", "test"})
	if err != nil {
		log.Println(err)
	}
}

func TestHmac(t *testing.T) {
	// // hmc := Hmac("kFOLecggKkSgaWGn_dyoFzZyuY8wFtzkvcncIU-J", "14refejeereire")
	// loc, _ := time.LoadLocation("Europe/Budapest")
	// now := time.Now().In(loc)
	// fmt.Println("ZONE : ", loc, " Time : ", now.Unix(), now.Hour())  // UTC
	// fmt.Println("Utc Time : ", time.Now().Unix(), time.Now().Hour()) // UTC

	// timeInUTC := time.Date(now.Year(), now.Month(), now.Day(), 11, 8, 0, 0, time.UTC)

	// fmt.Println("h : ", timeInUTC.Hour(), " m : ", timeInUTC.Minute(), "day", timeInUTC.Day()) // UTC
}

var zoneDirs = []string{
	// Update path according to your OS
	"/usr/share/zoneinfo/",
	"/usr/share/lib/zoneinfo/",
	"/usr/lib/locale/TZ/",
}

var zoneDir string

// func main() {
//     for _, zoneDir = range zoneDirs {
//         ReadFile("")
//     }
// }

func ReadFile(path string) {
	files, _ := ioutil.ReadDir(zoneDir + path)
	for _, f := range files {
		if f.Name() != strings.ToUpper(f.Name()[:1])+f.Name()[1:] {
			continue
		}
		if f.IsDir() {
			ReadFile(path + "/" + f.Name())
		} else {
			fmt.Println((path + "/" + f.Name())[1:])
		}
	}
}

// func TestZone(t *testing.T) {
// 	for _, zoneDir = range zoneDirs {
// 		ReadFile("")
// 	}
// }

// func TestParsePayment(t *testing.T) {
// 	em, err := ParseLedgerData("https://horizon.stellar.org/ledgers/26871047/payments")
// 	if err != nil {
// 		log.Println(err)
// 	} else {
// 		log.Println(em)
// 	}

// 	if len(em.Embed.Records) >= 2 {
// 		record := em.Embed.Records[1]

// 		if from, ok := record["from"]; ok {
// 			to, _ := record["to"]
// 			amount, _ := record["amount"]

// 			log.Println(from)
// 			log.Println(to)
// 			log.Println(amount)
// 			//return from, to, amount, nil
// 		} else {
// 			log.Println("Can not find key from")
// 			//return "", "", "", errors.New("Invalid ledger Id")
// 		}
// 	}
// }

// func TestParsePriceGRX(t *testing.T) {
// 	url := "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=GRX&counter_asset_issuer=GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333&order=desc&limit=1"
// 	n, d, err := GetPrice(url)
// 	log.Println(n, d, err, d/n)

// 	url = "https://horizon.stellar.org/trades?base_asset_type=native&counter_asset_type=credit_alphanum4&counter_asset_code=USD&counter_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&order=desc&limit=1"
// 	n, d, err = GetPrice(url)
// 	log.Println(n, d, err, n/d)
// }
