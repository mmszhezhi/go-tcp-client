package main

import "fmt"

func main() {

	head := make([]byte,2)
	head[0] = 0xFC
	head[1] = 0x01

	fmt.Println(head[0])
}
