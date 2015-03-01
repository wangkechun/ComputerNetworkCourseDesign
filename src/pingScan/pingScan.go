package main

import (
	"bytes"
	"encoding/binary"
	// "io"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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

func ping(host string, timeLimit int) (re PingReturn) {
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
		if strings.Index(err.Error(), "operation not permitted") != -1 {
			log.Fatalln("operation not permitted, please run it by sudo")
		}
		fmt.Printf("%+v\n", err.Error())
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
	PingLogger.Println("Runing Ping data ",
		printByte(buffer.Bytes()))
	conn.Write(buffer.Bytes())
	t_start := time.Now()
	conn.SetReadDeadline((time.Now().Add(
		time.Duration(timeLimit) * time.Millisecond)))
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

func PingList(hostList []string, waitTime int, timeLimit int) {
	successAlive := make([]PingReturn, 0)
	noRet := make(chan PingReturn, 255)
	var ticker *time.Ticker
	ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf("all:%d over:%d pre:%f\n", len(hostList), len(successAlive), 0.)
			}
		}
	}()
	for _, v := range hostList {
		go func(v string) {
			r := ping(v, timeLimit)
			// print("*")
			noRet <- r
		}(v)
	}
	func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(waitTime)):
				fmt.Println("timeout ", waitTime)
				return
			case r := <-noRet:
				successAlive = append(successAlive, r)
				if len(successAlive) == len(hostList) {
					return
				}
				continue
			}
			break
		}
	}()

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

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func parseTwoInt(s string) (l, r int, err error) {
	switch strings.Count(s, "-") {
	case 0:
		n, _ := strconv.Atoi(s)
		return n, n + 1, nil
	case 1:
		sp := strings.SplitN(s, "-", 2)
		l, _ = strconv.Atoi(sp[0])
		r, _ = strconv.Atoi(sp[1])
		return l, r + 1, nil
	default:
		return 0, 1, errors.New("IP interval illegal ")
	}
}

func parseIPList(IPInterval string) ([]string, error) {
	ip := strings.SplitN(IPInterval, ".", 4)
	ips := make([]string, 0)
	if len(ip) != 4 {
		return ips, errors.New("IP interval illegal ")
	}
	for l0, r0, err := parseTwoInt(ip[0]); l0 < r0; l0++ {
		if err != nil {
			return ips, errors.New("IP interval illegal ")
		}
		for l1, r1, _ := parseTwoInt(ip[1]); l1 < r1; l1++ {
			if err != nil {
				return ips, errors.New("IP interval illegal ")
			}
			for l2, r2, _ := parseTwoInt(ip[2]); l2 < r2; l2++ {
				if err != nil {
					return ips, errors.New("IP interval illegal ")
				}
				for l3, r3, _ := parseTwoInt(ip[3]); l3 < r3; l3++ {
					if err != nil {
						return ips, errors.New("IP interval illegal ")
					}
					now := fmt.Sprintf("%d.%d.%d.%d", l0, l1, l2, l3)
					ips = append(ips, now)
				}
			}
		}
	}
	return ips, nil
}

func init() {
	PingLogger = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	flag.Parse()
}

var ipInterval = flag.String("ip", "10.1.12.1-255", "")
var numRequests = flag.Int("n", 255, "Number of requests to perform")
var timeLimit = flag.Int("t", 3000, "Millisecond of ping timeout")
var waitTime = flag.Int("w", 3, "Second wait after no ans")

func main() {
	fmt.Println(*ipInterval)
	ips, err := parseIPList(*ipInterval)
	if err != nil {
		log.Fatal(err)
	}
	every_limit := *numRequests
	for i := 0; i < len(ips); i += every_limit {
		fmt.Println("now:", ips[i])
		PingList(ips[i:min(i+every_limit, len(ips))], *waitTime, *timeLimit)
	}
}
