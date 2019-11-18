package utils

import (
	//"fmt"
	"log"
	"testing"
)

func TestGetCity(t *testing.T) {
	url := "http://www.geoplugin.net/json.gp?ip=42.118.9.124"
	city, country := GetCityCountry(url)
	log.Println(city)
	log.Println(country)
}
