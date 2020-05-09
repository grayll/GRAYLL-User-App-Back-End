package api

import (
	jwttool "bitbucket.org/grayll/grayll.io-user-app-back-end/jwt-tool"
	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/huyntsgs/stellar-service/assets"
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
}
type ApiContext struct {
	Store           *firestore.Client
	Jwt             *jwttool.JwtToolkit
	Cache           *RedisCache
	Config          *Config
	Asset           assets.Asset
	CloudTaskClient *cloudtasks.Client
	//AlgoliaClient *search.Client
	OrderIndex *search.Index
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
