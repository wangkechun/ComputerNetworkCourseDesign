package main

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/internal/iana"
	"golang.org/x/net/ipv4"
	// "golang.org/x/net/ipv6"
	"log"
	// "net"
	"os"
)

func main() {

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	rb := make([]byte, 1500)
	// n, peer, err := c.ReadFrom(rb)
	if err != nil {
		log.Fatal(err)
	}
	rm, err := icmp.ParseMessage(iana.ProtocolICMP, rb[:n])
	if err != nil {
		log.Fatal(err)
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		log.Printf("got reflection from %v", peer)
	default:
		log.Printf("got %+v; want echo reply", rm)
	}
}
