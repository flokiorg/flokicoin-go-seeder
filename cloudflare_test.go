package main

import (
	"context"
	"os"
	"testing"
)

func setupConnection(t *testing.T) *CFHandler {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		t.Skip("CLOUDFLARE_API_TOKEN not set; skipping live Cloudflare integration test")
	}

	cf, err := NewCFHandler(context.Background(), apiToken, "dnsseed.myfloki.com")
	if err != nil {
		t.Fatal(err)
	}

	return cf
}

func TestListRecords(t *testing.T) {

	cf := setupConnection(t)

	records, err := cf.ListRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range records {
		t.Logf("record: %v : %v", r.Name, r.Content)
	}
}

func TestManageRecords(t *testing.T) {

	cf := setupConnection(t)

	ips := []string{
		"1.1.1.1",
		"2.2.2.2",
		"3.3.3.3",
		"4.4.4.4",
		"5.5.5.5",
	}

	for _, ip := range ips {
		if err := cf.DeleteARecord(context.Background(), ip); err != nil {
			t.Fatal(err)
		}
	}

	records, err := cf.ListRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prevLen := len(records)
	for _, ip := range ips {
		if err := cf.AddARecord(context.Background(), ip); err != nil {
			t.Fatal(err)
		}
	}

	for _, ip := range ips {
		if err := cf.DeleteARecord(context.Background(), ip); err != nil {
			t.Fatal(err)
		}
	}

	records, err = cf.ListRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if prevLen != len(records) {
		t.Logf("unexpected length")
	}

}

func TestPruneRecords(t *testing.T) {

	cf := setupConnection(t)

	records, err := cf.ListRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range records {
		t.Logf("deleting %s", r.Content)

		if err := cf.DeleteARecord(context.Background(), r.Content); err != nil {
			t.Fatal(err)
		}
	}
}
