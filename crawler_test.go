package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/peer"
	"github.com/flokiorg/go-flokicoin/wire"
)

const (
	nodeAddress = "117.46.95.19:15212"
)

func TestCrawlIP(t *testing.T) {

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
	select {}
}

func TestCrawler(t *testing.T) {

	// Create TCP connection
	conn, err := net.Dial("tcp", nodeAddress)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to node:", nodeAddress)

	// Create a new Bitcoin peer instance
	verack := make(chan struct{})
	cfg := &peer.Config{
		UserAgentName:    "MyDNSClient",
		UserAgentVersion: "0.1",
		ChainParams:      &chaincfg.MainNetParams,
		Services:         0,
		Listeners: peer.MessageListeners{

			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				fmt.Printf("outbound: received version: %v\n", msg.ProtocolVersion)
				return nil
			},
			OnVerAck: func(p *peer.Peer, msg *wire.MsgVerAck) {
				fmt.Printf("outbound: ver ack\n")
				verack <- struct{}{}
			},
			OnAddrV2: func(p *peer.Peer, msg *wire.MsgAddrV2) {
				fmt.Println("Received-1", len(msg.AddrList), "addresses:")
				for _, addr := range msg.AddrList {
					fmt.Printf(" - %s:%d\n", addr.Addr, addr.Port)
				}
			},
		},
	}
	nodePeer, err := peer.NewOutboundPeer(cfg, nodeAddress)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	// Associate the peer with the connection
	nodePeer.AssociateConnection(conn)

	select {
	case <-verack:
	case <-time.After(time.Second * 1):
		log.Fatal("verack timeout")
	}

	nodePeer.QueueMessage(wire.NewMsgGetAddr(), nil)
	// nodePeer.QueueMessage(wire.NewMsgAddrV2(), nil)

	<-time.After(time.Second * 5)
}

func TestCrawler_2(t *testing.T) {

	verack := make(chan struct{})
	onAddr := make(chan *wire.MsgAddr)
	peerCfg := &peer.Config{
		UserAgentName:    "flokicoin-seeder",
		UserAgentVersion: "0.1",
		Services:         0,
		ChainParams:      &chaincfg.MainNetParams,
		Listeners: peer.MessageListeners{
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				log.Printf("Remote version: %v\n", msg.ProtocolVersion)
				return nil
			},
			OnVerAck: func(p *peer.Peer, msg *wire.MsgVerAck) {
				verack <- struct{}{}
			},
			OnAddrV2: func(p *peer.Peer, msg *wire.MsgAddrV2) {
				log.Printf("OnAddrV2: %v", msg)
				onAddr <- msgAddrV2ToMsgAddr(msg)
			},
		},
	}

	// Create and start the outbound peer
	p, err := peer.NewOutboundPeer(peerCfg, nodeAddress)
	if err != nil {
		t.Fatal(&crawlError{"NewOutboundPeer: error", err})
	}

	// Establish the connection to the peer address and mark it connected.
	conn, err := net.Dial("tcp", p.Addr())
	if err != nil {
		t.Fatal(&crawlError{"net.Dial: error", err})
	}
	p.AssociateConnection(conn)

	defer p.WaitForDisconnect()
	defer p.Disconnect()

	// check verack
	select {
	case <-verack:
	case <-time.After(time.Second * 3):
		t.Fatal(&crawlError{"verack timeout", errors.New("")})
	}

	p.QueueMessage(wire.NewMsgGetAddr(), nil)

	addrMsg := new(wire.MsgAddr)
	select {
	case addrMsg = <-onAddr:
		log.Printf("receiving: %v", addrMsg)

	case <-time.After(time.Second * 5):
		log.Printf("timeout")
	}

	t.Logf("got %d addrs", len(addrMsg.AddrList))

	for _, addr := range addrMsg.AddrList {
		t.Logf("addr: %s:%d", addr.IP, addr.Port)
	}
}
