package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/miekg/dns"
)

// p2pRecords maps p2p domain names to IP addresses.
var p2pRecords = map[string]string{
	"example.p2p.": "127.0.0.1",
}

// p2pHandler responds to DNS queries for .p2p domains.
type p2pHandler struct{}

func (h *p2pHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	for _, q := range r.Question {
		// DNS queries include a trailing dot.
		if ip, ok := p2pRecords[q.Name]; ok && q.Qtype == dns.TypeA {
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}

	if err := w.WriteMsg(m); err != nil {
		log.Printf("failed to write DNS response: %v", err)
	}
}

// setLocalResolver ensures /etc/resolv.conf prefers the local DNS server and
// returns a function to restore the original file.
func setLocalResolver() (func(), error) {
	const ns = "nameserver 127.0.0.1\n"
	orig, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return func() {}, fmt.Errorf("unable to read resolv.conf: %w", err)
	}
	if bytes.HasPrefix(orig, []byte(ns)) {
		return func() {}, nil
	}
	if err := os.WriteFile("/etc/resolv.conf", append([]byte(ns), orig...), 0644); err != nil {
		return func() {}, fmt.Errorf("unable to write resolv.conf: %w", err)
	}
	log.Printf("/etc/resolv.conf updated to use local DNS server")
	return func() {
		if err := os.WriteFile("/etc/resolv.conf", orig, 0644); err != nil {
			log.Printf("failed to restore resolv.conf: %v", err)
		} else {
			log.Printf("/etc/resolv.conf restored")
		}
	}, nil
}

func startDNSServer(addr string) error {
	server := &dns.Server{Addr: addr, Net: "udp", Handler: &p2pHandler{}}
	log.Printf("DNS server listening on %s", addr)
	return server.ListenAndServe()
}

func startWebServer(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s!\n", r.Host)
	})
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func main() {
	restore, err := setLocalResolver()
	if err != nil {
		log.Printf("resolver setup failed: %v", err)
	}
	var once sync.Once
	cleanup := func() { once.Do(restore) }
	defer cleanup()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cleanup()
		os.Exit(0)
	}()

	go startWebServer(":8080")
	if err := startDNSServer(":53"); err != nil {
		log.Printf("DNS server failed: %v", err)
	}
}
