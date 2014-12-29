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

var _ = bytes.MinRead
var _ = binary.MaxVarintLen16
var _ = fmt.Scanln
var _ = log.Println
var _ = time.ANSIC

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

func ping(host string) {
	log.Println("Ping ", host)
	raddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Fatalln("resolve ip addr error", err)
	} else {
		log.Println("IP:", raddr)
	}
	conn, err := net.DialIP("ip4:icmp", nil, raddr)

	if err != nil {
		log.Fatalln(err)
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
	conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
	recv := make([]byte, 100)
	recv_len, err := conn.Read(recv)
	if err != nil {
		log.Fatalln(err)
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
	var host = flag.String("host", "www.baidu.com", "usage: -host=baidu.com")
	flag.Parse()
	for i := 0; i < 2; i++ {
		// ping("125.221.232.253")
		ping(*host)
		time.Sleep(time.Second / 2)
	}
}
