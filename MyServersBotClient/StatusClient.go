package main

import (
	"log"
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
	log.Println("send msg")
	msg := make([]byte, 1024)
	//为什么在发送数据包之后，还能收到服务端的写入（这里还有个READ的方法，会等待。。） —— 对于打开的socket，到达deadline之前，端口会一直开启 防火墙/NAT会维护UDP连接的连接表，超时后就会移除
	//conn.SetDeadline(time.Now().Add(time.Duration(5 * time.Second)))
	_, err = conn.Read(msg)
	if err != nil {
		log.Panic(err.Error())
	}
	log.Printf("msg is %s \n", msg)
}
