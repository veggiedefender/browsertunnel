package main

import (
	"log"

	"github.com/miekg/dns"
)

func listenMessages(messages chan string) {
	for {
		msg := <-messages
		log.Println(msg)
	}
}

func main() {
	srv := &dns.Server{Addr: ":53", Net: "udp"}
	tun := newTunnel("t1.jse.li.")
	dns.Handle(".", tun)

	go listenMessages(tun.Messages)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
