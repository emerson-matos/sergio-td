package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/sergio-td/server/internal/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ws.Health)
	mux.HandleFunc("/metrics", ws.Metrics)
	mux.HandleFunc("/ws", ws.Handle)

	addr := ":8080"
	slog.Info("server listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
