package handler

import (
	"net/http"
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

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			messanger.Unsubscribe(conn)
			conn.Close()
			return
		}
		messanger.Publish(msg)
	}
}