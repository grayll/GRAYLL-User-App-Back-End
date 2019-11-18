package api

import (
	jwttool "bitbucket.org/grayll/user-app-backend/jwt-tool"
	"cloud.google.com/go/firestore"
	"github.com/huyntsgs/stellar-service/assets"
)

type Config struct {
	IsMainNet        bool   `json:"isMainNet"`
	AssetCode        string `json:"assetCode"`
	IssuerAddress    string `json:"issuerAddress"`
	XlmLoanerSeed    string `json:"xlmLoanerSeed"`
	XlmLoanerAddress string `json:"xlmLoanerAddress"`
	RedisHost        string `json:"redisHost"`
	RedisPort        int    `json:"redisPort"`
	RedisPass        string `json:"redisPass"`
	HorizonUrl       string `json:"horizonUrl"`
	Host             string `json:"host"`
	Numberify        string `json:"numberify"`
}
type ApiContext struct {
	Store  *firestore.Client
	Jwt    *jwttool.JwtToolkit
	Cache  *RedisCache
	Config *Config
	Asset  assets.Asset
}
