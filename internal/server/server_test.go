package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRoot(t *testing.T) {
	s := New()
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET root failed: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	expected := "p2p web placeholder\n"
	if string(b) != expected {
		t.Fatalf("expected %q got %q", expected, b)
	}
}

func TestPeersEndpoint(t *testing.T) {
	s := New()
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// initial list should be empty
	resp, err := http.Get(ts.URL + "/peers")
	if err != nil {
		t.Fatalf("GET peers: %v", err)
	}
	var peers []string
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		t.Fatalf("decode peers: %v", err)
	}
	resp.Body.Close()
	if len(peers) != 0 {
		t.Fatalf("expected empty peers, got %v", peers)
	}

	// post a new peer
	body := bytes.NewBufferString(`{"addr":"peer1"}`)
	resp, err = http.Post(ts.URL+"/peers", "application/json", body)
	if err != nil {
		t.Fatalf("POST peer: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// verify list
	resp, err = http.Get(ts.URL + "/peers")
	if err != nil {
		t.Fatalf("GET peers: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		t.Fatalf("decode peers: %v", err)
	}
	resp.Body.Close()
	if len(peers) != 1 || peers[0] != "peer1" {
		t.Fatalf("unexpected peers: %v", peers)
	}
}
