package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func startDNSServer(addr string) error {
	server := &dns.Server{
		Addr:              addr,
		Net:               "udp",
		Handler:           &p2pHandler{},
		NotifyStartedFunc: func() { log.Printf("DNS server listening on %s", addr) },
	}
	return server.ListenAndServe()
}

func startWebServer(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s!\n", r.Host)
	})
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("HTTP server failed: %v", err)
	}
}

// configureResolver temporarily replaces /etc/resolv.conf so the host resolves
// names through the local DNS server. It returns a cleanup function that
// restores the original file.
func configureResolver() (func(), error) {
	const path = "/etc/resolv.conf"
	orig, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte("nameserver 127.0.0.1\n"), 0644); err != nil {
		return nil, err
	}
	return func() {
		os.WriteFile(path, orig, 0644)
	}, nil
}

func main() {
	cleanup, err := configureResolver()
	if err != nil {
		log.Fatalf("failed to configure resolver: %v", err)
	}
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
