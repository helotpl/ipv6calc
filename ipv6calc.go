package main

import (
	"errors"
	"fmt"
)

type ipv6token [4]byte

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

func cleanToken() *ipv6token {
	t := new(ipv6token)
	copy(t[:], "0000"[:])
	return t
}

func (t *ipv6token) pushHexChar(b byte) error {
	if checkHexChar(b) {
		copy(t[0:3], t[1:4])
		t[3] = b
	} else {
		return errors.New("not a hex char")
	}
	return nil
}

func (t *ipv6token) fillHexChar(b byte) error {
	if checkHexChar(b) {

	} else {
		return errors.New("not a hex char")
	}
	return nil
}

func (t *ipv6token) String() string {
	return string(t[:])
}

func tokenFromString(s string) *ipv6token {
	t := cleanToken()
	r := []rune(s)
	for _, v := range r {
		if v < 256 {
			t.pushHexChar(byte(v))
		}
	}
	return t
}

func main() {
	t := cleanToken()

	t.pushHexChar('a')
	t.pushHexChar('9')
	t.pushHexChar('x')
	// for i := 0; i < 10; i++ {
	// 	t.pushHexChar('0' + byte(i))
	// }
	fmt.Println(t)

	t = tokenFromString("3553869ąłś")
	fmt.Println(t)
}
