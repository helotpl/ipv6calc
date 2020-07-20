package main

import (
	"errors"
	"fmt"
	"strings"
)

type ipv6token [4]byte

type ipv6tokenized []ipv6token

type ipv6addr struct {
	high int64
	low  int64
}

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

func (t ipv6tokenized) String() string {
	s := make([]string, len(t))
	for i := range t {
		s[i] = t[i].String()
	}
	return strings.Join(s, ":")
}

func tokenFromString(s string) *ipv6token {
	t := cleanToken()
	r := []rune(s)

	curr := 3
	//reverse iterate over runes
	for i := len(r) - 1; i >= 0; i-- {
		rune := r[i]
		//skip unicode characters
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
func tokenizeIPv6(s string) (ret []string, e error) {
	parts := strings.Split(s, ":")

	if len(parts) > 8 {
		return parts, errors.New("too many colons in ipv6 address")
	}

	empty, err := findEmptyToken(parts)
	if err != nil {
		return parts, err
	}

	if empty > 0 {
		parts[empty] = "0"
	}
	missing := 8 - len(parts)
	if missing > 0 {
		if empty == -1 {
			return parts, errors.New("not enough colons in ipv6 address without double colon")
		}
		//we can fill up missing tokens
		fill := zeroTokens(missing)
		parts2 := make([]string, 8)
		copy(parts2[:empty], parts[:empty])
		copy(parts2[empty:empty+missing], fill[:missing])
		copy(parts2[empty+missing:], parts[empty:])
		parts = parts2
	}

	return parts, nil
}

func makeTokens(s []string) ipv6tokenized {
	ret := make(ipv6tokenized, len(s))
	for i, v := range s {
		ret[i] = *tokenFromString(v)
	}
	return ret
}

func findEmptyToken(ss []string) (num int, e error) {
	empty := -1
	for i := 1; i < len(ss)-1; i++ {
		if len(ss[i]) == 0 {
			if empty < 0 {
				empty = i
			} else {
				return -1, errors.New("double empty tokens")
			}
		}
	}
	return empty, nil
}

func zeroTokens(num int) []string {
	ret := make([]string, num)
	for i := range ret {
		ret[i] = "0"
	}
	return ret
}

func tokensToByteString(t []ipv6token) []byte {
	ret := make([]byte, len(t)*4)
	for i := range t {
		copy(ret[i*4:], t[i][:])
	}
	return ret
}

func hexToInt(b byte) int64 {
	if b >= '0' && b <= '9' {
		return int64(b) - '0'
	}
	if b >= 'a' && b <= 'f' {
		return int64(b) - 'a'
	}
	if b >= 'A' && b <= 'F' {
		return int64(b) - 'A'
	}
	return 0
}

func makeIPv6Addr(t ipv6tokenized) (i6 ipv6addr, e error) {
	if len(t) != 8 {
		return ipv6addr{0, 0}, errors.New("ipv6tokenized should have length of 8")
	}

	//tmp shut errors
	return ipv6addr{0, 0}, nil
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

	fmt.Println(tokenizeIPv6("342:356:34234:342:23434:3223"))
	fmt.Println(tokenizeIPv6("342:356:34234::23434:3223"))
	fmt.Println(tokenizeIPv6("342:356:34234::23434::3223"))
	fmt.Println(tokenizeIPv6("342:356:34234:::23434:3223"))
	fmt.Println(tokenizeIPv6("342:356:34234::aaa:a:23434:3223"))

	tt, err := tokenizeIPv6("342:356:34234::3223")
	if err == nil {
		fmt.Println(makeTokens(tt).String())
	}
	tt, err = tokenizeIPv6("a::")
	if err == nil {
		fmt.Println(makeTokens(tt).String())
	}

}
