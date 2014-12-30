package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)
import "golang.org/x/net/ipv4"

var _ = bytes.MinRead
var _ = binary.MaxVarintLen16
var _ = fmt.Scanln
var _ = log.Println
var _ = time.ANSIC
var _ = net.Dial
var _ = ipv4.DontFragment

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func CheckSum(data []byte) uint16 {
	length := len(data)
	var sum uint32
	var index int
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	return uint16(^sum)
}

func ping(host string, ttl int) {
	log.Println("Ping ", host)
	raddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Println("resolve ip addr error", err)
	} else {
		log.Println("IP:", raddr)
	}
	conn, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		log.Println(err)
	}
	p := ipv4.NewConn(conn)
	p2, _ := ipv4.NewRawConn(conn)
	if err := p.SetTTL(ttl); err != nil {
		log.Println(err)
	}
	var icmp ICMP
	icmp.Type = 8
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 0
	icmp.SequenceNum = 0
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = CheckSum(buffer.Bytes())
	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, icmp)
	log.Println("Runing Ping data ", printByte(buffer.Bytes()))
	conn.Write(buffer.Bytes())
	t_start := time.Now()
	conn.SetReadDeadline((time.Now().Add(time.Second * 1)))
	recv := make([]byte, 100)
	p2.ReadFrom(recv)
	// recv_len, err := conn.Read(recv)
	recv_len := 100
	if err != nil {
		log.Println(err)
	}

	// log.Println("Recv data check error", recv, buffer.Bytes())
	// }
	log.Println("Recv data ", printByte(recv[:recv_len]))
	t_end := time.Now()
	dur := t_end.Sub(t_start).Nanoseconds() / 1e6
	log.Println("Time spend ms", dur)
	log.Println("")
	defer conn.Close()
}

func printByte(b []byte) (r string) {
	l := len(b)
	for i := 0; i < l; i += 4 {
		r += fmt.Sprint(b[i:i+4], " ")
	}
	return
}

func main() {
	var host = flag.String("host", "125.221.232.253", "usage: -host=baidu.com")
	flag.Parse()
	for i := 1; i < 10; i++ {
		// ping("125.221.232.253")
		log.Println("ttl:", i)
		ping(*host, i)
		// time.Sleep(time.Second / 2)
	}
}
