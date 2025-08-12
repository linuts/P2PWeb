package main

import (
	"log"
	"net/http"

	"github.com/example/p2pweb/internal/server"
)

func main() {
	addr := ":8080"
	log.Printf("listening on %s", addr)
	s := server.New()
	if err := http.ListenAndServe(addr, s.Handler()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
