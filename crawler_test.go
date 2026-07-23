package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/peer"
	"github.com/flokiorg/go-flokicoin/wire"
)

const (
	nodeAddress = "195.201.235.46:15212"
	// nodeAddress = "62.171.141.201:15212"
)

func TestCrawlIP(t *testing.T) {
	if testing.Short() {
		t.Skip("requires a live network connection to a real flokicoin node")
	}

	s := &dnsseeder{
		maxSize: 1,
	}
	r := &result{
		node: nodeAddress,
	}

	addrs, err := crawlIP(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("addrs is empty")
	}

	t.Logf("got %d addrs", len(addrs))

	for _, addr := range addrs {
		t.Logf("addr: %s:%d", addr.IP, addr.Port)
	}
}

func TestCrawler(t *testing.T) {
	if testing.Short() {
		t.Skip("requires a live network connection to a real flokicoin node")
	}

	// Create TCP connection
	conn, err := net.Dial("tcp", nodeAddress)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to node:", nodeAddress)

	// Create a new Bitcoin peer instance
	verack := make(chan struct{})
	onaddr := make(chan struct{})

	cfg := &peer.Config{
		UserAgentName:    "MyDNSClient",
		UserAgentVersion: "0.1",
		ChainParams:      &chaincfg.MainNetParams,
		Listeners: peer.MessageListeners{

			OnVerAck: func(p *peer.Peer, _ *wire.MsgVerAck) {
				fmt.Printf("Adding peer %v with services %v pver %d\n", p.NA().Addr.String(), p.Services(), p.ProtocolVersion())
				verack <- struct{}{}
			},

			OnAddrV2: func(p *peer.Peer, msg *wire.MsgAddrV2) {
				fmt.Printf("Peer sent %v addresses\n", len(msg.AddrList))
				for _, addr := range msg.AddrList {
					fmt.Printf(" - %s:%d\n", addr.Addr, addr.Port)
				}
				onaddr <- struct{}{}
			},
		},
	}
	nodePeer, err := peer.NewOutboundPeer(cfg, nodeAddress)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Associate the peer with the connection
	nodePeer.AssociateConnection(conn)

	select {
	case <-verack:
	case <-time.After(time.Second * 1):
		t.Logf("verack timeout")
		return
	}

	nodePeer.QueueMessage(wire.NewMsgGetAddr(), nil)

	select {
	case <-onaddr:
	case <-time.After(time.Second * 10):
		t.Logf("getaddr timeout")
	}
}
