package main

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type ipv6token [4]byte

type ipv6tokenized []ipv6token

type ipv6addr struct {
	high uint64
	low  uint64
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
	maxssind := len(ss)
	for i := 1; i < maxssind-1; i++ {
		if len(ss[i]) == 0 {
			if empty < 0 {
				empty = i
			} else {
				return -1, errors.New("illegal double empty tokens")
			}
		}
	}
	//this checks are unnessesary, because these situations are NOT AMBIGOUS
	//but designers of IPv6 addressing apparently know better
	if len(ss[0]) == 0 && empty != 1 {
		return -1, errors.New("illegal empty token at the start")
	}
	if len(ss[maxssind-1]) == 0 && empty != maxssind-2 {
		return -1, errors.New("illegal empty token at the end")
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

func hexToInt(b byte) uint64 {
	if b >= '0' && b <= '9' {
		return uint64(b) - '0'
	}
	if b >= 'a' && b <= 'f' {
		return uint64(b) - 'a' + 10
	}
	if b >= 'A' && b <= 'F' {
		return uint64(b) - 'A' + 10
	}
	return 0
}

func mergeTokens(t []ipv6token) []byte {
	ret := make([]byte, len(t)*4)
	for i := range t {
		copy(ret[i*4:], t[i][:])
	}
	return ret
}

// 1 hex == 4 bits
// 16 hex == 64 bits
func hexStringToInt(b []byte) (r uint64, e error) {
	if len(b) > 16 {
		return 0, errors.New("hex string too long to fit into int64")
	}
	var ret uint64 = 0

	for _, v := range b {
		ret = ret<<4 + hexToInt(v)
	}

	return ret, nil
}

func (i6 *ipv6addr) asHex() string {
	ret := fmt.Sprintf("%016x%016x", i6.high, i6.low)
	return ret
}

func retokenize(s string) string {
	r := []rune(s)
	l := len(r)
	start := l % 4
	o := make([]rune, l+(l-start)/4)

	offset := 0
	if start > 0 {
		copy(o[:start], r[:start])
		o[start] = ':'
		offset++
	}
	for i := start; i < l; i += 4 {
		copy(o[i+offset:i+offset+4], r[i:i+4])
		if i+4 < l {
			o[i+offset+4] = ':'
			offset++
		}
	}

	return string(o)
}

func toHexToken(num uint64, token int, leadingZeros bool) string {
	num = (num >> (token * 16)) & 0xFFFF
	if leadingZeros {
		return fmt.Sprintf("%04x", num)
	}
	return fmt.Sprintf("%x", num)
}

func (i6 *ipv6addr) asHexToken(token int, leadingZeros bool) string {
	if token > 3 {
		return toHexToken(i6.high, token-4, leadingZeros)
	}
	return toHexToken(i6.low, token, leadingZeros)

}

func (i6 *ipv6addr) asBigInt() *big.Int {
	var h, l, ret big.Int

	h.SetUint64(i6.high)
	l.SetUint64(i6.low)

	ret.Lsh(&h, 64)
	ret.Add(&ret, &l)
	return &ret
}

type zeros struct {
	start uint
	stop  uint
}

func (z *zeros) count() uint {
	return z.stop - z.start
}

func findBestZeros(zz []zeros) zeros {
	best := -1
	bestLen := uint(0)
	for i, z := range zz {
		currLen := z.count()
		if currLen > bestLen {
			best = i
			bestLen = currLen
		}
	}
	if best > -1 {
		return zz[best]
	}
	return zeros{0, 0}
}

func findZerosInTokens(s []string) []zeros {
	ret := make([]zeros, 0, 10)
	inside := false
	start := uint(0)
	maxi := uint(0)
	for i := range s {
		maxi = uint(i)
		if s[i] == "0" {
			if !inside {
				inside = true
				start = uint(i)
			}
		} else {
			if inside {
				inside = false
				ret = append(ret, zeros{start, uint(i)})
			}
		}
	}
	if inside && start != maxi {
		ret = append(ret, zeros{start, maxi + 1})
	}
	return ret
}

func removeZeroTokens(s []string) []string {
	z := findBestZeros(findZerosInTokens(s))
	if z.start == 0 && z.stop == 0 {
		return s
	}
	if z.start == 0 {
		z.start = 1
		s[0] = ""
	}
	if z.stop == 8 {
		z.stop = 7
		s[7] = ""
	}
	newLen := 8 - z.count() + 1
	newS := make([]string, newLen)
	copy(newS[:z.start], s[:z.start])
	copy(newS[z.start+1:], s[z.stop:])
	return newS
}

func (i6 *ipv6addr) StringTokens(leadingZeros bool) []string {
	s := make([]string, 8)
	for i := range s {
		s[i] = i6.asHexToken(7-i, leadingZeros)
	}
	return s
}

func (i6 ipv6addr) String() string {
	s := i6.StringTokens(false)
	s = removeZeroTokens(s)
	return strings.Join(s, ":")
}

func (i6 *ipv6addr) LongString() string {
	s := i6.StringTokens(true)
	return strings.Join(s, ":")
}

func makeIPv6Addr(t ipv6tokenized) (i6 ipv6addr, e error) {
	if len(t) != 8 {
		return ipv6addr{0, 0}, errors.New("ipv6tokenized should have exactly 8 tokens")
	}

	high, err := hexStringToInt(mergeTokens(t[0:4]))
	if err != nil {
		return ipv6addr{0, 0}, err
	}
	low, err := hexStringToInt(mergeTokens(t[4:8]))
	if err != nil {
		return ipv6addr{0, 0}, err
	}

	return ipv6addr{high, low}, nil
}

//IPv6Addr ...
func makeIPv6AddrFromString(s string) (i6 ipv6addr, e error) {
	t, err := tokenizeIPv6(s)
	if err != nil {
		return ipv6addr{0, 0}, err
	}
	tt := makeTokens(t)
	return makeIPv6Addr(tt)
}

func makeIPv6AddrFromMask(mask uint) (i6 ipv6addr, e error) {
	if mask > 128 || mask < 0 {
		return ipv6addr{}, errors.New("incorrect mask")
	}
	return ipv6addr{}, nil
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

	tests := []string{"342:356:34234::3223",
		"0:a::",
		"0:a::f",
		"23:33:ffff::0:",
		"aa::1:0:0:0:1",
		"a:a:a:a:a:a:a:a",
		"FFFF:ffff:ffff::",
		"::"}
	for _, x := range tests {
		fmt.Print("Input: ")
		fmt.Println(x)
		i6, err := makeIPv6AddrFromString(x)
		if err != nil {
			fmt.Println(err)
		} else {
			s := i6.StringTokens(false)
			fmt.Print("Println(findZerosInTokens(s)): ")
			fmt.Println(findZerosInTokens(s))
			fmt.Print("Println(findBestZeros(findZerosInTokens(s))): ")
			fmt.Println(findBestZeros(findZerosInTokens(s)))
			fmt.Print("Println(i6): ")
			fmt.Println(i6)
			fmt.Print("Println(i6.asHex()): ")
			fmt.Println(i6.asHex())
			fmt.Print("Println(i6.asBigInt()): ")
			fmt.Println(i6.asBigInt())
			fmt.Print("Println(i6.LongString()): ")
			fmt.Println(i6.LongString())
		}
	}

}
