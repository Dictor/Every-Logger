package main

import (
	//"github.com/PuerkitoBio/goquery"
	ws "github.com/dictor/wswrapper"
	"log"
	"strings"
	"time"
)

type topicDataAdd struct {
	Name string
	Data *topicData
}

type topicData struct {
	Time  int
	Value float64
}

var (
	topicValue     map[string]*topicData
	topicSafeAdder chan *topicDataAdd
	topicDetail    map[string]interface{}
	clientTopic    map[*ws.WebsocketClient]string
)

func InitFetchTopic(root_dir string, is_chrome_log bool) {
	topicValue = make(map[string]*topicData)
	clientTopic = make(map[*ws.WebsocketClient]string)
	go TopicSafeAdder()
	/*
		go FetchRandom("test")
		go FetchFile("test-file", root_dir+"/db/test-file.txt", FetchStringStdCb)
		go FetchJson("btcusd", "https://api.cryptowat.ch/markets/bitfinex/btcusd/price", func(data map[string]interface{}) (float64, bool) {
			price, ok := (data["result"].(map[string]interface{}))["price"].(float64)
			if !ok {
				return 0.0, false
			} else {
				return price, true
			}
		})
	*/
	AddFetchChromeTopic("co19-cn-cur", "https://ncov.dxy.cn/ncovh5/view/pneumonia", ".count___3GCdh > li:nth-child(1) > strong", FetchStringStdCb)
	//go FetchChrome("co19-kr-all", "https://coronamap.site/", "div.wa > .content > div", FetchStringStdCb)
	AddFetchChromeTopic("co19-kr-all", "http://ncov.mohw.go.kr/index_main.jsp", "div.co_cur >ul > li:first-child > a", func(v string) (float64, bool) {
		return FetchStringStdCb(strings.Replace(v, " 명", "", -1))
	})
	StartFetchChrome(root_dir, is_chrome_log)
}

func BindTopicInfo(root_dir string) {
	prepareDirectory(root_dir + "/db")
	topicDetail = map[string]interface{}{}
	BindFileToJson(root_dir+"/db/topic_detail.json", &topicDetail)
}

func BindLatestValue() {
	cnt := 0
	for name, _ := range topicDetail {
		val, err := GetLatestValue(name)
		if err != nil {
			log.Printf("[BindLatestValue] Latest value binding failure : %s", err)
		} else {
			topicValue[name] = val
			cnt++
		}
	}
	log.Printf("[BindLatestValue] Binding %d topics latest value", cnt)
}

func TopicSafeAdder() {
	topicSafeAdder = make(chan *topicDataAdd, 10)
	for {
		select {
		case val := <-topicSafeAdder:
			topicValue[val.Name] = val.Data
		}
	}
}

func UpdateTopicValue(val *topicDataAdd) {
	topicSafeAdder <- val
}

func newTopicData(val float64) *topicData {
	return &topicData{int(time.Now().Unix()), val}
}
