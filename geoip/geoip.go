package geoip

import (
	"encoding/json"
	"net/http"
)

type geoIPData struct {
	Country string
	Region  string
	City    string
}

func GeoIP(w http.ResponseWriter, req *http.Request) {
	// Enable simple public access through CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var gd geoIPData
	gd.Country = req.Header.Get("X-AppEngine-Country")
	gd.Region = req.Header.Get("X-AppEngine-Region")
	gd.City = req.Header.Get("X-AppEngine-City")

	j, _ := json.Marshal(gd)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

//gcloud config set project grayll-app-f3f3f3
//gcloud functions deploy geoip --entry-point GeoIP --runtime go111 --trigger-http
