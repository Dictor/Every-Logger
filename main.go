package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dictor/justlog"
	ws "github.com/dictor/wswrapper"
	"log"
	"net/http"
	"strings"
	"time"
)

var sendPeriod, dataPeriod int

func main() {
	log_path := justlog.MustPath(justlog.SetPath())
	defer (justlog.MustStream(justlog.SetStream(log_path))).Close()

	hub := ws.NewHub()
	go hub.Run(wsEvent)
	go sendInfo(hub)

	OpenDB(justlog.ExePath)
	BindTopicInfo(justlog.ExePath)
	InitFetchTopic()

	staticFs := http.FileServer(http.Dir("front"))
	http.Handle("/", staticFs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.AddClient(w, r)
	})
	http.HandleFunc("/ival", func(w http.ResponseWriter, r *http.Request) {
		topic_name, ok := r.URL.Query()["topic"]
		if !ok {
			return
		}
		_, topic_vaild := topicDetail[topic_name[0]]
		if !topic_vaild {
			return
		}

		term, ok := r.URL.Query()["term"]
		if !ok {
			return
		}

		ivalue := []interface{}{}
		res, err := GetTopicData(topic_name[0], term[0])
		if err == nil {
			for _, val := range res {
				nowvalue := []interface{}{val.Time * 1000, val.Value}
				ivalue = append(ivalue, nowvalue)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ivalue)
		}
	})

	log.Println("[SERVER START]")
	var addr string
	flag.StringVar(&addr, "addr", ":80", "Server address")
	flag.IntVar(&sendPeriod, "sp", 2500, "Websocket sending term")
	flag.IntVar(&dataPeriod, "fp", 10000, "Fetching data term")
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("[SERVER ERROR] ", err)
	}
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

func sendInfo(h *ws.WebsocketHub) {
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
	return fmt.Sprintf("[%s][%d]", cli.Connection().RemoteAddr(), cli.Hub().Clients()[cli])
}
