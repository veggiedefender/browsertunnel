# browsertunnel

[![](https://godoc.org/github.com/veggiedefender/browsertunnel/tunnel?status.svg)](https://godoc.org/github.com/veggiedefender/browsertunnel/tunnel)

Browsertunnel is a tool for exfiltrating data from the browser using the DNS protocol. It achieves this by abusing [`dns-prefetch`](https://developer.mozilla.org/en-US/docs/Web/Performance/dns-prefetch), a feature intended to reduce the perceived latency of websites by doing DNS lookups in the background for specified domains. DNS traffic does not appear in the browser's debugging tools, is not blocked by a page's Content Security Policy (CSP), and is often not inspected by corporate firewalls or proxies, making it an ideal medium for smuggling data in constrained scenarios.

It's an old technique—DNS tunneling itself dates back to the '90s, and [Patrick Vananti](https://blog.compass-security.com/2016/10/bypassing-content-security-policy-with-dns-prefetching/) wrote about using `dns-prefetch` for it in 2016, but as far as I can tell, browsertunnel is the first open source, production-ready client/server demonstrating its use. Because `dns-prefetch` does not return any data back to client javascript, communication through browsertunnel is only unidirectional. Additionally, some browsers disable `dns-prefetch` by default, and in those cases, browsertunnel will silently fail.

<img src="https://user-images.githubusercontent.com/8890878/85884777-2b31cd80-b7b1-11ea-9e96-5f5ee8e10194.png" width="500">

The project comes in two parts:

1. A server, written in golang, functions as an authoritative DNS server which collects and decodes messages sent by browsertunnel.
2. A small javascript library, found in the [`html/`](https://github.com/veggiedefender/browsertunnel/tree/main/html) folder, encodes and sends messages from the client side.

## How it works

Browsertunnel can send arbitrary strings over DNS by encoding the string in a subdomain, which is forwarded to the browsertunnel server when the browser attempts to recursively resolve the domain.

<img src="https://user-images.githubusercontent.com/8890878/85882810-eeb0a280-b7ad-11ea-8f9e-709a268b0aa4.png" width="500">

Longer messages that cannot fit in one domain (253 bytes) are automatically split into multiple queries, which are reassembled and decoded by the server.

<img src="https://user-images.githubusercontent.com/8890878/85882813-efe1cf80-b7ad-11ea-94c7-063dcf6d0b06.png" width="500">

## Setup and usage

First, set up DNS records to delegate a subdomain to your server. For example, if your server's IP is `192.0.2.123` and you want to tunnel through the subdomain `t1.example.com`, then your DNS configuration will look like this:

```
t1		IN	NS	t1ns.example.com.
t1ns		IN	A	192.0.2.123
```

On your **server**, install browsertunnel using `go get`. Alternatively, compile browsertunnel on your own machine, and copy the binary to your server.

```
go get github.com/veggiedefender/browsertunnel
```

Next, run `browsertunnel`, specifying the subdomain you want to tunnel through.

```
browsertunnel t1.example.com
```

For full usage, run `browsertunnel -help`:

```
$ browsertunnel -help
Usage of browsertunnel:
  -deletionInterval int
    	seconds in between checks for expired messages (default 5)
  -expiration int
    	seconds an incomplete message is retained before it is deleted (default 60)
  -maxMessageSize int
    	maximum encoded size (in bytes) of a message (default 5000)
  -port int
    	port to run on (default 53)
```

For more detailed descriptions and rationale for these parameters, you may also consult the [godoc](https://godoc.org/github.com/veggiedefender/browsertunnel/tunnel).

Finally, test out your tunnel! You can use my demo page [here](https://jse.li/browsertunnel/html/index.html) or clone this repo and load [`html/index.html`](https://github.com/veggiedefender/browsertunnel/blob/main/html/index.html) locally. If everything works, you should be able to see messages logged to stdout.

For real-world applications of this project, you may want to fork and tweak the code as you see fit. Some inspiration:
* Write messages to a database instead of printing them to stdout
* Transpile or rewrite the client code to work with older browsers
* Make the ID portion of the domain larger or smaller, depending on the amount of traffic you get, and ID collisions you expect
* Authenticate and encrypt messages for secrecy and tamper-resistance (remember that DNS is a plaintext protocol)
