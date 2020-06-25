# browsertunnel

Browser-based DNS tunneling for surreptitious data exfiltration!

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

Finally, test out your tunnel! You can use my demo page [here](https://jse.li/browsertunnel/html/index.html) or clone this repo and load [`html/index.html`](https://github.com/veggiedefender/browsertunnel/blob/main/html/index.html) locally. If everything works, you should be able to see messages logged to stdout.
