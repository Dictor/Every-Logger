package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func newGoqDoc(html_path string) (*goquery.Document, bool) {
	s, succ := fetchHtml(html_path)
	if !succ {
		return nil, false
	}

	doc, err := goquery.NewDocumentFromReader(s)
	if err != nil {
		log.Printf("[FetchHtml Error][%s] %s\n", html_path, err)
		return nil, false
	}

	debug, _ := os.Create("debug_html.txt")
	defer debug.Close()
	hres, _ := doc.Html()
	fmt.Fprint(debug, hres)
	return doc, true
}

/*
func fetchHtml(topic_name string, html_path string, f func(*goquery.Document) string) {
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
		topicValue[topic_name] = fres
	}
}
*/

func FetchJson(topic_name string, html_path string, process_callback func(map[string]interface{}) (string, bool)) {
	for {
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
		hres, succ := fetchHtml(html_path)
		if !succ {
			continue
		}
		rawjson := []byte(streamToString(hres))
		var resjson map[string]interface{}
		json.Unmarshal(rawjson, &resjson)

		cres, csucc := process_callback(resjson)
		if !csucc {
			log.Printf("[FetchJson][%s] Process callback failure\n", html_path)
			continue
		}

		fres, err := strconv.ParseFloat(cres, 64)
		if err != nil {
			log.Printf("[FetchJson][%s] '%s' → float64 : %s \n", html_path, hres, err)
			continue
		}
		topicValue[topic_name] = fres
	}
}

func fetchHtml(html_path string) (io.ReadCloser, bool) {
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

func streamToString(s io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(s)
	str := buf.String()
	return str
}

func makeDummyData(topic_name string) {
	for {
		topicValue[topic_name] += rand.Float64()*25 - 10
		AddTopicData(topic_name, topicValue[topic_name])
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
	}
}
