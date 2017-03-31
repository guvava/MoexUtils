package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	moexlib "github.com/agareev/MoexLib/monitoring"
	config "github.com/agareev/MoexLib/other"
)

/*
{
"marketdata": {
        "metadata": {
                "UPDATETIME": {"type": "time", "bytes": 10, "max_size": 0}
        },
        "columns": ["UPDATETIME"],
        "data": [
                ["18:01:08"]
        ]
}}
*/

var configuration config.Config

// Request JSON description
type Request struct {
	Marketdata struct {
		Columns []string   `json:"columns"`
		Data    [][]string `json:"data"`
	} `json:"marketdata"`
	Trade struct {
		Columns []string   `json:"columns"`
		Data    [][]string `json:"data"`
	} `json:"trades"`
}

//
func randNum() string {
	return fmt.Sprintf("&rand=%v", rand.Intn(1000))
}

func urlReturn(engine, market, typeOfCheck string) string {
	// engine - stock, futures, currency, stock
	// market - index, forts, selt, shares
	var parturl string
	if typeOfCheck == "marketdata" {
		parturl = "/securities.json?iss.only=marketdata&sort_column=UPDATETIME&sort_order=desc&first=1&marketdata.columns=UPDATETIME"
		if market == "index" {
			parturl += "&securities=MICEXINDEXCF,RTSI,MICEXBMI,RTSSTD,RVI"
		}
	} else if typeOfCheck == "trades" {
		parturl = "/trades.json?reversed=1&limit=1&iss.only=trades&trades.columns=TRADETIME"
	} else {
		log.Fatal("unknown type of check")
		return "unknown type of check"
	}
	return "http://iss.moex.com/iss/engines/" + engine + "/markets/" + market + parturl + randNum()
}

// TODO split on 2 functions
func getURL(url string) string {
	var input Request
	var output string

	// log.Println(url)
	json.Unmarshal(moexlib.GetAllContents(url), &input)
	if input.Marketdata.Columns == nil {
		json.Unmarshal(moexlib.GetAllContents(url), &input)
		output = input.Trade.Data[0][0]
		return output
	}
	output = input.Marketdata.Data[0][0]
	return output

}

func main() {
	urls := map[string]string{
		"shares": "stock",
		"selt":   "currency",
		"forts":  "futures",
		"index":  "stock",
	}

	configuration = config.ReadConfig("config.json")
	checks := []string{"marketdata", "trades"}
	for _, typeOfCheck := range checks {
		for market, engine := range urls {
			url := urlReturn(engine, market, typeOfCheck)
			diff := moexlib.GetDelta(getURL(url))
			delta := fmt.Sprintf("%v", diff)
			// fmt.Println(engine+"--"+market, delta)
			ok := moexlib.Send2Graphite(delta, "iss."+typeOfCheck+"."+engine+"."+market, configuration.Server.IP, configuration.Server.Port)
			if ok == false {
				log.Fatal(ok)
			}

		}
	}

}
