package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"log"
	"net"
	"sync"
	"time"
)

const (
	configPkt             = 0
	ackPkt                = 1
	dataPkt               = 2
	respPkt               = 3
	pingPkt               = 0xff
	//rxSize                = 1024 * 1024
	keepAliveTimeoutCount = 3
)

var headerLength = binary.Size(Header{})

type Client struct {
	Name           string
	Address        string
	Concurrent     int
	KeepAlive      uint16 // keepalive interval
	keepAliveCount uint8  // the number of keepalive packet sent
	auth           bool   // client auth status
	data           string
	sendChan       chan string // channel for data to be sent
}

// Packet header
type Header struct {
	Pkt       uint8  // packet type
	Timestamp uint64 // packet timestamp in ms
	Length    uint32 // payload length
}

// Ensure all data are successfully written, or it
// will just block until the receiver can receive it
func WriteAll(conn net.Conn, data string) error {
	_,err := conn.Write([]byte(data))
	return err
}


func safeWriteChannel(ch chan string, data string) {
	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			log.Println("safeWriteChannel:", r)
			return
		}
	}()
	if ch != nil {
		ch <- data
	}
}

func NewClient(name string, address string, concurrent int) *Client {
	return &Client{

		Name:    name,
		Address: address,

		Concurrent: concurrent,
		data:       "{}",
		KeepAlive:  10,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("[%s][Client][%s]", c.Name, c.Address)
}

func (c *Client) Read() string {
	return c.data
}

func (c *Client) Write(data string) {

	safeWriteChannel(c.sendChan, data)
}






// Send thread
// Pull data from the send channel and send it
func (c *Client) send(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	defer close(c.sendChan)
	for {
		select {
		case data := <-c.sendChan:
			if err := WriteAll(conn, data); err != nil {
				log.Println(c, "send:", err)
				return
			}
		case <-ctx.Done():
			log.Println(c, "send:", "Connection lost")
			return
		}
	}
}

// Receive thread
// Receive byte stream from socket until a whole valid packet
// is received. The connection will be closed when any error
// occurred.
func (c *Client) receive(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	// The map that holds last received json data
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(c, "receive:", err)
			return
		}
		c.data = netData
	}
}

// Keepalive Thread
// Send a keepalive packet to server at a certain time interval. Once a packet
// was be sent, increase the counter. When the counter reach at a certain number,
// we believe the connection was lost and close it.
func (c *JSONClient) keepalive(ctx context.Context, cancel context.CancelFunc, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Println(c, "keepalive:", "Connection lost")
			return
		case <-time.After(time.Duration(c.KeepAlive) * time.Second):
			// The keepalive stage should after the auth stage
			if !c.auth {
				break
			}
			if c.keepAliveCount >= keepAliveTimeoutCount {
				log.Println(c, "keepalive: Timed out")
				conn.Close()
				return
			}
			keepalive, err := jsoniter.Marshal(map[string]interface{}{
				"keepalive": c.keepAliveCount, "timeout": keepAliveTimeoutCount, "time": time.Now().Unix()})
			if err != nil {
				log.Println(c, "keepalive:", err)
				break
			}
			var pkt []byte
			if pkt, err = pack(keepalive, pingPkt); err != nil {
				log.Println(c, "keepalive:", err)
				break
			}
			safeWriteChannel(c.sendChan, pkt)
			c.keepAliveCount++
		}
	}
}

func (c *JSONClient) Connect() {
	for {
		conn, err := net.Dial("tcp", c.Address)
		if err != nil {
			log.Println(c, "Connect:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println(c, "Connection established")
		var auth []byte
		if auth, err = authenticate(conn, "guest", ""); err != nil {
			conn.Close()
			continue
		}
		// Send channel
		c.sendChan = make(chan []byte)
		c.keepAliveCount = 0
		c.auth = false

		var wg sync.WaitGroup
		wg.Add(3)
		ctx, cancel := context.WithCancel(context.Background())
		go c.send(ctx, cancel, conn, &wg)
		go c.receive(ctx, cancel, conn, &wg)
		go c.keepalive(ctx, cancel, conn, &wg)
		// Begin authentication
		safeWriteChannel(c.sendChan, auth)
		// Wait for all goroutines to exit
		wg.Wait()
		conn.Close()
		c.sendChan = nil
		c.data = "{}"
		time.Sleep(3 * time.Second)
	}
}

func authenticate(conn net.Conn, user string, password string) (authPkt []byte, err error) {
	auth, err := jsoniter.Marshal(map[string]interface{}{"user": user, "password": password, "groups": []string{}, "flat": true, "compression": false})
	if err != nil {
		return
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	authPkt, err = pack(auth, configPkt)
	return
}
