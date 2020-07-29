package main

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

type ipv6token [4]byte

type ipv6tokenized []ipv6token

type ipv6addr struct {
	high uint64
	low  uint64
}

type ipv6prefix struct {
	addr     ipv6addr
	mask     uint
	addrMask *ipv6addr
}

type exposeChar struct {
	before   bool //if not before than after
	position uint
	char     rune
}

type hexPrintConfig struct {
	upcase          bool
	leasingZeros    bool
	exposeStartChar *rune
	exposeEndChar   *rune
}

func makeDefaultHexPrintConfig() hexPrintConfig {
	l := leftExposeRune
	r := rightExposeRune
	return hexPrintConfig{false, false, &l, &r}
}

const leftExposeChar = '<'
const rightExposeChar = '>'
const leftExposeRune = rune(leftExposeChar)
const rightExposeRune = rune(rightExposeChar)

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

func (e *exposeInToken) minZerosFromExpose() uint {
	if e.empty {
		return 0
	}
	if e.contLeft {
		return 4
	}
	return 4 - e.start
}

func toHexTokenExpose(num uint64, token int, localExp exposeInToken) string {
	num = (num >> (token * 16)) & 0xFFFF
	minZeros := localExp.minZerosFromExpose()
	var format string
	if minZeros > 0 {
		format = fmt.Sprintf("%%0%vx", minZeros)
	} else {
		format = "%x"
	}
	ret := fmt.Sprintf(format, num)
	//add := uint(0)
	if !localExp.empty {
		if !localExp.contLeft {
			space := localExp.start
			//if !(token == 0 && space == 0) {
			space += uint(len(ret)) - 4
			ret = ret[:space] + string(leftExposeChar) + ret[space:]
			//add++
			//}
		}
		if !localExp.contRight {
			space := localExp.stop
			//if !(token == 7 && space == 4) {
			space += uint(len(ret)) - 4 + 1
			ret = ret[:space] + string(rightExposeChar) + ret[space:]
			//}
		}
	}
	return ret
}

func toHexTokenMultiExpose(num uint64, token int, localExposes []exposeChar) string {
	num = (num >> (token * 16)) & 0xFFFF
	minZeros := uint(0)
	for i := range localExposes {
		nminZeros := 4 - localExposes[i].position
		if nminZeros > minZeros {
			minZeros = nminZeros
		}
	}
	var format string
	if minZeros > 0 {
		format = fmt.Sprintf("%%0%vx", minZeros)
	} else {
		format = "%v"
	}
	ret := fmt.Sprintf(format, num)
	for _, v := range localExposes {
		pos := uint(len(ret)) - 4 + v.position
		if v.before == false {
			pos++
		}
		ret = ret[:pos] + string(v.char) + ret[pos:]
	}
	return ret
}

func (i6 *ipv6addr) asHexToken(token int, leadingZeros bool) string {
	if token > 3 {
		return toHexToken(i6.high, token-4, leadingZeros)
	}
	return toHexToken(i6.low, token, leadingZeros)
}

func (e exposeInToken) localExpose(token int) *exposeInToken {
	if e.empty {
		return &e
	}
	e.start -= uint(token) * 4
	e.stop -= uint(token) * 4
	return &e
}

func (i6 *ipv6addr) asHexTokenExpose(token int, e exposeInToken) string {
	if token > 3 {
		return toHexTokenExpose(i6.high, token-4, e)
	}
	return toHexTokenExpose(i6.low, token, e)
}

func (i6 *ipv6addr) asBigInt() *big.Int {
	var h, l, ret big.Int

	h.SetUint64(i6.high)
	l.SetUint64(i6.low)

	ret.Lsh(&h, 64)
	ret.Add(&ret, &l)
	return &ret
}

func (i6 *ipv6addr) And(i *ipv6addr) *ipv6addr {
	nh := i6.high & i.high
	nl := i6.low & i.low
	return &ipv6addr{nh, nl}
}

func (i6 *ipv6addr) Or(i *ipv6addr) *ipv6addr {
	nh := i6.high | i.high
	nl := i6.low | i.low
	return &ipv6addr{nh, nl}
}

func (i6 *ipv6addr) Neg() *ipv6addr {
	return &ipv6addr{^i6.high, ^i6.low}
}

func (i6 *ipv6addr) Xor(i *ipv6addr) *ipv6addr {
	nh := i6.high ^ i.high
	nl := i6.low ^ i.low
	return &ipv6addr{nh, nl}
}

func (i6 *ipv6addr) CummulativeXor(i1, i2 *ipv6addr) *ipv6addr {
	return i6.Or(i1.Xor(i2))
}

