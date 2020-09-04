package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

//
var (
	serverMap map[string]Server
)

//状态监控使用udp来通信
func startStatusServer(db *leveldb.DB) {
	//为了在使用hostname的情况下也可以正常运行
	addr, err := net.ResolveUDPAddr("udp", "localhost:10086")
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	log.Println("状态监控服务器开启")
	//方法返回（退出后）释放资源，关闭连接
	defer conn.Close()

	//添加一个客户端，仅用作测试
	// serverMap = make(map[string]Server)
	// testServer := Server{
	// 	ServerName:   "test",
	// 	ServerIP:     "127.0.0.1",
	// 	ServerOnline: false,
	// }
	// testServer.LastActive = time.Now()
	// serverMap[testServer.ServerName] = testServer
	//测试： 先把测试的MAP序列化并存入数据库中
	// mServerMap, err := json.Marshal(serverMap)
	// checkError(err)
	//测试：写入测试服务器
	// err = db.Put([]byte("servers"), mServerMap, nil)
	// checkError(err)
	//读取数据库，获取所有的Servers
	dbServersJSON, err := db.Get([]byte("servers"), nil)
	if err == nil {
		//没出错的话，这个键值对才存在
		err = json.Unmarshal(dbServersJSON, &serverMap)
		checkError(err, "反序列化")
	}
	log.Printf("当前记录服务器:%d个", len(serverMap))
	//循环监听来自客户端的连接
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
	//如果这里不开启新的线程，多个数据包同时到达时会发生什么？ ——按顺序处理，如果只是简单操作应该不会有什么差别
	go func() {
		log.Println("接受到UDP数据包:", n, remoteAddr)
		//重新截取切片，不然反序列化的时候会出错
		buf = buf[:n]
		// log.Printf("Received data: %s", string(buf[:len(buf)]))

		server := Server{}
		if err = json.Unmarshal(buf, &server); err != nil {
			log.Panic(err.Error())
		}
		// log.Printf("%+v", server)
		//更新服务器状态，活跃时间等
		//如果服务器上线，发送通知
		if s, ok := serverMap[server.ServerName]; ok {
			if s.ServerOnline == false {
				s.ServerOnline = true
				log.Println("检测到服务器上线", s.ServerName)
			}
			s.LastActive = time.Now()
			serverMap[server.ServerName] = s
			// log.Println("更新服务器", serverMap[server.ServerName].LastActive.Unix())
		}
		//如果需要向客户端发送信息
		// b := []byte("收到")
		// conn.WriteToUDP(b, remoteAddr)

	}()
}

//检查服务器状态，定期检查服务器状态，如果太久没接收到信息，服务器状态改为离线。
//数据包5s一个，如果上次活跃时间是10s之前判定为离线
//如果离线过久发送警报 如离线60s以上。
func checkServers() {
	for {

		for key, server := range serverMap {
			// log.Printf("检测服务器在线情况 服务器:%s 活跃时间:%d 现在时间:%d\n", key, server.LastActive.Unix(), time.Now().Unix())
			if time.Now().Unix()-server.LastActive.Unix() > 10 {
				log.Println("检测到服务器离线", key)
				server.ServerOnline = false
				serverMap[key] = server
			}
		}
		//10s检测一次
		time.Sleep(10 * time.Second)
	}
}
