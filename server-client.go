package main

import (
"bufio"
"fmt"
"net"
"os"
"reflect"
//"strconv"
"strings"
)


func convert(c string) []byte {
	command := []byte(c)
	content := []byte(c)
	head := make([]byte, 3)
	head[0] = 0xFC
	head[2] = 0x01
	fmt.Println(reflect.TypeOf(0xf1))
	commandl:=len(command)
	head[1] = byte(commandl)
	tail := make([]byte, 1)
	tail[0] = 0xFE
	fuck := append(head, content...)
	com := append(fuck, tail...)
	return com
}




func main() {
	//arguments := os.Args
	//if len(arguments) == 1 {
	//	fmt.Println("Please provide host:port.")
	//	return
	//}p

	CONNECT := "127.0.0.1:7878"
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		msg := convert(text)
		//fmt.Println("send msg ",msg)
		c.Write(msg)

		message, _ := bufio.NewReader(c).ReadString(',')

		fmt.Println("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
