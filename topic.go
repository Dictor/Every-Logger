package main

import (
	//"github.com/PuerkitoBio/goquery"
	ws "github.com/dictor/wswrapper"
	"strconv"
	"strings"
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
	go FetchJson("btcusd", "https://api.cryptowat.ch/markets/bitfinex/btcusd/price", func(data map[string]interface{}) (float64, bool) {
		price, ok := (data["result"].(map[string]interface{}))["price"].(float64)
		if !ok {
			return 0.0, false
		} else {
			return price, true
		}
	})
	go FetchChrome("2019ncov-w", "https://ncov.dxy.cn/ncovh5/view/pneumonia", ".count___3GCdh > li:nth-child(1) > strong", func(val string) (float64, bool) {
		ival, err := strconv.Atoi(strings.Replace(val, ",", "", 1))
		if err != nil {
			return 0.0, false
		} else {
			return float64(ival), true
		}
	})
}

func BindTopicInfo(root_dir string) {
	prepareDirectory(root_dir + "/db")
	topicDetail = map[string]interface{}{}
	BindFileToJson(root_dir+"/db/topic_detail.json", &topicDetail)
}

func newTopicData(val float64) *topicData {
	return &topicData{int(time.Now().Unix()), val}
}
