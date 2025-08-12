# P2PWeb Demo

This repository now includes a minimal Go program that runs:

* a DNS server mapping `.p2p` domains to `127.0.0.1`
* a basic HTTP server that responds based on the request's Host header

## Running the demo

Run as root so the program can bind to port 53:

```sh
sudo go run .
```

The program will:

1. Start a DNS server on UDP port `53` that resolves `example.p2p` to `127.0.0.1`.
2. Start an HTTP server on port `8080`.
3. Replace `/etc/resolv.conf` so the host uses `127.0.0.1` for DNS, and restore
   the original file when the program exits.

With the program running, you can reach the demo site:

```
dig example.p2p +short
curl --noproxy '*' http://example.p2p:8080/
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
