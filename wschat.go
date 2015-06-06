package main

import (
	"net/http"
	"log"
	"fmt"
	"flag"
	"github.com/arvinkulagin/wschat/handler"
	"github.com/gorilla/mux"
	"github.com/arvinkulagin/websub"
)

func main() {
	addr := flag.String("Address", "localhost:8888", "Address")
	flag.Parse()

	broker := websub.NewBroker()

	r := mux.NewRouter()
	r.Handle("/chat/{tag}", handler.Chat{Broker: broker})
	http.Handle("/", r)
	fmt.Printf("Listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}