package main

import (
	"encoding/json"
	ws "github.com/dictor/wswrapper"
	"io/ioutil"
	"log"
	"strconv"
)

var topicValue map[string]float64
var clientTopic map[*ws.WebsocketClient]string
var topicDetail map[string]interface{}

func InitFetchTopic() {
	topicValue = make(map[string]float64)
	clientTopic = make(map[*ws.WebsocketClient]string)

	go makeDummyData("test")
	go FetchJson("btcusd", "https://api.cryptowat.ch/markets/bitfinex/btcusd/price", func(data map[string]interface{}) (string, bool) {
		price, ok := (data["result"].(map[string]interface{}))["price"].(float64)
		if !ok {
			return "", false
		} else {
			AddTopicData("btcusd", price)
			return strconv.FormatFloat(price, 'f', -1, 64), true
		}
	})
}

func BindTopicInfo(root_dir string) {
	prepareDirectory(root_dir + "/db")
	rawjson, err := ioutil.ReadFile(root_dir + "/db/topic_detail.json")
	if err != nil {
		log.Panic(err)
	}
	json.Unmarshal(rawjson, &topicDetail)
}
