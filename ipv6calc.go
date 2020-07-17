package main

import (
	"fmt"
)

type ipv6token []byte

func checkHexChar(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	if b >= 'a' && b <= 'f' {
		return true
	}
	if b >= 'A' && b <= 'F' {
		return true
	}
	return false
}

func (t *ipv6token) pushHexChar(b byte) {
	if checkHexChar(b) {
		*t = append(*t, b)
	}
}

func main() {
	t := new(ipv6token)

	t.pushHexChar('a')
	t.pushHexChar('9')
	t.pushHexChar('x')
	fmt.Println(string(*t))
}
