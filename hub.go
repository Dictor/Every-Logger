// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
)

type RecieveEvent struct {
	client *Client
	msg    *[]byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	recieve chan *RecieveEvent

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		recieve:    make(chan *RecieveEvent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run(recv_cb func(*RecieveEvent)) {
	for {
		select {
		case client := <-h.register:
			log.Println("[WS REG]", makeWsPrefix(client))
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Println("[WS UNREG]", makeWsPrefix(client))
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
					log.Printf("[WS BROAD]\n")
				default: // when client.send closed, close and delete client
					log.Println("[WS UNREG]", makeWsPrefix(client))
					close(client.send)
					delete(h.clients, client)
				}
			}
		case recv := <-h.recieve:
			recv_cb(recv)
		}
	}
}
