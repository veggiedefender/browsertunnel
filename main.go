package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"github.com/veggiedefender/browsertunnel/tunnel"
)

func listenMessages(messages chan string) {
	for {
		msg := <-messages
		log.Println("RECEIVED MESSAGE:", msg)
	}
}

func main() {
	port := flag.Int("port", 53, "port to run on")
	expiration := flag.Int("expiration", 60, "seconds an incomplete message is retained before it is deleted")
	deletionInterval := flag.Int("deletionInterval", 5, "seconds in between checks for expired messages")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("tunnel accepts exactly one argument for the top domain")
	}

	topDomain := flag.Arg(0)
	expirationDuration := time.Duration(*expiration) * time.Second
	deletionIntervalDuration := time.Duration(*deletionInterval) * time.Second

	tun := tunnel.NewTunnel(topDomain, expirationDuration, deletionIntervalDuration)
	dns.Handle(topDomain, tun)
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
