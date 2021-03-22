package api

import (
	jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"
	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/huyntsgs/stellar-service/assets"

	ccm "github.com/orcaman/concurrent-map"
)

type Config struct {
	ProjectId         string `json:"projectId"`
	DataReportQueueId string `json:"queueId"`

	LocationId string `json:"locationId"`

	DataReportUrl string `json:"dataReportUrl"`

	ServiceAccountEmail string `json:"serviceAccountEmail"`

	IsMainNet         bool    `json:"isMainNet"`
	AssetCode         string  `json:"assetCode"`
	IssuerAddress     string  `json:"issuerAddress"`
	XlmLoanerSeed     string  `json:"xlmLoanerSeed"`
	XlmLoanerAddress  string  `json:"xlmLoanerAddress"`
	RedisHost         string  `json:"redisHost"`
	RedisPort         int     `json:"redisPort"`
	RedisPass         string  `json:"redisPass"`
	HorizonUrl        string  `json:"horizonUrl"`
	Host              string  `json:"host"`
	Numberify         string  `json:"numberify"`
	SuperAdminAddress string  `json:"superAdminAddress"`
	SuperAdminSeed    string  `json:"superAdminSeed"`
	SellingPrice      float64 `json:"sellingPrice"`
	SellingPercent    int     `json:"sellingPercent"`
	NeverBounceApiKey string  `json:"neverBounceApiKey"`
}
type ApiContext struct {
	Store           *firestore.Client
	Jwt             *jwttool.JwtToolkit
	JwtAdmin        *jwttool.JwtToolkit
	Cache           *RedisCache
	Config          *Config
	Asset           assets.Asset
	CloudTaskClient *cloudtasks.Client
	//AlgoliaClient *search.Client
	OrderIndex *search.Index
	BlockIPs   ccm.ConcurrentMap
}

type ReportDataSetting struct {
	Frequency     string `json:"Frequency"`
	TimeZone      string `json:"TimeZone"`
	TimeHour      int    `json:"TimeHour"`
	TimeMinute    int    `json:"TimeMinute"`
	WalletBalance bool   `json:"WalletBalance"`
	AccountValue  bool   `json:"AccountValue"`
	AccountProfit bool   `json:"AccountProfit"`
	OpenPosition  bool   `json:"OpenPosition"`
	UserId        string `json:"UserId,omitempty"`
}

type Contact struct {
	Name         string `json:"name"`
	LName        string `json:"lname"`
	Email        string `json:"email"`
	BusinessName string `json:"businessName"`
	Phone        string `json:"phone"`
	RefererUid   string `json:"refererUid"`
	RefererName  string `json:"rname"`
	RefererLName string `json:"rlname"`
}
type Input struct {
	GrayllTxId       string  `json:"grayllTxId"`
	Algorithm        string  `json:"algorithm"`
	GrxUsd           float64 `json:"grxUsd"`
	PositionValue    float64 `json:"positionValue"`
	PositionValueGRX float64 `json:"positionValueGRX"`
}
type KYC struct {
	Status       string `json:"Status,omitempty"`
	AppType      string `json:"AppType,omitempty"`
	Name         string `json:"Name,omitempty"`
	LName        string `json:"LName,omitempty"`
	Nationality  string `json:"Nationality,omitempty"`
	GovId        string `json:"GovId,omitempty"`
	DoB          string `json:"DoB,omitempty"`
	Company      string `json:"Company,omitempty"`
	Registration string `json:"Registration,omitempty"`
	Address1     string `json:"Address1,omitempty"`
	Address2     string `json:"Address2,omitempty"`
	City         string `json:"City,omitempty"`
	Country      string `json:"Country,omitempty"`
}
type KYCCom struct {
	Name         string `json:"Name,omitempty"`
	Registration string `json:"Registration,omitempty"`
	Address1     string `json:"Address1,omitempty"`
	Address2     string `json:"Address2,omitempty"`
	City         string `json:"City,omitempty"`
	Country      string `json:"Country,omitempty"`
}
type KYCDocs struct {
	GovPassport       string `json:"GovPassport,omitempty"`
	GovNationalIdCard string `json:"GovNationalIdCard,omitempty"`
	GovDriverLicense  string `json:"GovDriverLicense,omitempty"`

	// requires tax return
	Income6MPaySlips   string `json:"Income6MPaySlips,omitempty"`
	Income6MBankStt    string `json:"Income6MBankStt,omitempty"`
	Income2YTaxReturns string `json:"Income2YTaxReturns,omitempty"`

	// requires at least 2 docs
	AddressUtilityBill        string `json:"AddressUtilityBill,omitempty"`
	AddressBankStt            string `json:"AddressBankStt,omitempty"`
	AddressRentalAgreement    string `json:"AddressRentalAgreement,omitempty"`
	AddressPropertyTaxReceipt string `json:"AddressPropertyTaxReceipt,omitempty"`
	AddressTaxReturn          string `json:"AddressTaxReturn,omitempty"`

	AssetsShareStockCert string `json:"AssetsShareStockCert,omitempty"`
	Assets2MBankAccStt   string `json:"Assets2MBankAccStt,omitempty"`
	Assets2MRetireAccStt string `json:"Assets2MRetireAccStt,omitempty"`
	Assets2MInvestAccStt string `json:"Assets2MInvestAccStt,omitempty"`

	// company documents
	CertIncorporation string `json:"CertIncorporation,omitempty"`
	// require all docs
	Company2YTaxReturns       string `json:"Company2YTaxReturns,omitempty"`
	Company2YFinancialStt     string `json:"Company2YFinancialStt,omitempty"`
	Company2YBalanceSheets    string `json:"Company2YBalanceSheets,omitempty"`
	Company6MBankStt          string `json:"Company6MBankStt,omitempty"`
	Company6MInvestmentAccStt string `json:"Company6MInvestmentAccStt,omitempty"`
}
