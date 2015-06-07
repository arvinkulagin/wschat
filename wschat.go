// Yet another websocket chat service.
package main

import (
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

func main() {
	broker := NewBroker()
	indexTemplate, err := template.ParseFiles("wschat.html")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexTemplate.Execute(w, nil)
	})
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Connect: %s\n", conn.RemoteAddr().String())
		broker.Subscribe(conn)
		defer broker.Unsubscribe(conn)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				log.Printf("Disconnect: %s\n", conn.RemoteAddr().String())
				return
			}
			broker.Publish(msg)
		}
	})
	log.Println("Listen on localhost:8888")
	log.Fatal(http.ListenAndServe("localhost:8888", nil))
}

type Broker struct {
	subscribe   chan *websocket.Conn
	unsubscribe chan *websocket.Conn
	publish     chan []byte
}

func NewBroker() *Broker {
	b := Broker{
		subscribe:   make(chan *websocket.Conn),
		unsubscribe: make(chan *websocket.Conn),
		publish:     make(chan []byte),
	}
	go func() {
		conns := []*websocket.Conn{}
		messages := [][]byte{}
		for {
			select {
			case conn := <-b.subscribe:
				conns = append(conns, conn)
			case conn := <-b.unsubscribe:
				for i, c := range conns {
					if c.RemoteAddr().String() == conn.RemoteAddr().String() {
						conns = append(conns[:i], conns[i+1:]...)
						break
					}
				}
			case msg := <-b.publish:
				messages = append(messages)
				for _, conn := range conns {
					conn.WriteMessage(websocket.TextMessage, msg)
				}
			}
		}
	}()
	return &b
}

func (b *Broker) Subscribe(conn *websocket.Conn) {
	b.subscribe <- conn
}

func (b *Broker) Unsubscribe(conn *websocket.Conn) {
	b.unsubscribe <- conn
}

func (b *Broker) Publish(msg []byte) {
	b.publish <- msg
}