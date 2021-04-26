package tcpClient

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	Address      string
	data         string
	sendChan     chan []byte
	responseChan chan []byte
}

func NewClient(address string) *Client {
	return &Client{
		Address: address,
	}
}

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

func (c *Client) receive(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	// The map that holds last received json data
	data := make(map[string]interface{})
	// The buffer for holding data that need to be process
	buffer := make([]byte, 0)
	// The buffer for one Read call
	readBuff := make([]byte, rxSize)
	for {
		n, err := conn.Read(readBuff)
	}

}

func (c *Client) Post(command string) {
	comm := convert(command)
	c.sendChan <- comm

}

func (c *Client) Connect() {
	for {
		conn, err := net.Dial("tcp", c.Address)
		if err != nil {
			fmt.Printf("connect to %s failed!", c.Address)
		}
		conn.Write()
	}
}
