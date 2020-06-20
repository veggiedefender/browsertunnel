package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/miekg/dns"
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

	expirationDuration := time.Duration(*expiration) * time.Second
	deletionIntervalDuration := time.Duration(*deletionInterval) * time.Second

	tun := newTunnel(flag.Arg(0), expirationDuration, deletionIntervalDuration)
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
