package main

import (
	"encoding/base32"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

var decoder = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding('0')

type tunnel struct {
	Messages  chan string
	fragments map[string]*msgFragmentList
	topDomain string
	domains   chan string
}

type msgFragment struct {
	id        string
	totalSize int
	offset    int
	data      string
}

type msgFragmentList struct {
	totalSize int
	fragments map[int]msgFragment
}

func newTunnel(topDomain string) *tunnel {
	tun := &tunnel{
		Messages:  make(chan string, 256),
		topDomain: topDomain,
		domains:   make(chan string, 256),
		fragments: make(map[string]*msgFragmentList),
	}
	go tun.listenDomains()
	return tun
}

func parseDomain(topDomain string, domain string) (msgFragment, error) {
	payload := strings.TrimSuffix(domain, topDomain)
	labels := strings.Split(payload, ".")
	if len(labels) < 4 {
		return msgFragment{}, fmt.Errorf("Domain has %d labels but expected at least 4", len(labels))
	}
	id := labels[0]

	totalSize, err := strconv.Atoi(labels[1])
	if err != nil {
		return msgFragment{}, err
	}

	offset, err := strconv.Atoi(labels[2])
	if err != nil {
		return msgFragment{}, err
	}

	data := strings.Join(labels[3:], "")

	return msgFragment{
		id:        id,
		totalSize: totalSize,
		offset:    offset,
		data:      data,
	}, nil
}

func (fl msgFragmentList) assemble() (string, error) {
	buf := make([]byte, fl.totalSize)
	for _, f := range fl.fragments {
		if f.offset >= fl.totalSize {
			return "", fmt.Errorf("Offset %d > total size %d", f.offset, fl.totalSize)
		}
		copy(buf[f.offset:], []byte(f.data))
	}
	dec, err := decoder.DecodeString(string(buf))
	if err != nil {
		return "", err
	}
	return string(dec), nil
}

func (tun *tunnel) listenDomains() {
	for {
		domain := <-tun.domains
		fragment, err := parseDomain(tun.topDomain, domain)
		if err != nil {
			log.Println(err)
			continue
		}

		if _, ok := tun.fragments[fragment.id]; !ok {
			tun.fragments[fragment.id] = &msgFragmentList{
				totalSize: 0,
				fragments: make(map[int]msgFragment),
			}
		}
		fragmentList := tun.fragments[fragment.id]
		fragmentList.totalSize = fragment.totalSize
		fragmentList.fragments[fragment.offset] = fragment

		totalBytes := 0
		for _, fragment := range fragmentList.fragments {
			totalBytes += len(fragment.data)
		}

		if totalBytes >= fragmentList.totalSize {
			msg, err := fragmentList.assemble()
			if err != nil {
				log.Println(err)
				return
			}
			tun.Messages <- msg
			delete(tun.fragments, fragment.id)
		}
	}
}

func (tun *tunnel) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) < 1 {
		return
	}

	domain := r.Question[0].Name

	if r.Question[0].Qtype == dns.TypeA {
		tun.domains <- domain
	}

	m := &dns.Msg{}
	m.SetReply(r)
	m.Answer = []dns.RR{
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: domain, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 0},
			Target: "blackhole-1.iana.org.",
		},
	}
	err := w.WriteMsg(m)
	if err != nil {
		log.Println(err)
	}
}
