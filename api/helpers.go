package api

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strconv"
	"strings"

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
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/operations"
	build "github.com/stellar/go/txnbuild"
)

func VerifyKycStatus(userInfo map[string]interface{}) (int, []string) {

	var ok, ok1, ok2, ok3, ok4, ok5 bool
	msg := []string{}
	kyc, okm := userInfo["Kyc"]
	if !okm {
		return 1, msg
	}

	kycDoc, okn := userInfo["KycDocs"]
	if !okn {
		return 2, msg
	}
	kycMap := kyc.(map[string]interface{})
	//kycMap := kyc.(map[string]interface{})
	kycDocMap := kycDoc.(map[string]interface{})

	_, ok = kycDocMap[GovPassport]
	_, ok1 = kycDocMap[GovNationalIdCard]
	_, ok2 = kycDocMap[GovDriverLicense]
	if !(ok || ok1 || ok2) {
		// Lack of gov id
		msg = append(msg, "One of GovermentID")
	}
	_, ok = kycDocMap[Income6MPaySlips]
	_, ok1 = kycDocMap[Income6MBankStt]
	_, ok2 = kycDocMap[Income2YTaxReturns]

	if !(ok || ok1) {
		// Lack of
		msg = append(msg, "Proof of Income  Last 6 Months Pay Slips or Proof of Income Last 6 Months Bank Statements")
	}
	if !ok2 {
		msg = append(msg, "Proof of Income Last 2 Years Tax Returns")
	}

	//log.Println("income:", ok, ok1, ok2, msg)
	// at least two docs
	sum := 0
	submsg := ""
	if _, ok = kycDocMap[AddressUtilityBill]; ok {
		sum += 1
		submsg = "Address Utility Bill"
	}
	if _, ok1 = kycDocMap[AddressBankStt]; ok1 {
		sum += 1
		submsg = "Address Bank Statement"
	}
	if _, ok2 = kycDocMap[AddressRentalAgreement]; ok2 {
		sum += 1
		submsg = "Address Rental/Lease Agreement"
	}
	if _, ok3 = kycDocMap[AddressPropertyTaxReceipt]; ok3 {
		sum += 1
		submsg = "Address Property Tax Receipt"
	}
	if _, ok4 = kycDocMap[AddressTaxReturn]; ok4 {
		sum += 1
		submsg = "Address Tax Return"
	}

	if sum == 1 {
		msg = append(msg, "One more Proof of Address Document apart from "+submsg)
	}
	if sum == 0 {
		msg = append(msg, "2 Proof of Address Documents")
	}
	_, ok = kycDocMap[AssetsShareStockCert]
	_, ok1 = kycDocMap[Assets2MBankAccStt]
	_, ok2 = kycDocMap[Assets2MRetireAccStt]
	_, ok3 = kycDocMap[Assets2MInvestAccStt]

	if !(ok || ok1 || ok2 || ok3) {
		msg = append(msg, "One of Proof of Asset documents")
	}

	if kycMap["AppType"].(string) != "Personal" {

		if _, ok = kycDocMap[CertIncorporation]; !ok {
			msg = append(msg, "Certificate of Incorporation/Formation")
		}
		if _, ok1 = kycDocMap[Company2YTaxReturns]; !ok1 {
			msg = append(msg, "Last 2 Years Tax Returns")
		}
		if _, ok2 = kycDocMap[Company2YFinancialStt]; !ok2 {
			msg = append(msg, "Last 2 Years Financial Statements")
		}
		if _, ok3 = kycDocMap[Company2YBalanceSheets]; !ok3 {
			msg = append(msg, "Last 2 Years Balance Sheets")
		}
		if _, ok4 = kycDocMap[Company6MBankStt]; !ok4 {
			msg = append(msg, "Last 6 Months Bank Statements")
		}
		if _, ok5 = kycDocMap[Company6MInvestmentAccStt]; !ok5 {
			msg = append(msg, "Last 6 Months Investment Account Statements")
		}
		if !(ok && ok1 && ok2 && ok3 && ok4 && ok5) {
			return 0, msg
		}

	}

	return 0, msg

}
func VerifyKycAuditResult(userInfo map[string]interface{}) (int, string) {

	var ok, ok1, ok2, ok3, ok4, ok5 bool
	var govRes, incomeRes, addressRes, assetRes, companyRes bool
	var res, res1, res2, res3, res4, res5 interface{}
	msg := ""
	kyc, okm := userInfo["Kyc"]
	if !okm {
		msg := "Not update kyc information"
		return 1, msg
	}

	kycDoc, okn := userInfo["KycDocs"]
	if !okn {
		msg := "Not upload kyc docs"
		return 2, msg
	}
	kycMap := kyc.(map[string]interface{})
	//kycMap := kyc.(map[string]interface{})
	kycDocMap := kycDoc.(map[string]interface{})

	res, ok = kycDocMap[GovPassportRes]
	res1, ok1 = kycDocMap[GovNationalIdCardRes]
	res2, ok2 = kycDocMap[GovDriverLicenseRes]

	if (ok && res.(int64) == 1) || (ok1 && res1.(int64) == 1) || (ok2 && res2.(int64) == 1) {
		govRes = true
	}
	res, ok = kycDocMap[Income6MPaySlipsRes]
	res1, ok1 = kycDocMap[Income6MBankSttRes]
	res2, ok2 = kycDocMap[Income2YTaxReturnsRes]

	if ((ok && res.(int64) == 1) || (ok1 && res1.(int64) == 1)) && (ok2 && res2.(int64) == 1) {
		incomeRes = true
	}

	sum := 0
	submsg := ""
	if res, ok = kycDocMap[AddressUtilityBillRes]; ok && res.(int64) == 1 {
		sum += 1
		submsg = "Address Utility Bill"
	}
	if res1, ok1 = kycDocMap[AddressBankSttRes]; ok1 && res1.(int64) == 1 {
		sum += 1
		submsg = "Address Bank Statement"
	}
	if res2, ok2 = kycDocMap[AddressRentalAgreementRes]; ok2 && res2.(int64) == 1 {
		sum += 1
		submsg = "Address Rental/Lease Agreement"
	}
	if res3, ok3 = kycDocMap[AddressPropertyTaxReceiptRes]; ok3 && res3.(int64) == 1 {
		sum += 1
		submsg = "Address Property Tax Receipt"
	}
	if res4, ok4 = kycDocMap[AddressTaxReturnRes]; ok4 && res4.(int64) == 1 {
		sum += 1
		submsg = "Address Tax Return"
	}

	if sum == 1 {
		msg = msg + ", One more Address document apart from " + submsg
	}
	if sum == 0 {
		msg = msg + ", Two Address documents"
	}
	if sum >= 2 {
		addressRes = true
	}

	//Asset
	res, ok = kycDocMap[AssetsShareStockCertRes]
	res1, ok1 = kycDocMap[Assets2MBankAccSttRes]
	res2, ok2 = kycDocMap[Assets2MRetireAccSttRes]
	res3, ok3 = kycDocMap[Assets2MInvestAccSttRes]

	if (ok && res.(int64) == 1) || (ok1 && res1.(int64) == 1) || (ok2 && res2.(int64) == 1) || (ok3 && res3.(int64) == 1) {
		assetRes = true
	} else {
		msg = msg + ", One of Asset documents"
	}

	if kycMap["AppType"].(string) != "Personal" {
		// _, ok = userInfo["KycCom"]
		// if !ok {
		// 	return 3, "Not update company information"
		// }

		res, ok = kycDocMap[CertIncorporationRes]
		//msg = msg + "Certificates of Incorporation/Formation"

		res1, ok1 = kycDocMap[Company2YTaxReturnsRes]
		//msg = msg + ", Company Last 2 Years Tax Returns"

		res2, ok2 = kycDocMap[Company2YFinancialSttRes]
		//msg = msg + "Company Last 2 Years Financial Statement"

		res3, ok3 = kycDocMap[Company2YBalanceSheetsRes]
		//msg = msg + ", Company Last 2 Years Balance Sheets"

		res4, ok4 = kycDocMap[Company6MBankSttRes]
		//msg = msg + ", Company Last 6 Months Bank Statement"

		res5, ok5 = kycDocMap[Company6MInvestmentAccSttRes]
		//msg = msg + ", Company Last 6 Months Investment Account Statement"

		if (ok && res.(int64) == 1) && (ok1 && res1.(int64) == 1) && (ok2 && res2.(int64) == 1) && (ok3 && res3.(int64) == 1) && (ok4 && res4.(int64) == 1) && (ok5 && res5.(int64) == 1) {
			companyRes = true
		} else {
			msg = msg + ", One of Asset documents"
		}
		//
		// if res, ok = kycDocMap[CertIncorporationRes]; !ok {
		// 	msg = msg + "Certificates of Incorporation/Formation"
		// }
		// if _, ok1 = kycDocMap[Company2YTaxReturnsRes]; !ok1 {
		// 	msg = msg + ", Company Last 2 Years Tax Returns"
		// }
		// if _, ok2 = kycDocMap[Company2YFinancialSttRes]; !ok2 {
		// 	msg = msg + "Company Last 2 Years Financial Statement"
		// }
		// if _, ok3 = kycDocMap[Company2YBalanceSheetsRes]; !ok3 {
		// 	msg = msg + ", Company Last 2 Years Balance Sheets"
		// }
		// if _, ok4 = kycDocMap[Company6MBankSttRes]; !ok4 {
		// 	msg = msg + ", Company Last 6 Months Bank Statement"
		// }
		// if _, ok5 = kycDocMap[Company6MInvestmentAccSttRes]; !ok5 {
		// 	msg = msg + ", Company Last 6 Months Investment Account Statement"
		// }
		// if !(ok && ok1 && ok2 && ok3 && ok4 && ok5) {
		// 	return 0, msg
		// }

	}
	if strings.HasPrefix(msg, ", ") {
		msg = msg[2:]
	}

	if kycMap["AppType"].(string) == "Personal" {
		if govRes && incomeRes && addressRes && assetRes {
			return 0, msg
		} else {
			return 100, msg
		}
	} else {
		if govRes && incomeRes && addressRes && assetRes && companyRes {
			return 0, msg
		} else {
			return 100, msg

		}
	}
}
func (h UserHandler) GetUserValue(userId string) UserValue {
	doc, err := h.apiContext.Store.Doc("users_meta/" + userId).Get(context.Background())
	userVale := UserValue{}
	if err == nil {
		userVale.xlm = doc.Data()["XLM"].(float64)
		userVale.grx = doc.Data()["GRX"].(float64)
		if val, ok := doc.Data()["total_gry1_current_position_value_$"]; ok {
			if val1, ok1 := val.(float64); ok1 {
				userVale.algoValue += val1
			} else if val1, ok1 := val.(int64); ok1 {
				userVale.algoValue += float64(val1)
			}
		}
		if val, ok := doc.Data()["total_gry2_current_position_value_$"]; ok {
			if val1, ok1 := val.(float64); ok1 {
				userVale.algoValue += val1
			} else if val1, ok1 := val.(int64); ok1 {
				userVale.algoValue += float64(val1)
			}

		}
		if val, ok := doc.Data()["total_gry3_current_position_value_$"]; ok {
			if val1, ok1 := val.(float64); ok1 {
				userVale.algoValue += val1
			} else if val1, ok1 := val.(int64); ok1 {
				userVale.algoValue += float64(val1)
			}
		}
		if val, ok := doc.Data()["total_grz_current_position_value_$"]; ok {
			if val1, ok1 := val.(float64); ok1 {
				userVale.algoValue += val1
			} else if val1, ok1 := val.(int64); ok1 {
				userVale.algoValue += float64(val1)
			}
		}
		userVale.pk = doc.Data()["PublicKey"].(string)

		if val, ok := doc.Data()["GRY"].(float64); ok {
			userVale.gry = val
		}
		if val, ok := doc.Data()["USDC"].(float64); ok {
			userVale.usdc = val
		}
	}

	return userVale

}
func GetFriendlyName(fieldName string) string {

	const (
		GovPassport       = "GovPassport"
		GovNationalIdCard = "GovNationalIdCard"
		GovDriverLicense  = "GovDriverLicense"

		// requires tax return
		Income6MPaySlips   = "Income6MPaySlips"
		Income6MBankStt    = "Income6MBankStt"
		Income2YTaxReturns = "Income2YTaxReturns"

		// requires at least 2 docs
		AddressUtilityBill        = "AddressUtilityBill"
		AddressBankStt            = "AddressBankStt"
		AddressRentalAgreement    = "AddressRentalAgreement"
		AddressPropertyTaxReceipt = "AddressPropertyTaxReceipt"
		AddressTaxReturn          = "AddressTaxReturn"

		AssetsShareStockCert = "AssetsShareStockCert"
		Assets2MBankAccStt   = "Assets2MBankAccStt"
		Assets2MRetireAccStt = "Assets2MRetireAccStt"
		Assets2MInvestAccStt = "Assets2MInvestAccStt"

		// company documents
		CertIncorporation = "CertIncorporation"
		// require all docs
		Company2YTaxReturns       = "Company2YTaxReturns"
		Company2YFinancialStt     = "Company2YFinancialStt"
		Company2YBalanceSheets    = "Company2YBalanceSheets"
		Company6MBankStt          = "Company6MBankStt"
		Company6MInvestmentAccStt = "Company6MInvestmentAccStt"
	)

	name := ""
	switch fieldName {
	case GovPassport:
		name = "Passport"
		break
	case GovNationalIdCard:
		name = "NationalId Card"
		break
	case GovDriverLicense:
		name = "Driver License"
		break
	case Income6MPaySlips:
		name = "6 Months Pay Slips (Income)"
		break
	case Income6MBankStt:
		name = "6 Month Bank Statement (Income)"
		break
	case Income2YTaxReturns:
		name = "2 Years Tax Returns (Income)"
		break
	case AddressUtilityBill:
		name = "Address Utility Bill"
		break
	case AddressBankStt:
		name = "Address Bank Statement"
		break
	case AddressRentalAgreement:
		name = "Address Rental/Lease Agreement"
		break
	case AddressPropertyTaxReceipt:
		name = "Address Property Tax Receipt"
		break
	case AddressTaxReturn:
		name = "Address Tax Return"
		break
	case AssetsShareStockCert:
		name = "Share/Stock Certification (Asset)"
		break
	case Assets2MBankAccStt:
		name = "2 Months Bank Account Statement (Asset)"
		break
	case Assets2MRetireAccStt:
		name = "2 Months Retire Account Statement (Asset)"
		break
	case Assets2MInvestAccStt:
		name = "2 Months Investment Account Statement (Asset)"
		break
		// company
	case CertIncorporation:
		name = "Certification Incorporation (Company)"
		break
	case Company2YTaxReturns:
		name = "2 Years Tax Returns (Company)"
		break
	case Company2YFinancialStt:
		name = "2 Years Financial Statement (Company)"
		break
	case Company2YBalanceSheets:
		name = "2 Years Balance Sheets (Company)"
		break
	case Company6MBankStt:
		name = "6 Months Bank Statement (Company)"
		break
	case Company6MInvestmentAccStt:
		name = "6 Months Investment Account Statement (Company)"
		break

	}

	return name

}
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

	loc, _ := time.LoadLocation(local)

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
func MergeAccountNChangeTrust(mergedAccount, loanSeed string) error {
	//_, _, err := stellar.MergeAccountNChangeTrust(mergedAccount, loanSeed, build.CreditAsset{Code: "GRX", Issuer: "GAQQZMUNB7UCL2SXHU6H7RZVNFL6PI4YXLPJNBXMOZXB2LOQ7LODH333"})
	return nil
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
func ParseOrderBookData(url string) (*horizon.OrderBookSummary, error) {
	ledger := horizon.OrderBookSummary{}

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

func GetOrderBook(url string) (float64, float64, error) {
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
