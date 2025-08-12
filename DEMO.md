# P2PWeb Demo

This repository now includes a minimal Go program that runs:

* a DNS server mapping `.p2p` domains to `127.0.0.1`
* a basic HTTP server that responds based on the request's Host header

## Running the demo

```
go run .
```

The program will:

1. Update `/etc/resolv.conf` to prefer `127.0.0.1`.
2. Start an HTTP server on port `8080`.
3. Start a DNS server on UDP port `53` that resolves `example.p2p` to `127.0.0.1`.

Port `53` may already be in use or require elevated privileges. If the DNS server
fails to start, a log message will be printed and `/etc/resolv.conf` will be
restored.

With the program running, you can reach the demo site:

```
curl http://example.p2p:8080/
```

The output should be:

```
Hello from example.p2p:8080!
```

Stop the program with `Ctrl+C` when finished. The original `/etc/resolv.conf`
is restored automatically on shutdown.
