package main

import (
	"fmt"
	"time"
)

func put(msg chan int,num int)  {
	for{
		msg <-num
	}

}

func main() {
	msg := make(chan int)
	go put(msg,1)
	go put(msg,0)
	for{
		id := <-msg
		time.Sleep(1*time.Second)
		fmt.Println(id)
	}


}
