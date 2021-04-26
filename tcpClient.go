package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)


func convert(c string) []byte {
	command := []byte(c)
	commandlen := hex.EncodedLen(len(command))
	dst1 := make([]byte, commandlen)
	hex.Encode(dst1, command)
	head := make([]byte, 3)
	head[0] = 0xFC
	head[3] = 0x01
	totallen := int64((commandlen / 2) + 3)
	tl := strconv.FormatInt(totallen, 16)
	btl := []byte(tl)
	head[1] = btl[0]
	tail := make([]byte, 1)
	tail[0] = 0xFE
	fuck := append(head, dst1...)
	com := append(fuck, tail...)
	return com
}


func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide host:port.")
		return
	}

	CONNECT := arguments[1]
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		//convert(text)
		fmt.Fprintf(c, text+"\n")

		message, _ := bufio.NewReader(c).ReadString(',')

		fmt.Println("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}