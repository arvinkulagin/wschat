package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"time"
)

const templatePath string = "wschat.html"

func main() {
	addr := flag.String("addr", ":8888", "Network address")
	size := flag.Int("buf", 50, "Message buffer size")
	flag.Parse()

	broker := NewBroker(*size)
	indexTemplate, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Messages := broker.Buffer()
		indexTemplate.Execute(w, Messages)
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
			session := []byte(fmt.Sprintf("%s: ", time.Now().Format("15:04:05")))
			out := append(session, msg...)
			broker.Publish(out)
		}
	})
	log.Printf("Listen on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type Broker struct {
	subscribe   chan *websocket.Conn
	unsubscribe chan *websocket.Conn
	publish     chan []byte
	buffer      chan chan []string
}

func NewBroker(size int) *Broker {
	b := Broker{
		subscribe:   make(chan *websocket.Conn),
		unsubscribe: make(chan *websocket.Conn),
		publish:     make(chan []byte),
		buffer:      make(chan chan []string),
	}
	go func() {
		conns := []*websocket.Conn{}
		messages := []string{}
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
				if len(messages) < size {
					messages = append(messages, string(msg))
				} else {
					messages = append(messages[1:], string(msg))
				}
				for _, conn := range conns {
					conn.WriteMessage(websocket.TextMessage, msg)
				}
			case ch := <-b.buffer:
				ch <- messages
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

func (b *Broker) Buffer() []string {
	ch := make(chan []string)
	b.buffer <- ch
	m := <-ch
	return m
}