// Original Code by : Copyright C) 2013 The Gorilla WebSocket Authors. All rights reserved.
// Modified by Dictor(kimdictor@gmail.com)

package main

type WebsocketEventKind int

const (
	EVENT_RECIEVE WebsocketEventKind = iota
	EVENT_REGISTER
	EVENT_UNREGISTER
	EVENT_BROADCAST
)

type WebsocketEvent struct {
	kind   WebsocketEventKind
	client *Client
	msg    *[]byte
}

type WebsocketHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	recieve    chan *WebsocketEvent
	register   chan *Client
	unregister chan *Client
}

func newWebsocketHub() *WebsocketHub {
	return &WebsocketHub{
		broadcast:  make(chan []byte),
		recieve:    make(chan *WebsocketEvent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
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

func (h *WebsocketHub) closeClient(cli *Client) {
	close(cli.send)
	delete(h.clients, cli)
}

func (h *WebsocketHub) sendSafe(cli *Client, msg *[]byte) bool {
	select {
	case cli.send <- *msg:
		return true
	default: // when client.send closed, close and delete client
		h.closeClient(cli)
		return false
	}
}
