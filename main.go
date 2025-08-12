package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

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

// setLocalResolver ensures /etc/resolv.conf prefers the local DNS server.
func setLocalResolver() {
	const ns = "nameserver 127.0.0.1\n"
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		log.Printf("unable to read resolv.conf: %v", err)
		return
	}
	if bytes.HasPrefix(data, []byte(ns)) {
		return
	}
	if err := os.WriteFile("/etc/resolv.conf", append([]byte(ns), data...), 0644); err != nil {
		log.Printf("unable to write resolv.conf: %v", err)
	} else {
		log.Printf("/etc/resolv.conf updated to use local DNS server")
	}
}

func startDNSServer(addr string) {
	server := &dns.Server{Addr: addr, Net: "udp", Handler: &p2pHandler{}}
	log.Printf("DNS server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("DNS server failed: %v", err)
	}
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
	setLocalResolver()
	go startWebServer(":8080")
	startDNSServer(":53")
}
