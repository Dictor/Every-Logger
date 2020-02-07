package main

import (
	"encoding/json"
	//"github.com/PuerkitoBio/goquery"
	ws "github.com/dictor/wswrapper"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type topicData struct {
	Time  int
	Value float64
}

var topicValue map[string]*topicData
var clientTopic map[*ws.WebsocketClient]string
var topicDetail map[string]interface{}

func InitFetchTopic() {
	topicValue = make(map[string]*topicData)
	clientTopic = make(map[*ws.WebsocketClient]string)

	go FetchRandom("test")
	go FetchJson("btcusd", "https://api.cryptowat.ch/markets/bitfinex/btcusd/price", func(data map[string]interface{}) (string, bool) {
		price, ok := (data["result"].(map[string]interface{}))["price"].(float64)
		if !ok {
			return "", false
		} else {
			return strconv.FormatFloat(price, 'f', -1, 64), true
		}
	})
	go FetchJson("2019ncov-w", "https://wuhanvirus.kr/stat.json", func(data map[string]interface{}) (string, bool) {
		idata, ok := (data["chart"].(map[string]interface{}))["global"].([]interface{})
		if !ok {
			return "retrieve idata fail", false
		} else {
			pdata, ok := idata[len(idata)-1].([]interface{})[1].(float64)
			if !ok {
				return "retrieve pdata fail", false
			} else {
				return strconv.FormatFloat(pdata, 'f', -1, 64), true
			}
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

func newTopicData(val float64) *topicData {
	return &topicData{int(time.Now().Unix()), val}
}
