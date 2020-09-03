package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

//状态监控使用udp来通信
func startStatusServer() {
	//如果用了域名方便解析
	addr, err := net.ResolveUDPAddr("udp", "localhost:10086")
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	log.Println("状态监控服务器开启")
	//方法返回（退出后）释放资源，关闭连接
	defer conn.Close()
	//循环监听来自客户端的连接
	//接收到连接之后开启新的线程
	// go handleTcpConnection(conn)
	for {
		recvUDPMessage(conn)
	}
}

//处理UDP数据
func recvUDPMessage(conn *net.UDPConn) {
	log.Println("接收到UDP数据包")
	buf := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(buf)
	checkError(err, "处理UDP数据包")
	timeNow := time.Now().Unix()
	fmt.Println("接受到UDP数据包:", n, remoteAddr)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(timeNow))
	conn.WriteToUDP(b, remoteAddr)
	for {
		buf := make([]byte, 512)
		len, err := conn.Read(buf)
		checkError(err, "读取出错")
		fmt.Printf("Received data: %v", string(buf[:len]))
	}
}
