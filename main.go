package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/miekg/dns"
)

func listenMessages(messages chan string) {
	for {
		msg := <-messages
		log.Println(msg)
	}
}

func main() {
	port := flag.Int("port", 53, "port to run on")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("tunnel accepts exactly one argument for the top domain")
	}

	tun := newTunnel(flag.Arg(0))
	dns.Handle(".", tun)
	go listenMessages(tun.Messages)

	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()
	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "tcp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set tcp listener %s\n", err.Error())
		}
	}()

	select {} // block forever
}