func (i6 *ipv6addr) BitsRange() (start, stop uint) {
	highBits := bits.OnesCount64(i6.high)
	lowBits := bits.OnesCount64(i6.low)

	if highBits > 0 {
		start = uint(bits.LeadingZeros64(i6.high))
	} else {
		if lowBits > 0 {
			start = uint(bits.LeadingZeros64(i6.low) + 64)
		} else {
			start = 0
		}
	}
	if lowBits > 0 {
		stop = uint(127 - bits.TrailingZeros64(i6.low))
	} else {
		if highBits > 0 {
			stop = uint(63 - bits.TrailingZeros64(i6.high))
		} else {
			stop = 0
		}
	}
	return start, stop
}

func (i6 *ipv6addr) Inc() *ipv6addr {
	nl := i6.low + 1
	//check carry
	if i6.low > nl {
		//carry to high
		nh := i6.high + 1
		//check carry
		if i6.high > nh {
			//carry with high
			return nil
		} else {
			return &ipv6addr{nh, nl}
		}
	} else {
		return &ipv6addr{i6.high, nl}
	}
}

func (i6 *ipv6addr) Dec() *ipv6addr {
	nl := i6.low - 1
	if i6.low < nl {
		//carry, borrow from high
		nh := i6.high - 1
		//check carry
		if i6.high < nh {
			//carry on hight
			return nil
		} else {
			return &ipv6addr{nh, nl}
		}
	} else {
		return &ipv6addr{i6.high, nl}
	}
}

type zeros struct {
	start uint
	stop  uint
}

type exposeInToken struct {
	empty     bool
	start     uint
	contLeft  bool
	stop      uint
	contRight bool
}

func (e exposeInToken) String() string {
	if e.empty {
		return "empty"
	}
	var l, r string
	if e.contLeft {
		l = "<-"
	} else {
		l = strconv.FormatUint(uint64(e.start), 10)
	}
	if e.contRight {
		r = "->"
	} else {
		r = strconv.FormatUint(uint64(e.stop), 10)
	}
	return strings.Join([]string{l, r}, ";")
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
				if start+1 < uint(i) {
					ret = append(ret, zeros{start, uint(i)})
				}
			}
		}
	}
	if inside && start != maxi {
		ret = append(ret, zeros{start, maxi + 1})
	}
	return ret
}

