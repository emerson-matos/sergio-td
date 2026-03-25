package main

import (
	"log"
	"net/http"

	"github.com/sergio-td/server/internal/ws"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ws.Health)
	mux.HandleFunc("/ws", ws.Handle)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
