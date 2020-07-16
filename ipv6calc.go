package main

import (
	"fmt"
)

type quadOctet [4]string

func (qo *quadOctet) push(c rune) {

}

func main() {
	fmt.Println("it works!")
	const sample = "\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98"

	fmt.Println(sample)
	for i := 0; i < len(sample); i++ {
		fmt.Printf("%x ", sample[i])
	}
}