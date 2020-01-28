package main

import (
	"flag"
	"fmt"
	"github.com/dictor/justlog"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var clientTopic map[*WebsocketClient]string
var topicValue map[string]string
var sendPeriod, dataPeriod, fakedata int

func main() {
	log_path := justlog.MustPath(justlog.SetPath())
	defer (justlog.MustStream(justlog.SetStream(log_path))).Close()

	hub := newWebsocketHub()
	go hub.run(wsRecieve)
	go sendInfo(hub)

	topicValue = make(map[string]string)
	clientTopic = make(map[*WebsocketClient]string)
	fakedata = rand.Intn(100)
	go makeFakeData()

	staticFs := http.FileServer(http.Dir("front"))
	http.Handle("/", staticFs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.addClient(w, r)
	})

	log.Println("[SERVER START]")

	var addr string

	flag.StringVar(&addr, "addr", ":80", "Server address")
	flag.IntVar(&sendPeriod, "sp", 500, "Websocket sending term")
	flag.IntVar(&dataPeriod, "fp", 100, "Fetching data term")
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("[SERVER ERROR]", err)
	}
}

func wsRecieve(evt *WebsocketEvent) {
	str := string(*evt.msg)
	pstr := strings.Split(str, ",")
	switch len(pstr) {
	case 2:
		switch pstr[0] {
		case "TOPIC":
			log.Printf("[TOPIC CHANGE]%s : %s → %s", makeWsPrefix(evt.client), clientTopic[evt.client], pstr[1])
			clientTopic[evt.client] = pstr[1]
		}
	}
}

func sendInfo(h *WebsocketHub) {
	for {
		for cli, _ := range h.clients {
			select {
			case cli.send <- []byte(fmt.Sprintf("VALUE,%s,%s", clientTopic[cli], topicValue[clientTopic[cli]])):
			default: // when client.send closed, close and delete client
				log.Println("[WS UNREG]", makeWsPrefix(cli))
				close(cli.send)
				delete(h.clients, cli)
			}
		}
		time.Sleep(time.Duration(sendPeriod) * time.Millisecond)
	}
}

func makeFakeData() {
	for {
		fakedata += (rand.Intn(20) - 10)
		topicValue["test"] = strconv.Itoa(fakedata)
		time.Sleep(time.Duration(dataPeriod) * time.Millisecond)
	}
}

func makeWsPrefix(cli *WebsocketClient) string {
	return fmt.Sprintf("[%s][%p]", cli.conn.RemoteAddr(), &cli)
}
