package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var _ = bufio.NewReader
var _ = fmt.Printf
var _ = log.Fatalln

type Config struct {
}
type Server struct {
	config *Config
}

func New(conf *Config) (*Server, error) {

	server := &Server{
		config: conf,
	}
	return server, nil

}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
	return nil
}

func (s *Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) ServeConn(conn net.Conn) error {
	defer conn.Close()
	rBuf := bufio.NewReader(conn)
	wBuf := bufio.NewWriter(conn)
	VER, err := rBuf.ReadByte()
	NMETHODS, err := rBuf.ReadByte()

	if err != nil {
		log.Println(err)
		return err
	}
	if VER != byte(5) {
		log.Println("socks5 VER ERROR")
		return nil
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
			return ServerConnAuthSuccess(rBuf, wBuf)
		}
	}
	conn.Write([]byte{5, 0xff})
	log.Println("no accept methods")
	return nil
}

func readMethods(n int, r io.Reader) ([]byte, error) {
	methods := make([]byte, n)
	_, err := io.ReadAtLeast(r, methods, n)
	return methods, err
}

func ServerConnAuthSuccess(rBuf io.Reader, wBuf io.Writer) error {

	return nil
}

func main() {
	s, err := New(&Config{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(s.ListenAndServe("tcp", "0.0.0.0:8777"))

}
