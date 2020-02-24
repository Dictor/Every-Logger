package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

var (
	FetchChromeTempDir   string
	FetchChromeRootCtx   context.Context
	FetchChromeLogEnable bool
	FetchChromeTopics    []*fetchChromeParam
)

type fetchChromeParam struct {
	TopicName string
	Url       string
	Selector  string
	Callback  FetchStringCallback
}

func InitFetchChrome(root_dir string, log_enable bool) {
	FetchChromeTempDir = root_dir + "/chrome_temp"
	FetchChromeLogEnable = log_enable
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(FetchChromeTempDir),
	)
	FetchChromeRootCtx, _ = chromedp.NewExecAllocator(
		context.Background(),
		opts...,
	)
	FetchChromeTopics = make([]*fetchChromeParam, 0)
	log.Printf("[InitFetchChrome] Initialized chrome temp directory is : %s\n", FetchChromeTempDir)
}

func AddFetchChromeTopic(topic_name string, url string, selector string, callback FetchStringCallback) {
	FetchChromeTopics = append(FetchChromeTopics, &fetchChromeParam{topic_name, url, selector, callback})
}

func detailChromeLog(format string, v ...interface{}) {
	if FetchChromeLogEnable {
		log.Printf(format, v...)
	}
}

func FetchChrome(params []*fetchChromeParam) {
	var (
		res string
		ctx context.Context
	)
	var clean_loop = func() {
		if ctx != nil {
			err := chromedp.Cancel(ctx)
			if err != nil {
				log.Printf("[FetchChrome(%p)] Context closing failure (%s)\n", ctx, err)
			}
			InterruptCounter.Done()
			detailChromeLog("[FetchChrome(%p)] Context closed\n", ctx)
		}
	}
	defer clean_loop()

	for {
		clean_loop()
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)

		ctx, _ = chromedp.NewContext(
			FetchChromeRootCtx,
			chromedp.WithLogf(log.Printf),
		)
		InterruptCounter.Add(1)
		chromedp.ListenBrowser(ctx, func(ev interface{}) {
			//detailChromeLog("[FetchChrome(%p)] Browser event : %+v\n", ctx, ev)
		})
		detailChromeLog("[FetchChrome(%p)] Context opened\n", ctx)
		detailChromeLog("[FetchChrome(%p)] Start fetching with %d topics\n", ctx, len(params))

		for _, param := range params {
			detailChromeLog("[FetchChrome(%p)][%s] Fetching topic start\n", ctx, param.TopicName)
			err := chromedp.Run(ctx,
				chromedp.Navigate(param.Url),
				chromedp.Text(param.Selector, &res),
			)
			if err != nil {
				log.Printf("[FetchChrome(%p)][%s] Running chrome failure (%s)\n", ctx, param.TopicName, err)
				continue
			}

			detailChromeLog("[FetchChrome(%p)][%s] Fetching topic raw data = '%s'", ctx, param.TopicName, res)
			cres, csucc := param.Callback(res)
			if !csucc {
				log.Printf("[FetchChrome(%p)][%s] Process callback failure (%s)\n", ctx, param.TopicName, cres)
				continue
			}

			tdata := newTopicData(cres)
			AddValue(param.TopicName, tdata)
			UpdateTopicValue(&topicDataAdd{param.TopicName, tdata})
			detailChromeLog("[FetchChrome(%p)][%s] Fetching topic complete", ctx, param.TopicName)
		}

		select {
		case <-InterruptNotice:
			return
		default:
		}
	}
}
