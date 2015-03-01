package main

import (
	"fmt"
	"testing"
)

func Test_1(t *testing.T) {
	s, err := parseIPList("10.1.1-12.1-10")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	if len(s) != 120 {
		t.Fail()
	}
}
func Test_2(t *testing.T) {
	l, r, err := parseTwoInt("100")
	if l != 100 || r != 101 || err != nil {
		t.Fail()
	}
}
func Test_3(t *testing.T) {
	l, r, err := parseTwoInt("100-200")
	if l != 100 || r != 201 || err != nil {
		t.Fail()
	}
}
