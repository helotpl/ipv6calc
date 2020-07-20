package main

import (
	"errors"
	"fmt"
	"strings"
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

	curr := 3
	//reverse iterate over runes
	for i := len(r) - 1; i >= 0; i-- {
		rune := r[i]
		if rune < 256 {
			b := byte(rune)
			if checkHexChar(b) {
				t[curr] = b
				curr--
				if curr < 0 {
					break
				}
			}
		}
	}
	return t
}

//ff = 8bit
//ffff = 16bit
//128/16 = 8 tokens
//function accepts only bare IPv6 address
func tokenizeIPv6(s string) []ipv6token {
	parts := strings.Split(s, ":")
	
	return make([]ipv6token, 0)
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

	t = tokenFromString("869ąłś")
	fmt.Println(t)

	tokenizeIPv6("342:356:34234:342:23434:3223")
}
