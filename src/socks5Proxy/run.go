package main

import "log"

func main() {
	s := &Server{}
	host := "0.0.0.0:8777"
	log.Println("my-socks5-proxy run ", host)
	s.ListenAndServe("tcp", host)
}
