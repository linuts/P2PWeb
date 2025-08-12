package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

func handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	for _, q := range r.Question {
		if q.Qtype == dns.TypeA && strings.HasSuffix(q.Name, ".p2p.") {
			rr := &dns.A{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
				A:   net.ParseIP("127.0.0.1"),
			}
			m.Answer = append(m.Answer, rr)
		}
	}

	if err := w.WriteMsg(m); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func startDNSServer() error {
	addr := "127.1.1.153:53"
	dns.HandleFunc(".", handleDNS)
	server := &dns.Server{Addr: addr, Net: "udp"}
	log.Printf("Starting DNS server on %s", addr)
	return server.ListenAndServe()
}

func startWebServer() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s\n", r.Host)
	})
	addr := ":8080"
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func main() {
	errc := make(chan error, 2)
	go func() { errc <- startDNSServer() }()
	go func() { errc <- startWebServer() }()
	log.Fatalf("server error: %v", <-errc)
}
