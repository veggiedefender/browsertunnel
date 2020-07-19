package tunnel

import (
	"encoding/base32"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

var decoder = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding('0')

// A Tunnel listens for DNS queries. Messages that are collected and decoded are outputted through
// the Messages channel.
type Tunnel struct {
	Messages       chan string
	cancel         chan struct{}
	fgLists        map[string]*fragmentList
	fgListsLock    sync.Mutex
	topDomain      string
	domains        chan string
	maxMessageSize int
}

type fragmentList struct {
	totalSize int
	fragments map[int]fragment
	expiresAt time.Time
}

type fragment struct {
	id        string
	totalSize int
	offset    int
	data      string
}

// NewTunnel creates a new tunnel and starts goroutines to manage messages.
//
// The expiration argument decides how long (at a minimum) a partial message is kept in memory
// before being deleted. Updating a message resets its expiration timer.
//
// A goroutine running in the background periodically loops through each partial message in memory
// and removes messages that are expired. The deletionInterval argument controls how often this loop
// runs. Checking for expiration requires a full lock on the internal map of messages; therefore,
// values of deletionInterval that are too frequent may hurt performance.
//
// The maxMessageSize argument configures the maximum size of an encoded message that the tunnel
// will accept. Messages that declare a size greater than maxMessageSize will be discarded.
func NewTunnel(topDomain string, expiration time.Duration, deletionInterval time.Duration, maxMessageSize int) *Tunnel {
	tun := &Tunnel{
		Messages:       make(chan string, 256),
		cancel:         make(chan struct{}),
		topDomain:      topDomain,
		domains:        make(chan string, 256),
		fgLists:        make(map[string]*fragmentList),
		maxMessageSize: maxMessageSize,
	}
	go tun.listenDomains(expiration)
	go tun.removeExpiredMessages(deletionInterval)
	return tun
}

// Close cleans up and stops the goroutines created by the tunnel. Calling Close() more than once
// will panic.
func (tun *Tunnel) Close() {
	close(tun.cancel)
}

func parseDomain(topDomain string, domain string) (fragment, error) {
	if !strings.HasSuffix(domain, "."+topDomain) {
		return fragment{}, fmt.Errorf("Domain %q does not have top domain %q", domain, topDomain)
	}
	payload := strings.TrimSuffix(domain, "."+topDomain)
	labels := strings.Split(payload, ".")
	if len(labels) < 4 {
		return fragment{}, fmt.Errorf("Domain %q has %d labels but expected at least 4", domain, len(labels))
	}
	id := labels[0]

	totalSize, err := strconv.Atoi(labels[1])
	if err != nil {
		return fragment{}, err
	}

	offset, err := strconv.Atoi(labels[2])
	if err != nil {
		return fragment{}, err
	}

	data := strings.Join(labels[3:], "")

	return fragment{
		id:        id,
		totalSize: totalSize,
		offset:    offset,
		data:      data,
	}, nil
}

func (fl fragmentList) assemble() (string, error) {
	buf := make([]rune, fl.totalSize)
	for _, f := range fl.fragments {
		if f.offset >= fl.totalSize {
			return "", fmt.Errorf("Offset %d > total size %d", f.offset, fl.totalSize)
		}
		copy(buf[f.offset:], []rune(f.data))
	}
	dec, err := decoder.DecodeString(string(buf))
	if err != nil {
		return "", err
	}
	return string(dec), nil
}

func (tun *Tunnel) listenDomains(expiration time.Duration) {
	for {
		select {
		case <-tun.cancel:
			return
		case domain := <-tun.domains:
			func() {
				tun.fgListsLock.Lock()
				defer tun.fgListsLock.Unlock()

				fg, err := parseDomain(tun.topDomain, domain)
				if err != nil {
					log.Println(err)
					return
				}
				if fg.totalSize <= 0 {
					log.Printf("Received message that declares non-positive length %d.", fg.totalSize)
					return
				}
				if fg.totalSize > tun.maxMessageSize {
					log.Printf("Received message that declares length %d. Max message size is %d", fg.totalSize, tun.maxMessageSize)
					return
				}

				if _, ok := tun.fgLists[fg.id]; !ok {
					tun.fgLists[fg.id] = &fragmentList{
						totalSize: 0,
						fragments: make(map[int]fragment),
						expiresAt: time.Now().Add(expiration),
					}
				}
				fgList := tun.fgLists[fg.id]
				fgList.totalSize = fg.totalSize
				fgList.fragments[fg.offset] = fg
				fgList.expiresAt = time.Now().Add(expiration)

				totalBytes := 0
				for _, fg := range fgList.fragments {
					totalBytes += len(fg.data)
				}

				if totalBytes >= fgList.totalSize {
					msg, err := fgList.assemble()
					if err != nil {
						log.Println(err)
						return
					}
					tun.Messages <- msg
					delete(tun.fgLists, fg.id)
				}
			}()
		}
	}
}

func (tun *Tunnel) removeExpiredMessages(deletionInterval time.Duration) {
	ticker := time.NewTicker(deletionInterval)
	for {
		select {
		case <-tun.cancel:
			ticker.Stop()
			return
		case <-ticker.C:
			tun.fgListsLock.Lock()
			now := time.Now()
			for id, fgList := range tun.fgLists {
				if fgList.expiresAt.Before(now) {
					delete(tun.fgLists, id)
				}
			}
			tun.fgListsLock.Unlock()
		}
	}
}

// ServeDNS handles DNS queries, records them, and replies with a CNAME to blackhole-1.iana.org.
func (tun *Tunnel) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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
