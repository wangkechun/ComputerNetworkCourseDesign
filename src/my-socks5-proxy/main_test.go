package main

import (
	"log"
	"os/exec"
	"strings"
	"testing"
)

var _ = log.Ldate

func Test_1(t *testing.T) {
	s := &Server{}
	host := "0.0.0.0:8777"
	log.Println("my-socks5-proxy run ", host)
	go func() {
		err := s.ListenAndServe("tcp", host)
		if err != nil {
			log.Println(err)
		}
	}()
	c := exec.Command("sh", "-c", "curl --socks5 localhost:8777 baidu.com")
	d, err := c.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	if strings.Index(string(d), "http://www.baidu.com/") == -1 {
		t.Error("curl result not match")
		log.Println(string(d)) //输出和command里边的output一样哈
	}
}
