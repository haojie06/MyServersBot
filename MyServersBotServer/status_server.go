package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

//
var (
	serverMap map[string]Server
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
	//添加一个客户端，仅用作测试
	serverMap = make(map[string]Server)
	testServer := Server{
		ServerName:   "test",
		ServerIP:     "127.0.0.1",
		ServerOnline: false,
	}
	testServer.LastActive = time.Now()
	serverMap[testServer.ServerName] = testServer

	//循环监听来自客户端的连接
	//接收到连接之后开启新的线程
	// go handleTcpConnection(conn)
	for {
		recvUDPMessage(conn)
	}
}

//处理UDP数据
//超时未连接的服务器状态要改成离线
//需要维护一个服务器结构体数组
func recvUDPMessage(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	//如果没有数据包传来，线程会在这里阻塞
	n, remoteAddr, err := conn.ReadFromUDP(buf)
	checkError(err, "处理UDP数据包")

	log.Println("接受到UDP数据包:", n, remoteAddr)
	//重新截取切片，不然反序列化的时候会出错
	buf = buf[:n]
	log.Printf("Received data: %s", string(buf[:len(buf)]))
	b := []byte("收到")
	server := Server{}
	if err = json.Unmarshal(buf, &server); err != nil {
		log.Panic(err.Error())
	}
	log.Printf("%+v", server)
	//更新服务器状态，活跃时间等
	if _, ok := serverMap[server.ServerName]; ok {
		editServer := serverMap[server.ServerName]
		editServer.LastActive = time.Now()
		serverMap[server.ServerName] = editServer
		log.Println("更新服务器", serverMap[server.ServerName].LastActive.Unix())
	}

	conn.WriteToUDP(b, remoteAddr)
	//开启一个线程处理
}

//检查服务器状态，定期检查服务器状态，如果太久没接收到信息，服务器状态改为离线。
//数据包5s一个，如果上次活跃时间是10s之前判定为离线
func checkServers() {
	for {
		log.Println("检测服务器在线情况")
		for key, server := range serverMap {
			if time.Now().Unix()-server.LastActive.Unix() > 1000 {
				log.Println("检测到服务器离线", key)
			}
		}
		//10s检测一次
		time.Sleep(10 * time.Second)
	}
}
