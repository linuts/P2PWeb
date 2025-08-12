# P2PWeb Demo

This repository now includes a minimal Go program that runs:

* a DNS server mapping `.p2p` domains to `127.0.0.1`
* a basic HTTP server that responds based on the request's Host header

## Running the demo

Run as root so the program can configure DNS and bind to port 80:

```sh
sudo go run .
```

The program will:

1. Start a DNS server on UDP port `5350` that resolves `example.p2p` to `127.0.0.1`.
2. Start an HTTP server on port `80`.
3. Use `resolvectl` to direct `.p2p` lookups to `127.0.0.1#5350` and revert the change on exit.

With the program running, you can reach the demo site:

```
dig example.p2p +short
curl --noproxy '*' http://example.p2p/
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
