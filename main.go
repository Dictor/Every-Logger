package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dictor/justlog"
	ws "github.com/dictor/wswrapper"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var sendPeriod, dataPeriod int
var config map[string]interface{}

func main() {
	attachInterruptHandler()
	log_path := justlog.MustPath(justlog.SetPath())
	defer (justlog.MustStream(justlog.SetStream(log_path))).Close()

	BindTopicInfo(justlog.ExePath)
	OpenDB(justlog.ExePath)
	InitFetchTopic()

	hub := ws.NewHub()
	go hub.Run(wsEvent)
	go PublishValue(hub)

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

	var addr string
	flag.StringVar(&addr, "addr", ":80", "Server address")
	flag.IntVar(&sendPeriod, "sp", 2500, "Websocket sending term")
	flag.IntVar(&dataPeriod, "fp", 10000, "Fetching data term")
	flag.Parse()

	config := map[string]interface{}{}
	BindFileToJson(justlog.ExePath+"/config.json", &config)
	ws_origin := config["ws_origin"].([]interface{})
	sws_origin := []string{}
	for _, val := range ws_origin {
		sws_origin = append(sws_origin, val.(string))
	}
	log.Printf("%d origins added to websocket upgrader", len(sws_origin))
	hub.AddUpgraderOrigin(sws_origin)

	SetRouting(main_server, hub)
	log.Println("[SERVER START]")
	log.Fatal("[SERVER TERMINATED] ", main_server.Start(addr))
}

func attachInterruptHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[InterruptDetector] Terminal interrupt detected!")
		close(InterruptNotice)
		CloseDB()
		log.Println("[InterruptDetector] Waiting until all fetch routine is closed...")
		InterruptCounter.Wait()
		log.Println("[InterruptDetector] Closing process is finished! Goodbye!")
		os.Exit(0)
	}()
}

func wsEvent(evt *ws.WebsocketEvent) {
	switch evt.Kind {
	case ws.EVENT_RECIEVE:
		str := string(*evt.Msg)
		pstr := strings.Split(str, ",")
		switch len(pstr) {
		case 2:
			switch pstr[0] {
			case "TOPIC":
				detail, ok := topicDetail[pstr[1]]
				if ok {
					detailval := detail.(map[string]interface{})
					evt.Client.Hub().Send(evt.Client, []byte("TOPIC,"+detailval["name"].(string)+","+detailval["detail"].(string)))
					log.Printf("[TOPIC CHANGE]%s : %s → %s", makeWsPrefix(evt.Client), clientTopic[evt.Client], pstr[1])
					clientTopic[evt.Client] = pstr[1]
				} else {
					evt.Client.Hub().Send(evt.Client, []byte("ERROR,NOTOPIC"))
					log.Printf("[TOPIC CHANGE]%s : %s → %s : No topic", makeWsPrefix(evt.Client), clientTopic[evt.Client], pstr[1])
				}
			}
		}
	case ws.EVENT_REGISTER:
		log.Printf("[WS_REG]%s", makeWsPrefix(evt.Client))
	case ws.EVENT_UNREGISTER:
		log.Printf("[WS_UNREG]%s", makeWsPrefix(evt.Client))
	case ws.EVENT_ERROR:
		log.Printf("[WS_ERROR]%s %s", makeWsPrefix(evt.Client), evt.Err)
	}
}

func PublishValue(h *ws.WebsocketHub) {
	for {
		for cli, _ := range h.Clients() {
			valt, okt := clientTopic[cli]
			val, ok := topicValue[valt]
			if ok && okt {
				h.Send(cli, []byte(fmt.Sprintf("VALUE,%s,%d,%f", valt, val.Time, val.Value)))
			}
		}
		time.Sleep(time.Duration(sendPeriod) * time.Millisecond)
	}
}

func makeWsPrefix(cli *ws.WebsocketClient) string {
	if cli == nil {
		return "[?]"
	}
	return fmt.Sprintf("[%s][%d]", cli.Connection().RemoteAddr(), cli.Hub().Clients()[cli])
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
