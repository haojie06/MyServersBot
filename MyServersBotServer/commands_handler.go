package main

//机器人命令处理
import (
	"log"
	"time"

	"github.com/spf13/viper"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

//当前机器人没有考虑到同时有多个人对话（回调方法中的变量应该是独立的，但是如果使用了全局变量就会错乱），操作变量的情况，但是对于表单，需要建立一个id——结构体的键值对
//为了及时删除信息，考虑添加一些全局变量来记录发送的信息，或者直接创建一个数组，数组也要放到MAP中，和id对应，避免混乱
func registerCommandHandler(bot *tb.Bot, db *leveldb.DB, v *viper.Viper) {
	//id——会话map
	idConversionMap := make(map[int]*Conversion)
	// var (
	// 	//是否需要初始化管理密码，连接密码
	// 	initAdminPasswd   = false
	// 	initConnectPasswd = false
	// )
	//start之后储存用户信息，订阅，第一个start的用户要求设置管理密码，并把其设置为管理员
	bot.Handle("/start", func(m *tb.Message) {
		//保存到map中
		var c Conversion
		c.CurConversion = ""
		c.Permission = "visitor"
		c.HistoryMsg = make(map[string]*tb.Message)
		idConversionMap[m.Sender.ID] = &c

		// sendAutoDeleteMsg(bot, m.Sender, "你好，我是服务器状态监控机器人", 30*time.Second)
		m, err := bot.Send(m.Sender, "你好，我是服务器状态监控机器人")
		checkError(err)
		//这一行有问题
		log.Println("map中的指针", idConversionMap[m.Sender.ID].HistoryMsg["helloMsg"])
		idConversionMap[m.Sender.ID].HistoryMsg["helloMsg"] = m
		//先看看数据库中有管理密码不，没有的话让第一个用户新建
		// password, err := db.Get([]byte("password"), nil)
		adminPassword := v.GetString("adminPassword")
		connectPassword := v.GetString("connectPassword")
		if adminPassword == "" || connectPassword == "" {
			//要求用户输入密码

			if adminPassword == "" {
				idConversionMap[m.Sender.ID].CurConversion = "askAdminPassword"
				m, _ = bot.Send(m.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)")
				idConversionMap[m.Sender.ID].HistoryMsg["promptMsg"] = m
				m, _ = bot.Send(m.Sender, "请先输入管理密码")
				idConversionMap[m.Sender.ID].HistoryMsg["askAdminPasswordMsg"] = m
			} else {
				idConversionMap[m.Sender.ID].CurConversion = "askConnectPassword"
				m, _ = bot.Send(m.Sender, "管理密码已经设置，请输入客户端连接密码")
				idConversionMap[m.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
			}

			// sendAutoDeleteMsg(bot, m.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)", 15*time.Second)
		}
	})

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(m *tb.Message) {
		switch idConversionMap[m.Sender.ID].CurConversion {

		case "askAdminPassword":
			{
				//构建inline keyboards
				var (
					inlineMenu = &tb.ReplyMarkup{}
					btnConfirm = inlineMenu.Data("确认", "confirmSetPassword", "")
					btnCancel  = inlineMenu.Data("取消", "cancelSetPassword", "")
				)
				inlineMenu.Inline(
					inlineMenu.Row(btnConfirm, btnCancel),
				)
				m, _ := bot.Send(m.Sender, "确认设置初始密码为"+m.Text+"吗", inlineMenu)
				idConversionMap[m.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m
				bot.Handle(&btnCancel, func(c *tb.Callback) {
					sendAutoDeleteMsg(bot, c.Sender, "请重新输入初始密码", 10*time.Second)
					bot.Respond(c, &tb.CallbackResponse{})
				})
				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置管理密码
					v.Set("adminPassword", m.Text)
					v.WriteConfig()
					//确定添加后，删除之前的提示信息
					// bot.Delete(idConversionMap[m.Sender.ID].HistoryMsg["askConfirmPassword"])
					// delete(idConversionMap[m.Sender.ID].HistoryMsg, "askConfirmPassword")
					deleteHistoryMsg(bot, idConversionMap[m.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[m.Sender.ID].HistoryMsg, "askAdminPasswordMsg")
					sendAutoDeleteMsg(bot, c.Sender, "成功设置管理密码，之后也可以在配置文件bot.yaml中修改", 10*time.Second)
					m, _ = bot.Send(m.Sender, "接着，请输入客户端连接密码")
					idConversionMap[m.Sender.ID].CurConversion = "askConnectPassword"
					idConversionMap[m.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
					bot.Respond(c, &tb.CallbackResponse{})
				})
			}
		case "askConnectPassword":
			{
				//询问设置服务器连接密码
				var (
					inlineMenu = &tb.ReplyMarkup{}
					btnConfirm = inlineMenu.Data("确认", "confirmSetConnectPassword", "")
					btnCancel  = inlineMenu.Data("取消", "cancelSetConnectPassword", "")
				)
				inlineMenu.Inline(
					inlineMenu.Row(btnConfirm, btnCancel),
				)
				m, _ := bot.Send(m.Sender, "确认设置连接密码为"+m.Text+"吗", inlineMenu)
				idConversionMap[m.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m

				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置连接
					v.Set("connectPassword", m.Text)
					v.WriteConfig()
					deleteHistoryMsg(bot, idConversionMap[m.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[m.Sender.ID].HistoryMsg, "askConnectPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[m.Sender.ID].HistoryMsg, "promptMsg")
					sendAutoDeleteMsg(bot, c.Sender, "成功设置连接密码，接着添加服务器吧", 10*time.Second)

					bot.Respond(c, &tb.CallbackResponse{})
				})
			}
		}
	})
}
