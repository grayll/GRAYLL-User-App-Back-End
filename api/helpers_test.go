package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"testing"
	//"bitbucket.org/grayll/grayll.io-user-app-back-end/mail"
	stellar "github.com/huyntsgs/stellar-service"
	build "github.com/stellar/go/txnbuild"
)

func TestVerifyEmail(t *testing.T) {
	//VerifyEmailNeverBounce("private_f1db1a9eeccd5dd76347bf58596becf4", "huykbc@gmail.com")
	//VerifyEmailNeverBounce("private_f1db1a9eeccd5dd76347bf58596becf4", "huykbc1@gmail.com")

	mergeAccountNormal()
}
func mergeAccountNormal() {
	stellar.SetupParams(float64(1000), true)
	_, _, err := stellar.MergeAccountNormal("GBDLGL5BOMQ3DLKDXXOCQZYCDCAZRYRVTDR2GRQU4WVFUBFMHVZDRPS2", "SDZMO6BAXATHUEOVCXO5UUETIPEGLFMELIHXSKIDHFO3GWDDVGARDUGK",
		build.CreditAsset{Code: "GRX", Issuer: "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333"})
	log.Println(err)
}

func mergeAccount() {
	stellar.SetupParams(float64(1000), true)
	err := MergeAccount("GB4ZOYYZ7QYIINMZEU2JFKWJG4D3XLQHDIVSHKTP7TX7HECS3JS4P3KI", "")
	log.Println(err)

}

func mergeAccounts() {
	stellar.SetupParams(float64(1000), true)
	accounts := []string{"GBAKJG5LHVUR5SMHQH5XCO4MYFK3NLB7PRI3VWN73DQPV6DX5CJRIQS3", "GCZVYLRZ4BRXEWEXQBKLZKQK6GW7IHSHZGYMNAL3FDIV43RZ37NA5DWR", "GAGOZC3I5R6EUN2CAL62JXSNURDNSLCJ7MEA3UCC6QOSBR3VLUV52VK6",
		"GBCWPT2WEMOEC5WINH4NBEIZNG6H32WFZIHMMDLDTGWQHFFNBGGZWLQF", "GCM3LTXFZUZKHZCHGC266JHGCNJPUFUN2PHW5QI65RRD5QG2UHYUWGWC", "GCLOFLUZZJOHS7IHJBYTSGUBJJRI32Y73ZIFIQKCYBI72QTKURAJNMJI"}
	//"GALS6XF2FNME4XZGRH7PLZPXUZQHB6TZ6MS5VNIAYIYWXS3ATBVH5FWM", "GA6ZQM3WGGLEC53NKOVRXTNFSGTKW7YD4JJL2OTLWNID7HNVIBDUKPKO", "GBVQ2AHQL6A7FCNL7KM3XZWKIWU7F6FXNYIGMHS5UL5YAEZZ36SQN3VO", "GDBEAEKAJBMZDTI4JHJ52DFPXVI5OIJVMDMRGTNZKMNVZTH5D7NC5QWJ"}
	for _, acc := range accounts {
		err := MergeAccount(acc, "")
		log.Println(acc, err)
		time.Sleep(15 * time.Second)
	}

	accounts = []string{"EcJYFfKy8p7KmsI_OB2DgO0Uve-ILNcpti_DCJ_KA9Q", "6NJYVFDg9YUItCgeGMN7TYZsNHdG6Qqs2by2nEj3_ZA", "JDfjGeSff4Zbs4iA2Ts8f2yeewhrPiSXkkP-J06SXoQ",
		"Igve9tNXUdIcNnOxyG7wBPzbFTEXsaOZzAjJPWBe7lw", "Ur2JL0q3ZEFmjRGnLMkw3JuC8-2wvTSZ5tEh0yME-js", "58c9PeW2-cUp6KdyLtexmcsSxE3y_5IEEhSC5y1aibU",
		"oqA8g_cFJVOkNG0AjP3Ag2mlEzy1xwMy3w62nuX_uZE", "0nIWsBzoBCLfe8g7udwMWrqAcsV2xMM1PTw-ibdiWbw", "gPdMYAAnQIW2G1WpDfDoCVIOoLsyS2_eAs2thJ9nGf4",
		"DxO7wOx4ua2VHuXhkdI1WlmE_vLpdeymgS2j6lvwhP8", "QkEfJeZVfx-185XB1Cyn4B1rO-ImMC3M9298_6cGUZE"}
	// for _, acc := range accounts {
	// 	err := (acc, "SATORSIMUQSQRV6H2TJRE7DO5YLES36JUHBGNQENSLXOAVBGHVI7K64B")
	// 	log.Println(acc, err)
	// 	time.Sleep(15 * time.Second)
	// }
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
