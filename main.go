package main

import (
	"encoding/json"
	"flag"
	"github.com/dictor/justlog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

var sendPeriod, dataPeriod int
var config map[string]interface{}

func main() {
	// Setting log and os signal handler
	attachInterruptHandler()
	log_path := justlog.MustPath(justlog.SetPath())
	defer (justlog.MustStream(justlog.SetStream(log_path))).Close()

	// Bind CLI parameters
	var (
		faddr string
		fclog bool
	)
	flag.StringVar(&faddr, "addr", ":80", "Server address")
	flag.IntVar(&sendPeriod, "sp", 2500, "Websocket sending term")
	flag.IntVar(&dataPeriod, "fp", 10000, "Fetching data term")
	flag.BoolVar(&fclog, "chrome-log", false, "Enable detail chrome log")
	flag.Parse()

	// Initiating topic data
	BindTopicInfo(justlog.ExePath)
	OpenDB(justlog.ExePath)
	InitFetchTopic(justlog.ExePath, fclog)
	BindLatestValue()

	// Initiation echo server
	main_server := echo.New()
	request_count := 0
	main_server.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			request_count++
			return strconv.Itoa(request_count)
		},
	}))
	main_server.HTTPErrorHandler = func(err error, cxt echo.Context) {
		log.Println(makeEchoPrefix(cxt, "HTTP_ERROR"), err)
		main_server.DefaultHTTPErrorHandler(err, cxt)
	}

	// Start echo server
	SetRouting(main_server, wsInit(justlog.ExePath))
	log.Println("[SERVER START]")
	log.Fatal("[SERVER TERMINATED] ", main_server.Start(faddr))
}

func attachInterruptHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[InterruptDetector] Terminal interrupt detected!")
		CloseDB()
		close(InterruptNotice)
		log.Println("[InterruptDetector] Waiting until all fetch routine is closed...")

		c := make(chan struct{})
		go func() {
			defer close(c)
			for {
				var (
					t = atomic.LoadInt32(&totalTaskCount)
					f = atomic.LoadInt32(&finishedTaskCount)
					d = atomic.LoadInt32(&disclaimTaskCount)
				)
				if t == f+d {
					return
				}
				time.Sleep(time.Second)
			}
		}()
		select {
		case <-c:
		case <-time.After(30 * time.Second):
			log.Println("[InterruptDetector] Waiting timeout! Forced finish without gracefully.")
		}
		log.Println("[InterruptDetector] Closing process is finished! Goodbye!")
		os.Exit(0)
	}()
}

func makeEchoPrefix(cxt echo.Context, func_name string) string {
	id := cxt.Request().Header.Get(echo.HeaderXRequestID)
	if id == "" {
		id = cxt.Response().Header().Get(echo.HeaderXRequestID)
	}
	var params = [...]string{func_name, id, cxt.RealIP(), cxt.Request().URL.String()}
	var result string
	for _, val := range params {
		result += "[" + val + "]"
	}
	return result
}

func BindFileToJson(file_path string, data *map[string]interface{}) {
	rawjson, err := ioutil.ReadFile(file_path)
	if err != nil {
		log.Panic(err)
	}
	json.Unmarshal(rawjson, &data)
}
