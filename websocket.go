package main

import (
	"fmt"
	ws "github.com/dictor/wswrapper"
	"log"
	"strings"
	"time"
)

func wsInit(root_dir string) *ws.WebsocketHub {
	hub := ws.NewHub()
	go hub.Run(wsEvent)
	go PublishValue(hub)
	config := map[string]interface{}{}
	BindFileToJson(root_dir+"/config.json", &config)
	ws_origin := config["ws_origin"].([]interface{})
	sws_origin := []string{}
	for _, val := range ws_origin {
		sws_origin = append(sws_origin, val.(string))
	}
	log.Printf("[wsInit] %d origins added to websocket upgrader", len(sws_origin))
	hub.AddUpgraderOrigin(sws_origin)
	return hub
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
			} else {
				log.Printf("[PublishValue] Publishing Error : get cli topic = %t, get topic value = %t", okt, ok)
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
