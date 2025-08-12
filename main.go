package main

import (
        "bufio"
        "bytes"
        "fmt"
        "log"
        "net/http"
        "os"
        "os/exec"
        "os/signal"
        "strings"
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

func defaultLink() (string, error) {
        out, err := exec.Command("resolvectl", "status").Output()
        if err != nil {
                return "", err
        }
        scanner := bufio.NewScanner(bytes.NewReader(out))
        var link string
        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if strings.HasPrefix(line, "Link ") {
                        start := strings.Index(line, "(")
                        end := strings.Index(line, ")")
                        if start >= 0 && end > start {
                                link = line[start+1 : end]
                        }
                }
                if strings.HasPrefix(line, "DefaultRoute: yes") && link != "" {
                        return link, nil
                }
        }
        if err := scanner.Err(); err != nil {
                return "", err
        }
        return "", fmt.Errorf("no default link found")
}

// configureResolver directs `.p2p` lookups to the local DNS server and returns
// a cleanup function to restore settings.
func configureResolver(port int) (func(), error) {
        link, err := defaultLink()
        if err != nil {
                return nil, err
        }
        addr := fmt.Sprintf("127.0.0.1#%d", port)
        if err := exec.Command("resolvectl", "dns", link, addr).Run(); err != nil {
                return nil, err
        }
        if err := exec.Command("resolvectl", "domain", link, "~p2p").Run(); err != nil {
                exec.Command("resolvectl", "revert", link).Run()
                return nil, err
        }
        return func() {
                exec.Command("resolvectl", "revert", link).Run()
        }, nil
}

func main() {
        cleanup, err := configureResolver(5350)
        if err != nil {
                log.Printf("failed to configure resolver: %v", err)
        } else {
                defer cleanup()
        }

        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
        go func() {
                <-sigCh
                if cleanup != nil {
                        cleanup()
                }
                os.Exit(0)
        }()

        go startWebServer(":80")
        if err := startDNSServer(":5350"); err != nil {
                log.Printf("DNS server failed: %v", err)
        }
}
