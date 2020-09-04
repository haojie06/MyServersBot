package main

//机器人命令处理
import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

func registerCommandHandler(bot *tb.Bot, db *leveldb.DB) {
	bot.Handle("/start", func(m *tb.Message) {
		log.Printf("接收到start请求:\n%+v\n", m.Sender)
		bot.Send(m.Sender, "你好，我是服务器探针机器人")
		//先看看数据库中有管理密码不，没有的话让第一个用户新建
		password, err := db.Get([]byte("password"), nil)
		checkError(err)
		if string(password) == "" {
			//要求用户输入密码
			bot.Send(m.Sender, "这是机器人首次运行，请设置管理密码")
		}
		//先看看用户都有啥信息
		//start之后储存用户信息，订阅，之后的消息推送，第一个start的人作为管理？还是说第一个start的人来设置管理密码？

	})
	bot.Handle("/hello", func(m *tb.Message) {
		bot.Send(m.Sender, "Hello World")
	})
}
