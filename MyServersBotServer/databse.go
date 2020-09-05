package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
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
	//初始化admin
	exist, err = db.Has([]byte("admin"), nil)
	checkError(err, "Key检查")
	if !exist {
		admin := make(map[int]bool, 1)
		mAdmin, err := json.Marshal(admin)
		checkError(err, "序列化")
		db.Put([]byte("admin"), mAdmin, nil)
	}
	exist, err = db.Has([]byte("subscriber"), nil)
	checkError(err, "Key检查")
	//初始化订阅者
	if !exist {
		//考虑到splice没有in方法判断元素是否在其中，这里直接使用map了
		subscriber := make(map[int]*tb.User, 1)
		mSubscriber, err := json.Marshal(subscriber)
		checkError(err, "序列化")
		db.Put([]byte("subscriber"), mSubscriber, nil)
	}

}

//添加订阅
func addSubscriber(db *leveldb.DB, user *tb.User) {
	mSubscriber, err := db.Get([]byte("subscriber"), nil)
	if err != nil {
		log.Panic(err.Error())
	}
	var subscriberMap map[int]*tb.User
	err = json.Unmarshal(mSubscriber, &subscriberMap)
	if err != nil {
		log.Panic("反序列化错误", err.Error())
	}
	subscriberMap[user.ID] = user
	mSubscriber, err = json.Marshal(subscriberMap)
	checkError(err)
	err = db.Put([]byte("subscriber"), mSubscriber, nil)
	if err == nil {
		log.Println("添加订阅成功，用户:", user.Username)
	} else {
		log.Panic(err.Error())
	}
}

//添加管理
func addAdmin(db *leveldb.DB, id int) {
	mAdmin, err := db.Get([]byte("admin"), nil)
	if err != nil {
		log.Panic(err.Error())
	}
	var adminMap map[int]bool
	err = json.Unmarshal(mAdmin, &adminMap)
	if err != nil {
		log.Panic("反序列化错误", err.Error())
	}
	if _, exist := adminMap[id]; !exist {
		log.Print("管理员还未添加，添加管理员")
		adminMap[id] = true
		mAdmin, _ := json.Marshal(adminMap)
		err := db.Put([]byte("admin"), mAdmin, nil)
		if err == nil {
			log.Println("添加管理员成功")
		} else {
			log.Panic(err.Error())
		}
	} else {
		log.Print("管理员已经存在，不重复添加了")
	}
}

func closeDB(db *leveldb.DB) {
	err := db.Close()
	checkError(err)
}
