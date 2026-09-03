package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dayu255/myhp-ogo/src/og"
)

const LISTEN_PORT = 8080

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/og.png", og.Og)
	log.Printf("Listening PORT: %d\n", LISTEN_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", LISTEN_PORT), mux)
}
