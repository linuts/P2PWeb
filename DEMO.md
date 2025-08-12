# P2PWeb Demo

This repository now includes a minimal Go program that runs:

* a DNS server mapping `.p2p` domains to `127.0.0.1`
* a basic HTTP server that responds based on the request's Host header

## Running the demo

Run as root so the program can bind to port 80:

```sh
sudo go run .
```

The program will:

1. Start an HTTP server on port `80`.
2. Start a DNS server on UDP port `5353` that resolves `example.p2p` to `127.0.0.1`.

Because the DNS server listens on port `5353`, your system resolver can stay
running.

With the program running, you can reach the demo site:

```
dig @127.0.0.1 -p 5353 example.p2p +short
curl -H 'Host: example.p2p' http://127.0.0.1/
```

The output from `dig` should show:

```
127.0.0.1
```

The output from `curl` should be:

```
Hello from example.p2p!
```

Stop the program with `Ctrl+C` when finished.
