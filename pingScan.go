package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

var _ = bytes.MinRead
var _ = binary.MaxVarintLen16
var _ = fmt.Scanln
var _ = log.Println
var _ = time.ANSIC
var _ = os.DevNull
var _ = ioutil.NopCloser
var _ = flag.ContinueOnError

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

type PingReturn struct {
	success bool
	msg     string
	host    string
	err     error
}

var PingLogger *log.Logger

func init() {
	PingLogger = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

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

func ping(host string) (re PingReturn) {
	re.success = false
	re.host = host
	PingLogger.Println("Ping ", host)
	raddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		PingLogger.Println("resolve ip addr error", err)
		re.msg = "ip error"
		re.err = err
		return
	} else {
		PingLogger.Println("IP:", raddr)
	}
	conn, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		PingLogger.Println(err)
		re.msg = "dial error"
		re.err = err
		return
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
	PingLogger.Println("Runing Ping data ", printByte(buffer.Bytes()))
	conn.Write(buffer.Bytes())
	t_start := time.Now()
	conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
	recv := make([]byte, 100)
	recv_len, err := conn.Read(recv)
	if err != nil {
		re.msg = "read error"
		re.err = err
		PingLogger.Println(err)
		return
	}
	PingLogger.Println("Recv data ", printByte(recv[:recv_len]))
	t_end := time.Now()
	dur := t_end.Sub(t_start).Nanoseconds() / 1e6
	PingLogger.Println("Time spend ms", dur)
	PingLogger.Println("")
	re.success = true
	defer conn.Close()
	return
}

func printByte(b []byte) (r string) {
	l := len(b)
	for i := 0; i < l; i += 4 {
		r += fmt.Sprint(b[i:i+4], " ")
	}
	return
}

func PingList(hostList []string) {
	successAlive := make([]PingReturn, 0)
	noRet := make(chan PingReturn, 255)
	for _, v := range hostList {
		go func(v string) {
			r := ping(v)
			print("*")
			noRet <- r
		}(v)
	}

	for {
		select {
		case <-time.After(time.Second * 5):
			fmt.Println("timeout 3")
			break
		case r := <-noRet:
			successAlive = append(successAlive, r)
			continue
		}
		break
	}

	var suc, err int
	for _, v := range successAlive {
		if v.success {
			suc++
			fmt.Printf("ip:%s success:%t\n", v.host, v.success)
		} else {
			err++
			// fmt.Println(v.msg, v.err.Error())
		}
	}
	fmt.Printf("###########################\nsuccess:%d error:%d\n", suc, err)

}

func main() {
	for j := 12; j < 13; j++ {
		hosts := make([]string, 0)
		for i := 1; i < 255; i++ {
			host := fmt.Sprintf("10.1.%d.%d", j, i)
			hosts = append(hosts, host)
		}
		PingList(hosts)
	}
}
