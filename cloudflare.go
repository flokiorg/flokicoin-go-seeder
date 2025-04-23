package main

import (
	"context"
	"log"
	"strings"
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

	return &CFHandler{
		api:     api,
		dnsname: dnsname,
		zoneID:  cloudflare.ZoneIdentifier(zoneID),
	}, nil
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

	if strings.Contains(err.Error(), "An identical record already exists") {
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
