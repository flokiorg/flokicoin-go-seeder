package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"golang.org/x/net/publicsuffix"
)

type RecordAction string

const (
	Add    RecordAction = "add"
	Delete RecordAction = "delete"
)

var (
	connectionTimeout = 30 * time.Second
)

type ARecordMessage struct {
	Domain string
	IP     string
	Action RecordAction
}

type CFHandler struct {
	api     *cloudflare.API
	zoneID  *cloudflare.ResourceContainer
	dnsname string
	cache   map[string]time.Time

	mu sync.Mutex
}

func NewCFHandler(ctx context.Context, apiToken, dnsname string) (*CFHandler, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, err
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(dnsname)
	if err != nil {
		return nil, err
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	cf := &CFHandler{
		api:     api,
		dnsname: dnsname,
		zoneID:  cloudflare.ZoneIdentifier(zoneID),
		cache:   make(map[string]time.Time),
	}

	if err := cf.initCache(ctx); err != nil {
		return nil, fmt.Errorf("cache failed: %v", err)
	}

	return cf, nil
}

func (cf *CFHandler) AddARecord(ctx context.Context, ip string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	record := cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    cf.dnsname,
		Content: ip,
		TTL:     120,
	}

	_, err := cf.api.CreateDNSRecord(ctxTimeout, cf.zoneID, record)

	if err != nil && strings.Contains(err.Error(), "An identical record already exists") {
		return nil
	}
	return err
}

func (cf *CFHandler) DeleteARecord(ctx context.Context, ip string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	records, _, err := cf.api.ListDNSRecords(ctxTimeout, cf.zoneID, cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: cf.dnsname,
	})
	if err != nil {
		return err
	}
	for _, r := range records {
		if r.Content == ip {
			ctxDel, cancelDel := context.WithTimeout(ctx, connectionTimeout)
			defer cancelDel()
			err = cf.api.DeleteDNSRecord(ctxDel, cf.zoneID, r.ID)
			if err != nil {
				log.Printf("failed to delete record: %v", err)
			}
		}
	}
	return nil
}

func (cf *CFHandler) ListRecords(ctx context.Context) ([]cloudflare.DNSRecord, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	records, _, err := cf.api.ListDNSRecords(ctxTimeout, cf.zoneID, cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: cf.dnsname,
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (cf *CFHandler) initCache(ctx context.Context) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	ctxTimeout, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	records, _, err := cf.api.ListDNSRecords(ctxTimeout, cf.zoneID, cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: cf.dnsname,
	})
	if err != nil {
		return err
	}

	for _, r := range records {
		cf.cache[r.Content] = time.Now()
	}

	return nil
}

func (cf *CFHandler) Addresses() []string {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	addrs := make([]string, 0, len(cf.cache))
	for addr := range cf.cache {
		addrs = append(addrs, addr)
	}

	return addrs
}

func (cf *CFHandler) Update(name string, cgList []string) (added, deleted int) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	goodIPs := make(map[string]struct{}, len(cgList))
	for _, ip := range cgList {
		goodIPs[ip] = struct{}{}
	}

	// Add IPs not in cache
	for ip := range goodIPs {
		if _, exists := cf.cache[ip]; !exists {
			if err := cf.AddARecord(context.Background(), ip); err != nil {
				log.Printf("%s: failed to add A record for IP %s: %v", name, ip, err)
			} else {
				cf.cache[ip] = time.Now()
				added++
			}
		}
	}

	// Delete IPs not in the good list
	for ip := range cf.cache {
		if _, isGood := goodIPs[ip]; !isGood {
			if err := cf.DeleteARecord(context.Background(), ip); err != nil {
				log.Printf("%s: failed to delete A record for IP %s: %v", name, ip, err)
			} else {
				delete(cf.cache, ip)
				deleted++
			}
		}
	}

	return added, deleted
}

func (cf *CFHandler) ListenAndProcess(ctx context.Context, ch <-chan ARecordMessage) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Cloudflare handler stopped")
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var err error
			switch msg.Action {
			case Add:
				err = cf.AddARecord(ctx, msg.IP)
			case Delete:
				err = cf.DeleteARecord(ctx, msg.IP)
			}
			if err != nil {
				log.Printf("error processing %s for %s: %v", msg.Action, msg.Domain, err)
			}
		}
	}
}
