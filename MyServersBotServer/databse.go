package main

import (
	"encoding/json"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

func startDB() *leveldb.DB {
	db, err := leveldb.OpenFile("db/botdb", nil)
	checkError(err)
	return db
}

func initDB(db *leveldb.DB) {
	//数据库初始化(如放入默认配置，默认服务器列表等)
	exist, err := db.Has([]byte("password"), nil)
	checkError(err, "Key检查")
	if !exist {
		err = db.Put([]byte("password"), []byte(""), nil)
		checkError(err, "写入测试服务器")
	}
	//初始化servers
	exist, err = db.Has([]byte("servers"), nil)
	checkError(err, "Key检查")
	if !exist {
		//添加一个客户端，仅用作测试
		serverMap := make(map[string]Server)
		testServer := Server{
			ServerName:   "test",
			ServerIP:     "127.0.0.1",
			ServerOnline: false,
		}
		testServer.LastActive = time.Now()
		serverMap[testServer.ServerName] = testServer
		// 测试： 先把测试的MAP序列化并存入数据库中
		mServerMap, err := json.Marshal(serverMap)
		checkError(err, "map序列化")
		// 测试：写入测试服务器
		err = db.Put([]byte("servers"), mServerMap, nil)
		checkError(err, "写入测试服务器")
	}
	//初始化settings，settings还是用外部配置文件算了，因为为了方便用户部署，bot token要方便设置
}

func closeDB(db *leveldb.DB) {
	err := db.Close()
	checkError(err)
}
