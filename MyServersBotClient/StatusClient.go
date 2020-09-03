package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("udp", "localhost:10086")
	defer conn.Close()
	if err != nil {
		os.Exit(1)
	}

	conn.Write([]byte("Hello world!"))

	fmt.Println("send msg")

	var msg [20]byte
	conn.Read(msg[0:])

	fmt.Println("msg is", string(msg[0:10]))
}
