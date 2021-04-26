package main

import (
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"time"
)

const MIN = 1
const MAX = 100

func random() int {
	return rand.Intn(MAX-MIN) + MIN
}

func convert(c string) []byte {
	command := []byte(c)
	content := []byte(c)
	head := make([]byte, 3)
	head[0] = 0xFC
	head[2] = 0x01
	fmt.Println(reflect.TypeOf(0xf1))
	commandl := len(command)
	head[1] = byte(commandl)
	tail := make([]byte, 1)
	tail[0] = 0xFE
	fuck := append(head, content...)
	com := append(fuck, tail...)
	return com
}

func reverse(msg []byte)string  {
	l :=msg[1]
	response := string(msg[3:3+int(l)])
	return response
}


func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	buf := make([]byte,1024)
	for {
		n,_ := c.Read(buf)

		msgfromclient := reverse(buf[:n])
		//msgfromclient := string(buf[3:n])
		fmt.Printf("client say: %s \n",msgfromclient)
		result := strconv.Itoa(random())
		response := "(@" +msgfromclient +")"  +" 没有感情的聊天机器：" + " " + result + " mfuck you look thity,"

		c.Write(convert(response))
	}
	c.Close()
}

func main() {
	//arguments := os.Args
	////arguments := 8989
	//if len(arguments) == 1 {
	//	fmt.Println("Please provide a port number!")
	//	return
	//}


	PORT := ":7878" // + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}

