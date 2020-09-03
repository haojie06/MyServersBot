package main

import (
	"log"
	"net"
)

//状态监控使用udp来通信
func startStatusServer() {
	//为了在使用hostname的情况下也可以正常运行
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

// func printUDPMessage(c *net.UDPConn) {
// 	buf := make([]byte, 512)
// 	len, _, err := c.ReadFromUDP(buf)
// 	checkError(err, "读取出错")

// }

//处理UDP数据
func recvUDPMessage(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	//如果没有数据包传来，线程会在这里阻塞
	n, remoteAddr, err := conn.ReadFromUDP(buf)
	checkError(err, "处理UDP数据包")

	log.Println("接受到UDP数据包:", n, remoteAddr)
	log.Printf("Received data: %s", string(buf[:len(buf)]))
	b := []byte("收到")
	conn.WriteToUDP(b, remoteAddr)
	//开启一个线程处理
}
