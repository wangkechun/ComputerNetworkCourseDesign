package main

import (
	"bufio"
	"fmt"
	"github.com/k0kubun/pp"
	"io"
	"log"
	"net"
	"time"
)

var _ = bufio.NewReader
var _ = fmt.Printf
var _ = log.Fatalln
var _ = pp.Fatalln

type Server struct {
}

func (s *Server) Serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go s.ServeConn(conn)
		}
	}
}

func (s *Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	s.Serve(l)
	return nil
}

func (s *Server) ServeConn(conn net.Conn) {
	defer func() {
		if p := recover(); p != nil {
			log.Println(p)
		}
	}()
	defer conn.Close()

	rBuf := bufio.NewReader(conn)
	wBuf := bufio.NewWriter(conn)
	VER, err := rBuf.ReadByte()
	NMETHODS, err := rBuf.ReadByte()

	if err != nil {
		log.Println(err)
		return
	}
	if VER != byte(5) {
		log.Println("socks5 VER ERROR")
		return
	}
	log.Printf("Have %d NMETHODS\n", NMETHODS)
	methods, err := readMethods(int(NMETHODS), rBuf)
	if err != nil {
		log.Println("readMethods error", err)
	} else {
		log.Println("methods:", methods)
	}
	for _, method := range methods {
		if method == 0 {
			conn.Write([]byte{5, 0})
			ServerConnAuthSuccess(rBuf, wBuf)
			return
		}
	}
	conn.Write([]byte{5, 0xff})
	log.Println("no accept methods")
}

func readMethods(n int, r io.Reader) ([]byte, error) {
	methods := make([]byte, n)
	_, err := io.ReadAtLeast(r, methods, n)
	return methods, err
}

type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

const (
	connectCommand   = uint8(1)
	bindCommand      = uint8(2)
	associateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

func ServerConnAuthSuccess(rBuf *bufio.Reader, wBuf *bufio.Writer) {
	header := []byte{0, 0, 0}
	rBuf.Read(header)
	if header[0] != 5 {
		log.Println("Unsupported command version")
	}
	addrType, _ := rBuf.ReadByte()
	d := &AddrSpec{}
	switch addrType {
	case ipv4Address:
		addr := make([]byte, 4)
		rBuf.Read(addr)
		d.IP = net.IP(addr)
		log.Println("ipv4")
	case ipv6Address:
		addr := make([]byte, 4)
		rBuf.Read(addr)
		d.IP = net.IP(addr)
		log.Println("ipv6")
	case fqdnAddress:
		addrLen, _ := rBuf.ReadByte()
		fqdn := make([]byte, addrLen)
		rBuf.Read(fqdn)
		d.FQDN = string(fqdn)
		log.Println("host", d.FQDN)
	default:
		log.Println("unknow addr")
		return
	}
	port := []byte{0, 0}
	rBuf.Read(port)
	d.Port = (int(port[0]) << 8) | int(port[1])
	// pp.Println(d, string(d.IP.String()))
	if d.FQDN != "" {
		addr, err := net.ResolveIPAddr("ip", d.FQDN)
		if err != nil {
			log.Println("dns resolve error", d.FQDN, addr.String())
			return
		}
		d.IP = addr.IP
	}
	switch header[1] {
	case connectCommand:
		log.Println("connectCommand")
		handleConnect(rBuf, wBuf, d)
	// case bindCommand:
	// case associateCommand:
	default:
		log.Println("unsupported command")
	}
}

const (
	successReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

func sendReply(wBuf *bufio.Writer, resp uint8, addr *AddrSpec) error {
	var addrType uint8
	var addrBody []byte
	var addrPort uint16
	switch {
	case addr == nil:
		addrType = ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0
	case addr.FQDN != "":
		addrType = fqdnAddress
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)
	case addr.IP.To4() != nil:
		addrType = ipv4Address
		addrBody = []byte(addr.IP.To4())
		addrPort = uint16(addr.Port)
	case addr.IP.To16() != nil:
		addrType = ipv6Address
		addrBody = []byte(addr.IP.To16())
		addrPort = uint16(addr.Port)
	default:
		log.Println("failed to format address")
		return nil
	}
	msg := make([]byte, 6+len(addrBody))
	msg[0] = 5
	msg[1] = resp
	msg[2] = 0
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)
	_, err := wBuf.Write(msg)
	wBuf.Flush()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func handleConnect(rBuf *bufio.Reader, wBuf *bufio.Writer, dest *AddrSpec) {
	log.Println("handleConnect")
	addr := net.TCPAddr{IP: dest.IP, Port: dest.Port}
	target, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		log.Println(err)
	}
	defer target.Close()
	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	sendReply(wBuf, successReply, &bind)
	errCh := make(chan error, 2)
	go proxy("target", target, rBuf, errCh)
	go proxy("client", wBuf, target, errCh)
	select {
	case _ = <-errCh:
		return
	}
}

func proxy(name string, dst io.Writer, src io.Reader, errCh chan error) {
	n, err := io.Copy(dst, src)
	log.Printf("Copied %d bytes to %s", n, name)
	time.Sleep(10 * time.Millisecond)
	errCh <- err
}

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}
