package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Server provides HTTP handlers for managing peers.
type Server struct {
	mu    sync.Mutex
	peers []string
}

// New returns a ready-to-use Server.
func New() *Server { return &Server{} }

// Handler exposes HTTP endpoints for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "p2p web placeholder")
	})
	mux.HandleFunc("/peers", s.peersHandler)
	return mux
}

func (s *Server) peersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		peers := append([]string(nil), s.peers...)
		s.mu.Unlock()
		if err := json.NewEncoder(w).Encode(peers); err != nil {
			http.Error(w, "encode", http.StatusInternalServerError)
		}
	case http.MethodPost:
		var req struct {
			Addr string `json:"addr"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.peers = append(s.peers, req.Addr)
		s.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
