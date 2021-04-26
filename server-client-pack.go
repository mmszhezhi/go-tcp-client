package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"time"

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
	commandl := len(command) -1
	head[1] = byte(commandl)
	tail := make([]byte, 1)
	tail[0] = 0xFE
	fuck := append(head, content...)
	com := append(fuck, tail...)
	return com
}

const rxSize = 1024 * 1024

const (
	Idle      = 0
	Received  = 1
	Tasking   = 2
	Exception = 3
	Done      = 4
)

type Client struct {
	Address      string
	data         string
	sendChan     chan []byte
	responseChan chan []byte
	Status       int
}

func NewClient(address string) *Client {
	return &Client{
		Address: address,
		Status:  Idle,
	}
}

func (c *Client) send(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	defer close(c.sendChan)
	for {
		select {
		case data := <-c.sendChan:
			conn.Write(data)
		}
	}

}

func reverse(msg []byte) string {
	l := msg[1]
	response := string(msg[3 : 3+int(l)])
	return response
}

func (c *Client) receive(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()

	// The buffer for one Read call
	readBuff := make([]byte, rxSize)
	for {
		n, err := conn.Read(readBuff)
		if err != nil {
			log.Println(c, "Write:", err)
			return
		}
		result := reverse(readBuff[:n])
		if result == "received" {
			c.Status = Received
		}
		c.data = result
	}

}

func (c *Client) Post(command string) {
	if c.Status == Idle || c.Status == Done {
		comm := convert(command)
		c.sendChan <- comm
	}
}

func (c *Client) Connect() {
	for {
		conn, err := net.Dial("tcp", c.Address)
		if err != nil {
			fmt.Printf("connect to %s failed!", c.Address)
		}

		c.sendChan = make(chan []byte)

		var wg sync.WaitGroup
		wg.Add(2)
		ctx, cancel := context.WithCancel(context.Background())
		go c.send(ctx, cancel, conn, &wg)
		go c.receive(ctx, cancel, conn, &wg)

		wg.Wait()
		conn.Close()
		c.sendChan = nil
		c.data = ""
		time.Sleep(3 * time.Second)

	}
}

func main() {

	CONNECT := "127.0.0.1:7878"
	client := &Client{
		Address: CONNECT,
	}
	go client.Connect()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		client.Post(text)
		//message, _ := bufio.NewReader(c).ReadString(',')
		time.Sleep(2*time.Second)
		fmt.Println("->: " + client.data)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
