package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	InterruptNotice  chan bool      = make(chan bool)
	InterruptCounter sync.WaitGroup = sync.WaitGroup{}
)

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
		topicValue[topic_name] = newTopicData(fres)
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
		topicValue[topic_name] = tdata
	}
}

func FetchChrome(topic_name string, url string, selector string, process_callback func(val string) (float64, bool)) {
	InterruptCounter.Add(1)
	ctx, close_ctx := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer func() {
		close_ctx()
		InterruptCounter.Done()
		log.Printf("[FetchChrome] Chrome context %p is closed!", ctx)
	}()

	var res string
FETCH_LOOP:
	for {
		select {
		case <-InterruptNotice:
			break FETCH_LOOP
		default:
		}

		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			//chromedp.WaitVisible(selector),
			chromedp.Sleep(time.Second*2),
			chromedp.Text(selector, &res),
		)
		if err != nil {
			log.Printf("[FetchChrome][%s] Running chrome failure (%s)\n", url, err)
			continue
		}

		cres, csucc := process_callback(res)
		if !csucc {
			log.Printf("[FetchChrome][%s] Process callback failure (%s)\n", url, cres)
			continue
		}

		tdata := newTopicData(cres)
		AddValue(topic_name, tdata)
		topicValue[topic_name] = tdata
	}
}

func FetchRandom(topic_name string) {
	for {
		val, ok := topicValue[topic_name]
		var ival float64
		if ok {
			ival = val.Value
		} else {
			ival = 0
		}

		topicValue[topic_name] = newTopicData(ival + rand.Float64()*25 - 10)
		AddValue(topic_name, topicValue[topic_name])
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
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
