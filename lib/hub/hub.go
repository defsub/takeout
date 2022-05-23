// Copyright (C) 2022 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package hub

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/defsub/takeout/lib/log"
	"net"
	"net/http"
	"time"
	"errors"
	"os"
	"strings"
)

type Authenticator interface {
	Authenticate(string) bool
}

type Message struct {
	sender *Client
	body   []byte
}

type Hub struct {
	nextId     int64
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

type Conn net.Conn

type Client struct {
	id   int64
	hub  *Hub
	conn Conn
	send chan Message
}

func NewHub() *Hub {
	return &Hub{
		nextId:     1,
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) done(client *Client) {
	delete(h.clients, client)
	close(client.send)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			client.id = h.nextId
			h.nextId++
			log.Printf("register: clients %d\n", len(h.clients))
			for k := range h.clients {
				log.Printf("(reg) client: %d\n", k.id)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.done(client)
			}
			log.Printf("unregister: clients %d\n", len(h.clients))
		case message := <-h.broadcast:
			for client := range h.clients {
				if client == message.sender {
					// don't send to self
					continue
				}
				select {
				case client.send <- message:
				default:
					h.done(client)
				}
			}
		}
	}
}

func (h *Hub) Handle(auth Authenticator, w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Println(err)
		return
	}

	c := &Client{
		id:   0,
		hub:  h,
		conn: conn,
		send: make(chan Message, 3),
	}

	// token, err := wsutil.ReadClientText(c.conn)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// if auth.Authenticate(string(token)) == false {
	// 	log.Printf("bad token %s\n", string(token))
	// 	return
	// }
	// log.Printf("token is good\n")
	// c.hub.register <- c

	go c.reader(auth)
	go c.writer()
}

func (c *Client) ping() error {
	err := wsutil.WriteServerMessage(c.conn, ws.OpPing, []byte{})
	return err
}

func (c *Client) reader(auth Authenticator) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// auth is required first
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	msg, err := wsutil.ReadClientText(c.conn)
	if err != nil {
		// timeout or error
		log.Println(err)
		return
	}
	log.Printf("got: %s\n", string(msg))
	cmd := strings.Split(string(msg), " ")
	if cmd[0] != "/auth" {
		// only auth is allowed
		log.Println("not /auth")
		return
	}
	token := cmd[1]
	if auth.Authenticate(string(token)) == false {
		// auth failed
		log.Printf("bad token %s\n", string(token))
		return
	}

	// register authenticate client
	c.hub.register <- c

	for {
		c.conn.SetReadDeadline(time.Now().Add(45 * time.Second))
		msg, err := wsutil.ReadClientText(c.conn)
		if err != nil && errors.Is(err, os.ErrDeadlineExceeded) {
			// keep alive with pings
			err = c.ping()
			if err != nil {
				log.Println(err)
				return
			}
			continue
		} else if err != nil {
			log.Println(err)
			return
		}
		log.Printf("got: %s\n", string(msg))
		if msg[0] == byte('/') {
			cmd := strings.Split(string(msg[1:]), " ")
			switch cmd[0] {
			case "ping":
				if len(cmd) == 2 {
					// "/ping time"
					pong := fmt.Sprintf("/pong %s", cmd[1])
					c.send <- Message{body: []byte(pong)}
				}
			default:
				log.Printf("ignore '%s'\n", cmd[0])
			}
		} else {
			c.hub.broadcast <- Message{sender: c, body: msg}
		}
	}
}

func (c *Client) writer() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			err := wsutil.WriteServerText(c.conn, message.body)
			if err != nil {
				log.Println(err)
				return
			}

			// drain the queue
			queued := len(c.send)
			for i := 0; i < queued; i++ {
				message = <-c.send
				err := wsutil.WriteServerText(c.conn, message.body)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
