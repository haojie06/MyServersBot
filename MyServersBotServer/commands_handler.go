package main

//机器人命令处理
import (
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
	bot.Handle("/start", func(msg *tb.Message) {
		//保存到map中
		var c Conversion
		c.CurConversion = ""
		c.Permission = "visitor"
		c.HistoryMsg = make(map[string]*tb.Message)
		idConversionMap[msg.Sender.ID] = &c
		m, err := bot.Send(msg.Sender, "你好，我是服务器状态监控机器人")
		checkError(err)
		idConversionMap[msg.Sender.ID].HistoryMsg["helloMsg"] = m
		//如果配置文件中没有管理密码没有的话让第一个用户新建
		// password, err := db.Get([]byte("password"), nil)
		adminPassword := v.GetString("adminPassword")
		connectPassword := v.GetString("connectPassword")
		if adminPassword == "" || connectPassword == "" {
			//要求用户输入密码
			if adminPassword == "" {
				idConversionMap[msg.Sender.ID].CurConversion = "askAdminPassword"
				m, _ = bot.Send(msg.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)")
				idConversionMap[msg.Sender.ID].HistoryMsg["startMsg"] = m
				m, _ = bot.Send(msg.Sender, "请先输入管理密码")
				idConversionMap[msg.Sender.ID].HistoryMsg["askAdminPasswordMsg"] = m
			} else {
				idConversionMap[msg.Sender.ID].CurConversion = "askConnectPassword"
				m, _ = bot.Send(msg.Sender, "管理密码已经设置，请输入客户端连接密码")
				idConversionMap[msg.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
			}
			// sendAutoDeleteMsg(bot, msg.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)", 15*time.Second)
		}
	})

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(msg *tb.Message) {
		idConversionMap[msg.Sender.ID].HistoryMsg["userSendMsg"] = msg
		switch idConversionMap[msg.Sender.ID].CurConversion {
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
				m, _ := bot.Send(msg.Sender, "确认设置初始密码为"+msg.Text+"吗", inlineMenu)
				v.Set("adminPassword", msg.Text)
				idConversionMap[msg.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m
				bot.Handle(&btnCancel, func(c *tb.Callback) {
					bot.Respond(c, &tb.CallbackResponse{
						Text: "请重新输入初始密码",
					})
					deleteHistoryMsg(bot, idConversionMap[msg.Sender.ID].HistoryMsg, "userSendMsg")
					bot.Delete(m)
				})

				//确认按钮
				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置管理密码
					v.WriteConfig()
					//确定添加后，删除之前的提示信息
					deleteHistoryMsg(bot, idConversionMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					deleteHistoryMsg(bot, idConversionMap[c.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[c.Sender.ID].HistoryMsg, "askAdminPasswordMsg")
					//把该用户加到管理员中
					addAdmin(db, c.Sender.ID)
					addSubscriber(db, c.Sender)
					m, _ = bot.Send(c.Sender, "接着，请输入客户端连接密码")
					idConversionMap[c.Sender.ID].CurConversion = "askConnectPassword"
					idConversionMap[c.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
					bot.Respond(c, &tb.CallbackResponse{
						Text: "成功设置管理密码，之后也可以在配置文件bot.yaml中修改",
					})
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
				v.Set("connectPassword", msg.Text)
				m, _ := bot.Send(msg.Sender, "确认设置连接密码为"+msg.Text+"吗", inlineMenu)
				idConversionMap[msg.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m

				bot.Handle(&btnCancel, func(c *tb.Callback) {
					bot.Delete(m)
					deleteHistoryMsg(bot, idConversionMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					bot.Respond(c, &tb.CallbackResponse{
						Text: "请重新输入连接密码",
					})
				})
				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置连接
					v.WriteConfig()
					deleteHistoryMsg(bot, idConversionMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					deleteHistoryMsg(bot, idConversionMap[msg.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[msg.Sender.ID].HistoryMsg, "askConnectPasswordMsg")
					deleteHistoryMsg(bot, idConversionMap[msg.Sender.ID].HistoryMsg, "helloMsg")
					bot.Respond(c, &tb.CallbackResponse{
						Text: "成功设置连接密码，接着添加服务器吧。",
					})
				})
			}
		}
	})
}
