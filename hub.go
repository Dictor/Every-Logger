// Original Code by : Copyright C) 2013 The Gorilla WebSocket Authors. All rights reserved.
// Modified by Dictor(kimdictor@gmail.com)

package main

import (
	"net/http"
)

type WebsocketEventKind int

const (
	EVENT_RECIEVE WebsocketEventKind = iota
	EVENT_REGISTER
	EVENT_UNREGISTER
	EVENT_BROADCAST
)

type WebsocketEvent struct {
	kind   WebsocketEventKind
	client *WebsocketClient
	msg    *[]byte
}

type WebsocketHub struct {
	clients    map[*WebsocketClient]bool
	broadcast  chan []byte
	recieve    chan *WebsocketEvent
	register   chan *WebsocketClient
	unregister chan *WebsocketClient
	err        chan *WebsocketEvent
}

func newWebsocketHub() *WebsocketHub {
	return &WebsocketHub{
		broadcast:  make(chan []byte),
		recieve:    make(chan *WebsocketEvent),
		register:   make(chan *WebsocketClient),
		unregister: make(chan *WebsocketClient),
		clients:    make(map[*WebsocketClient]bool),
		err:        make(chan *WebsocketEvent),
	}
}

func (h *WebsocketHub) run(event_callback func(*WebsocketEvent)) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			event_callback(&WebsocketEvent{EVENT_REGISTER, client, nil})
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeClient(client)
				event_callback(&WebsocketEvent{EVENT_UNREGISTER, client, nil})
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if h.sendSafe(client, &message) {
					event_callback(&WebsocketEvent{EVENT_BROADCAST, client, nil})
				} else {
					event_callback(&WebsocketEvent{EVENT_UNREGISTER, client, nil})
				}
			}
		case evt := <-h.recieve:
			event_callback(evt)
		}
	}
}

func (h *WebsocketHub) closeClient(cli *WebsocketClient) {
	close(cli.send)
	delete(h.clients, cli)
}

func (h *WebsocketHub) sendSafe(cli *WebsocketClient, msg *[]byte) bool {
	select {
	case cli.send <- *msg:
		return true
	default: // when client.send closed, close and delete client
		h.closeClient(cli)
		return false
	}
}

func (h *WebsocketHub) addClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &WebsocketClient{hub: h, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