func findZerosInTokensExpose(s []string, exposeTokens []exposeInToken) []zeros {
	ret := make([]zeros, 0, 10)
	inside := false
	start := uint(0)
	maxi := uint(0)
	for i := range s {
		maxi = uint(i)
		if s[i] == "0" && exposeTokens[i].empty {
			if !inside {
				inside = true
				start = uint(i)
			}
		} else {
			if inside {
				inside = false
				if start+1 < uint(i) {
					ret = append(ret, zeros{start, uint(i)})
				}
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

func removeZeroTokensExpose(s []string, exposeTokens []exposeInToken) []string {
	z := findBestZeros(findZerosInTokensExpose(s, exposeTokens))
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

func tokenizeExpose(exposeHexStart, exposeHexEnd uint) []exposeInToken {
	s := make([]exposeInToken, 8)
	for i := range s {
		hS := uint(i) * 4
		hE := uint(i)*4 + 3
		s[i] = exposeInToken{}
		if exposeHexEnd < hS || exposeHexStart > hE {
			s[i].empty = true
		} else {
			s[i].empty = false
			if exposeHexStart < hS {
				s[i].contLeft = true
			} else {
				s[i].contLeft = false
				s[i].start = exposeHexStart
			}
			if hE < exposeHexEnd {
				s[i].contRight = true
			} else {
				s[i].contRight = false
				s[i].stop = exposeHexEnd
			}
		}
	}
	if leftExposeChar == rightExposeChar {
		if s[0].empty == false && s[0].start == 0 {
			s[0].contLeft = true
		}
		if s[7].empty == false && s[7].stop == 3 {
			s[7].contRight = true
		}
	}
	return s
}

func tokenizeMultiExpose(e []exposeChar) [][]exposeChar {
	r := make([][]exposeChar, 8)
	for i := range e {
		toknum := e[i].position / 4
		r[toknum] = append(r[toknum], e[i])
	}
	return r
}

//BitToHexNum ...
// 0 ... 3 = hex 0
// 4 ... 7 = hex 1 .....
// total bits = 128, total hex = 4*8 = 32...
// 124 ... 127 = hex 31
func BitToHexNum(bit uint) uint {
	return bit / 4
}

//bits are counted as mask, end bit is +1, works as array index
//for example start = 10, end = 11 means that only 10 bit is exposed
func (i6 *ipv6addr) StringTokensExpose(exposeTokens []exposeInToken) []string {
	s := i6.StringTokens(false)
	for i := range s {
		s[i] = i6.asHexTokenExpose(7-i, *(exposeTokens[i].localExpose(i)))
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

func (i6 *ipv6addr) ExposeString(exposeBitStart, exposeBitEnd uint) string {
	es := BitToHexNum(exposeBitStart)
	ee := BitToHexNum(exposeBitEnd)

	te := tokenizeExpose(es, ee)

	s := i6.StringTokensExpose(te)
	s = removeZeroTokensExpose(s, te)
	return strings.Join(s, ":")
}

func (i6 *ipv6addr) MultiExposeString(exposes []exposeChar) string {
	s := i6.StringTokens(false)
	s = removeZeroTokens(s)
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

func makeIPv6AddrFromString2(s string) (i6 *ipv6addr, e error) {
	ss := strings.Split(s, ":")
	if len(ss) > 8 {
		return nil, errors.New("too many colons in address")
	}
	empty := false
	if ss[0] == "" {
		ss[0] = "0"
	}
	if ss[len(ss)-1] == "" {
		ss[len(ss)-1] = "0"
	}
	for _, v := range ss {
		if len(v) == 0 {
			if empty {
				return nil, errors.New("detected more than one double colon")
			}
			empty = true
		}
	}
	if empty == false && len(ss) != 8 {
		return nil, errors.New("too short address when there is no double colon")
	}
	addr := ipv6addr{0, 0}
	for i := 0; i < len(ss); i++ {
		if ss[i] == "" {
			break
		}
		a, err := strconv.ParseUint(ss[i], 16, 16)
		if err != nil {
			return nil, err
		}
		if i < 4 {
			addr.high = addr.high + (a << (16 * (3 - i)))
		} else {
			addr.low = addr.low + (a << (16 * (7 - i)))
		}
	}
	//there is an doublecolon (empty token) we have to walk address in reverse order
	if empty {
		for i := 0; i < len(ss); i++ {
			ssind := len(ss) - i - 1
			if ss[ssind] == "" {
				break
			}
			a, err := strconv.ParseUint(ss[ssind], 16, 16)
			if err != nil {
				return nil, err
			}
			if i > 3 {
				addr.high = addr.high + (a << (16 * (i - 4)))
			} else {
				addr.low = addr.low + (a << (16 * i))
			}
		}
	}
	return &addr, nil
}

func makeIPv6AddrFromMask(mask uint) (i6 ipv6addr, e error) {
	if mask > 128 || mask < 0 {
		return ipv6addr{}, errors.New("incorrect mask")
	}
	//high
	var h uint64
	if mask >= 64 {
		h = 0xFFFFFFFFFFFFFFFF
	} else {
		h = 0xFFFFFFFFFFFFFFFF << (64 - mask)
	}
	//low
	var l uint64
	if mask <= 64 {
		l = 0
	} else {
		l = 0xFFFFFFFFFFFFFFFF << (128 - mask)
	}
	return ipv6addr{h, l}, nil
}

func makeIPv6PrefixFromString(s string) (prefix *ipv6prefix, e error) {
	ss := strings.Split(s, "/")
	if len(ss) > 2 {
		return nil, errors.New("too many / in prefix")
	}
	var mask uint64
	if len(ss) == 2 {
		mask, e = strconv.ParseUint(ss[1], 10, 32)
		if e != nil {
			return nil, e
		}
		if mask > 128 {
			return nil, errors.New("mask is too long")
		}

	} else {
		mask = 128
	}
	i6, err := makeIPv6AddrFromString2(ss[0])
	if err != nil {
		return nil, err
	}
	return &ipv6prefix{*i6, uint(mask), nil}, nil
}

func (p *ipv6prefix) getAddrMask() *ipv6addr {
	if p.addrMask == nil {
		am, err := makeIPv6AddrFromMask(p.mask)
		if err == nil {
			p.addrMask = &am
		}
	}
	return p.addrMask
}

func (p *ipv6prefix) firstAddressFromSubnet() *ipv6addr {
	return p.addr.And(p.getAddrMask())
}

func (p *ipv6prefix) lastAddressFromSubnet() *ipv6addr {
	return p.addr.Or(p.getAddrMask().Neg())
}

func (p *ipv6prefix) nextPrefix() *ipv6prefix {
	nextaddr := p.lastAddressFromSubnet().Inc()
	if nextaddr == nil {
		return nil
	}
	return &ipv6prefix{*nextaddr, p.mask, nil}
}

func (p *ipv6prefix) makeSubnetAddress() *ipv6prefix {
	p.addr = *p.firstAddressFromSubnet()
	return p
}

func (p *ipv6prefix) prevPrefix() *ipv6prefix {
	prevaddr := p.firstAddressFromSubnet().Dec()
	if prevaddr == nil {
		return nil
	}
	newprefix := ipv6prefix{*prevaddr, p.mask, nil}
	return newprefix.makeSubnetAddress()
}

func (p *ipv6prefix) String() string {
	return fmt.Sprintf("%v/%v", p.addr, p.mask)
}

func (p *ipv6prefix) SubnetString() string {
	return fmt.Sprintf("%v/%v", p.firstAddressFromSubnet(), p.mask)
}

func (p *ipv6prefix) ExposeString(exposeBitStart, exposeBitEnd uint) string {
	return fmt.Sprintf("%v/%v", p.addr.ExposeString(exposeBitStart, exposeBitEnd), p.mask)
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

	tests := []string{"342:356:4234::3223",
		"0:a::",
		"0:a::f",
		"23:33:ffff::0:",
		"aa::1:0:0:0:1",
		"a:a:a:a:a:a:a:a",
		"a:0:a:0:a:0:a:0",
		"FFFF:ffff:ffff::",
		"::"}
	testmasks := []uint{1, 5, 64, 100, 126}
	for _, x := range tests {
		fmt.Print("Input: ")
		fmt.Println(x)
		i6, err := makeIPv6AddrFromString2(x)
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
			for _, mask := range testmasks {
				m, _ := makeIPv6AddrFromMask(mask)
				fmt.Printf("ANDing with mask %v : ", mask)
				fmt.Println(*(i6.And(&m)))
				fmt.Print("ORing with negmask: ")
				fmt.Println(*(i6.Or((&m).Neg())))
			}
		}
	}
	for i := 0; i <= 128; i += 4 {
		fmt.Printf("mask %v: ", i)
		m, err := makeIPv6AddrFromMask(uint(i))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Print(m)
			fmt.Print(" neg: ")
			fmt.Println(*(m.Neg()))
		}
	}
	tests2 := []string{"342:356:4234::3223/120",
		"0:a::/40",
		"0:a::f/32",
		"23:33:ffff::0:/48",
		"aa::1:0:0:0:1",
		"a:a:a:a:a:a:a:a/0",
		"a:0:a:0:a:0:a:0/13",
		"FFFF:ffff:ffff::/30",
		"::/39"}
	for _, x := range tests2 {
		fmt.Print("Input: ")
		fmt.Println(x)
		p, err := makeIPv6PrefixFromString(x)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(p)
			fmt.Print("First: ")
			fmt.Println(p.firstAddressFromSubnet())
			fmt.Print("Last:  ")
			fmt.Println(p.lastAddressFromSubnet())
			pp := p.prevPrefix()
			if pp != nil {
				fmt.Print("Prev:  ")
				fmt.Println(pp)
			} else {
				fmt.Println("No prev.")
			}

			np := p.nextPrefix()
			if np != nil {
				fmt.Print("Next:  ")
				fmt.Println(np)
			} else {
				fmt.Println("No next.")
			}
		}
	}
	test4, _ := makeIPv6AddrFromString("0:a::")
	for i := 58; i < 68; i += 1 {
		for j := i; j < 68; j += 1 {
			hi := BitToHexNum(uint(i))
			hj := BitToHexNum(uint(j))
			fmt.Printf("i: %v, j: %#v, hi: %v, hj: %v -> ", i, j, hi, hj)
			//te := tokenizeExpose(hi, hj)
			//fmt.Printf("%+v\n", te)
			fmt.Println(test4.ExposeString(uint(i), uint(j)))
			// for i := range te {
			// 	te[i] = *(te[i].localExpose(i))
			// }
			// fmt.Printf("i: %v, j: %#v, hi: %v, hj: %v -> ", i, j, hi, hj)
			// fmt.Printf("%+v\n", te)
		}
	}
	test5, _ := makeIPv6PrefixFromString("0:1:1:1::/64")
	cum := &ipv6addr{}
	fmt.Printf("Starting from: %v\n", test5)
	addrs := make([]*ipv6prefix, 0, 21)
	for i := 0; i < 20; i++ {
		test5n := test5.nextPrefix()
		cum = cum.CummulativeXor(&test5.addr, &test5n.addr)
		test5 = test5n
		addrs = append(addrs, test5)
	}
	fmt.Printf("Cummulative XOR: %v\n", cum)
	start, end := cum.BitsRange()
	fmt.Printf("Bits from CumulativeXOR: %v - %v\n", start, end)
	for i, v := range addrs {
		fmt.Printf("Next %02v: %v", i, v.ExposeString(start, end))
		if i > 0 {
			fmt.Printf(" bitdiff: %v\n", (addrs[i-1].addr.Xor(&v.addr)).LongString())
		} else {
			fmt.Printf("\n")
		}
	}

	e1 := []exposeChar{exposeChar{true, 2, '_'}, exposeChar{false, 1, '^'}}
	fmt.Println(toHexTokenMultiExpose(uint64(1024), 0, e1))
}
