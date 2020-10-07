package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	//	"bitbucket.org/grayll/grayll.io-user-app-back-end/api"
)

func parseConfig(path string) *Config {

	var raw []byte
	var err error
	config := new(Config)
	if path == "" {
		path = "config.json"
	}
	if strings.HasPrefix(path, "http") {
		res, err := http.Get(path)
		if err != nil {
			log.Fatalf("Read file: %s error: %v\n", path, err)
		}
		defer res.Body.Close()

		raw, err = ioutil.ReadAll(res.Body)

	} else {
		raw, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Read file: %s error: %v\n", path, err)
		}
	}

	err = json.Unmarshal(raw, config)
	if err != nil {
		log.Fatalf("Parse json error: %v\n", err)
	}

	return config

}
