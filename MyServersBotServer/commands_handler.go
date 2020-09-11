package main

//机器人命令处理
import (
	"time"

	"github.com/spf13/viper"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

//当前机器人没有考虑到同时有多个人对话（回调方法中的变量应该是独立的，但是如果使用了全局变量就会错乱），操作变量的情况，但是对于表单，需要建立一个id——结构体的键值对
//为了及时删除信息，考虑添加一些全局变量来记录发送的信息，或者直接创建一个数组，数组也要放到MAP中，和id对应，避免混乱
func registerCommandHandler(bot *tb.Bot, db *leveldb.DB, v *viper.Viper) {
	//id——会话map
	idConversationMap := make(map[int]*Conversation)
	// var (
	// 	//是否需要初始化管理密码，连接密码
	// 	initAdminPasswd   = false
	// 	initConnectPasswd = false
	// )
	//start之后储存用户信息，订阅，第一个start的用户要求设置管理密码，并把其设置为管理员
	bot.Handle("/start", func(msg *tb.Message) {
		Start(bot, msg, idConversationMap, v)
	})
	//处理订阅命令
	bot.Handle("/subscribe", func(msg *tb.Message) {
		bot.Delete(msg)
		editSubscriber(db, msg.Sender, true)
		sendAutoDeleteMsg(bot, msg.Sender, "你已经成功订阅通知", 5*time.Second)
	})
	bot.Handle("/unsubscribe", func(msg *tb.Message) {
		bot.Delete(msg)
		editSubscriber(db, msg.Sender, false)
		sendAutoDeleteMsg(bot, msg.Sender, "你已经成功退订通知", 5*time.Second)
	})
	//添加监听的服务器，简化操作直接输入 add 服务器名字即可（还是以多个步骤添加好了，服务器名字/服务器备注/服务器地区）
	//考虑到机器人是让用户自己部署的，所以实际添加的服务器列表不要和对话中的用户绑定，一个机器人只维护一个服务器列表即可
	bot.Handle("/add", func(msg *tb.Message) {
		AddServer(bot, db, msg, idConversationMap)
	})

	bot.Handle("/list", func(msg *tb.Message) {
		ListServers(bot, db, msg.Sender)
	})

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(msg *tb.Message) {
		checkIfMapNil(msg.Sender.ID, idConversationMap)
		idConversationMap[msg.Sender.ID].HistoryMsg["userSendMsg"] = msg
		switch idConversationMap[msg.Sender.ID].CurConversation {
		case "askAdminPassword":
			{
				AskAdminPassword(bot, msg, db, idConversationMap, v)
			}
		case "askConnectPassword":
			{
				AskConnectPassword(bot, msg, db, idConversationMap, v)
			}
		case "addServer":
			{
				//设置服务器表单，三个阶段
				HandleAddServer(bot, db, msg, idConversationMap)
			}
		}
	})
}
