package main

import (
	"html/template"
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
	tmpl := template.Must(template.New("dashboard").Parse(`<!DOCTYPE html>
<html>
<head><title>P2P Web Dashboard</title></head>
<body>
<h1>P2P Web Dashboard</h1>
<ul>
<li>Domain suffix: {{.DomainSuffix}}</li>
<li>DNS server: {{.DNSServer}}</li>
<li>Example domain: <a href="http://{{.ExampleDomain}}">{{.ExampleDomain}}</a></li>
<li>Web server port: {{.WebPort}}</li>
</ul>
</body>
</html>`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			DomainSuffix  string
			DNSServer     string
			ExampleDomain string
			WebPort       string
		}{
			DomainSuffix:  ".p2p",
			DNSServer:     "127.1.1.153:53",
			ExampleDomain: "server.p2p",
			WebPort:       "80",
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("failed to render template: %v", err)
		}
	})
	addr := ":80"
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func main() {
	errc := make(chan error, 2)
	go func() { errc <- startDNSServer() }()
	go func() { errc <- startWebServer() }()
	log.Fatalf("server error: %v", <-errc)
}
