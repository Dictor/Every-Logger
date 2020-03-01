package main

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var (
	InterruptNotice chan bool = make(chan bool)
)

type FetchStringCallback func(string) (float64, bool)

func newGoqDoc(html_path string) (*goquery.Document, bool) {
	s, succ := getHtml(html_path)
	if !succ {
		return nil, false
	}

	doc, err := goquery.NewDocumentFromReader(s)
	if err != nil {
		log.Printf("[FetchHtml Error][%s] %s\n", html_path, err)
		return nil, false
	}

	return doc, true
}

func FetchHtml(topic_name string, html_path string, f func(*goquery.Document) string) {
	for {
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
		doc, succ := newGoqDoc(html_path)
		if !succ {
			continue
		}

		hres := f(doc)
		fres, err := strconv.ParseFloat(hres, 64)
		if err != nil {
			log.Printf("[fetchHtml][%s] '%s' → float64 : %s \n", html_path, hres, err)
			continue
		}

		UpdateTopicValue(&topicDataAdd{topic_name, newTopicData(fres)})
	}
}

func FetchJson(topic_name string, html_path string, process_callback func(map[string]interface{}) (float64, bool)) {
	for {
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
		hres, succ := getHtml(html_path)
		if !succ {
			log.Printf("[FetchJson][%s] Get html document failure (%s)\n", html_path, hres)
			continue
		}
		rawjson := []byte(streamToString(hres))
		var resjson map[string]interface{}
		json.Unmarshal(rawjson, &resjson)

		cres, csucc := process_callback(resjson)
		if !csucc {
			log.Printf("[FetchJson][%s] Process callback failure (%s)\n", html_path, cres)
			continue
		}

		tdata := newTopicData(cres)
		AddValue(topic_name, tdata)
		UpdateTopicValue(&topicDataAdd{topic_name, tdata})
	}
}

func FetchFile(topic_name string, file_path string, process_callback func(val string) (float64, bool)) {
	for {
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
		fdata, err := ioutil.ReadFile(file_path)
		if err != nil {
			log.Printf("[FetchFile][%s] Read file failure (%s)\n", file_path, err)
			continue
		}

		cres, csucc := process_callback(string(fdata))
		if !csucc {
			log.Printf("[FetchFile][%s] Process callback failure (%s)\n", file_path, cres)
			continue
		}

		tdata := newTopicData(cres)
		AddValue(topic_name, tdata)
		UpdateTopicValue(&topicDataAdd{topic_name, tdata})
	}
}

func FetchRandom(topic_name string) {
	for {
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)

		val, ok := topicValue[topic_name]
		var ival float64
		if ok {
			ival = val.Value
		} else {
			ival = 0
		}

		tdata := newTopicData(ival + rand.Float64()*25 - 10)
		AddValue(topic_name, tdata)
		UpdateTopicValue(&topicDataAdd{topic_name, tdata})
	}
}

func streamToString(s io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(s)
	str := buf.String()
	return str
}

func getHtml(html_path string) (io.ReadCloser, bool) {
	res, err := http.Get(html_path)
	if err != nil {
		log.Printf("[FetchHtml][%s] %s\n", html_path, err)
		return nil, false
	}
	if res.StatusCode != 200 {
		log.Printf("[FetchHtml][%s] %s (%d)\n", html_path, res.Status, res.StatusCode)
		return nil, false
	}
	return res.Body, true
}
