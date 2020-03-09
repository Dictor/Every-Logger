package main

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"log"
	"sync/atomic"
	"time"
)

var (
	fetchChromeTempDir                                   string
	fetchChromeRootCtx                                   context.Context
	fetchChromeLogEnable                                 bool
	fetchChromeTopics                                    []*fetchChromeParam = make([]*fetchChromeParam, 0)
	totalTaskCount, finishedTaskCount, disclaimTaskCount int32
)

type fetchChromeParam struct {
	TopicName string
	Url       string
	Selector  string
	Callback  FetchStringCallback
}

type ChromeTaskResult struct {
	Success bool
	Error   error
	Id      int32
	Value   *ChromeTaskValue
}

type ChromeTaskValue struct {
	TopicValue *topicData
	TopicName  string
}

func AddFetchChromeTopic(topic_name string, url string, selector string, callback FetchStringCallback) {
	fetchChromeTopics = append(fetchChromeTopics, &fetchChromeParam{topic_name, url, selector, callback})
}

func StartFetchChrome(root_dir string, log_enable bool) {
	fetchChromeTempDir = root_dir + "/chrome_temp"
	fetchChromeLogEnable = log_enable
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(fetchChromeTempDir),
	)
	fetchChromeRootCtx, _ = chromedp.NewExecAllocator(
		context.Background(),
		opts...,
	)
	log.Printf("[InitFetchChrome] Initialized chrome temp directory is : %s\n", fetchChromeTempDir)
	go fetchChrome(fetchChromeTopics)
}

func detailChromeLog(format string, v ...interface{}) {
	if fetchChromeLogEnable {
		log.Printf(format, v...)
	}
}

func fetchChrome(params []*fetchChromeParam) {
	result := make(chan *ChromeTaskResult, 10)
	for {
		select {
		case res := <-result:
			if res.Success {
				AddValue(res.Value.TopicName, res.Value.TopicValue)
				UpdateTopicValue(&topicDataAdd{res.Value.TopicName, res.Value.TopicValue})
				log.Printf("[FetchChrome] Task(%d) success: %s", res.Id, res.Value.TopicName)
			} else {
				log.Printf("[FetchChrome] Task(%d) error: %s", res.Id, res.Error)
			}
		case <-time.After(time.Duration(dataPeriod) * time.Millisecond * 2):
			go startChromeTask(result, params, time.Duration(dataPeriod)*time.Millisecond)
			t := atomic.LoadInt32(&totalTaskCount)
			f := atomic.LoadInt32(&finishedTaskCount)
			d := atomic.LoadInt32(&disclaimTaskCount)
			log.Printf("[FetchChrome] Task count: (total %d) = (finished %d) + (disclaimed %d) + (running %d)", t, f, d, t-f-d)
		case <-InterruptNotice:
			return
		}
	}
}

func startChromeTask(result chan<- *ChromeTaskResult, params []*fetchChromeParam, timeout time.Duration) {
	cres := make(chan *ChromeTaskResult, 2)
	cid := atomic.AddInt32(&totalTaskCount, 1)
	var close_ctx func()

	go func() {
		var (
			pctx context.Context
			sres string
		)
		pctx, close_ctx = context.WithTimeout(fetchChromeRootCtx, timeout)
		ctx, _ := chromedp.NewContext(pctx, chromedp.WithLogf(log.Printf))
		for _, param := range params {
			err := chromedp.Run(ctx,
				chromedp.Navigate(param.Url),
				chromedp.Text(param.Selector, &sres, chromedp.AtLeast(0)))
			if err != nil {
				cres <- &ChromeTaskResult{false, err, cid, nil}
			} else {
				cbres, cbsucc := param.Callback(sres)
				if !cbsucc {
					cres <- &ChromeTaskResult{false, errors.New("Callback function returned error."), cid, nil}
				} else {
					cres <- &ChromeTaskResult{true, nil, cid, &ChromeTaskValue{newTopicData(cbres), param.TopicName}}
				}
			}
		}
		chromedp.Cancel(ctx)
		close(cres)
	}()

	for {
		select {
		case <-time.After(timeout * 2):
			atomic.AddInt32(&disclaimTaskCount, 1)
			close_ctx()
			result <- &ChromeTaskResult{false, errors.New("Task timeout! Disclaim current task routine."), cid, nil}
			return
		case res, open := <-cres:
			if open {
				result <- res
			} else {
				atomic.AddInt32(&finishedTaskCount, 1)
				return
			}
		}
	}
}
