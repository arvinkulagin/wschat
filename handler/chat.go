package handler

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/arvinkulagin/websub"
)

type Chat struct {
	Broker *websub.Broker
}

func (c Chat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tag := mux.Vars(r)["tag"]

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	messanger := c.Broker.Subscribe(tag, conn)

	log.Printf("Connect: %s\n", conn.RemoteAddr().String())

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			messanger.Unsubscribe(conn)
			conn.Close()
			log.Printf("Disconnect: %s\n", conn.RemoteAddr().String())
			return
		}
		messanger.Publish(msg)
	}
}