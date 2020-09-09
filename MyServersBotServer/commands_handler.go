package main

//机器人命令处理
import (
	"github.com/spf13/viper"

	"time"

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
		//保存到map中
		// var c Conversation
		// c.CurConversation = ""
		// c.Permission = "visitor"
		// c.HistoryMsg = make(map[string]*tb.Message)
		checkIfMapNil(msg.Sender.ID, idConversationMap)
		// idConversationMap[msg.Sender.ID] = &c
		m, err := bot.Send(msg.Sender, "你好，我是服务器状态监控机器人")
		checkError(err)
		idConversationMap[msg.Sender.ID].HistoryMsg["helloMsg"] = m
		//如果配置文件中没有管理密码没有的话让第一个用户新建
		// password, err := db.Get([]byte("password"), nil)
		adminPassword := v.GetString("adminPassword")
		connectPassword := v.GetString("connectPassword")
		if adminPassword == "" || connectPassword == "" {
			//要求用户输入密码
			if adminPassword == "" {
				idConversationMap[msg.Sender.ID].CurConversation = "askAdminPassword"
				m, _ = bot.Send(msg.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)")
				idConversationMap[msg.Sender.ID].HistoryMsg["startMsg"] = m
				m, _ = bot.Send(msg.Sender, "请先输入管理密码")
				idConversationMap[msg.Sender.ID].HistoryMsg["askAdminPasswordMsg"] = m
			} else {
				idConversationMap[msg.Sender.ID].CurConversation = "askConnectPassword"
				m, _ = bot.Send(msg.Sender, "管理密码已经设置，请输入客户端连接密码")
				idConversationMap[msg.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
			}
			// sendAutoDeleteMsg(bot, msg.Sender, "这是机器人首次运行，请设置管理密码(用于管理机器人和服务器)，与连接密码(客户端连接服务器时使用)", 15*time.Second)
		}
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
		checkIfMapNil(msg.Sender.ID, idConversationMap)
		m, _ := bot.Send(msg.Sender, "准备开始添加服务器，请在聊天栏中依次输入服务器的名字，描述，地点")
		idConversationMap[msg.Sender.ID].HistoryMsg["askAddServerMsg"] = m
		idConversationMap[msg.Sender.ID].CurConversation = "addServer"
		idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器名称"
		showAddServerForm(bot, db, msg.Sender, idConversationMap)
		bot.Delete(msg)
		// var (
		// 	inlineMenu = &tb.ReplyMarkup{}
		// 	btnConfirm = inlineMenu.Data("确认", "confirmAddServer", "")
		// 	btnCancel  = inlineMenu.Data("取消", "cancelAddServer", "")
		// )
		// inlineMenu.Inline(
		// 	inlineMenu.Row(btnConfirm, btnCancel),
		// )
		// m, _ := bot.Send(msg.Sender, "确定要添加服务器:"+msg.Payload+"吗?", inlineMenu)
		// // idConversationMap[msg.Sender.ID].HistoryMsg["askAddServerMsg"] = m
		// //取消添加服务器
		// bot.Handle(&btnCancel, func(c *tb.Callback) {
		// 	bot.Respond(c, &tb.CallbackResponse{
		// 		Text: "请重新输入服务器名字",
		// 	})
		// 	// deleteHistoryMsg(bot, idConversationMap[msg.Sender.ID].HistoryMsg, "askAddServerMsg")
		// 	bot.Delete(m)
		// })

	})

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(msg *tb.Message) {
		idConversationMap[msg.Sender.ID].HistoryMsg["userSendMsg"] = msg
		switch idConversationMap[msg.Sender.ID].CurConversation {
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
				idConversationMap[msg.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m
				bot.Handle(&btnCancel, func(c *tb.Callback) {
					bot.Respond(c, &tb.CallbackResponse{
						Text: "请重新输入初始密码",
					})
					deleteHistoryMsg(bot, idConversationMap[msg.Sender.ID].HistoryMsg, "userSendMsg")
					bot.Delete(m)
				})

				//确认按钮
				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置管理密码
					v.WriteConfig()
					//确定添加后，删除之前的提示信息
					deleteHistoryMsg(bot, idConversationMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					deleteHistoryMsg(bot, idConversationMap[c.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversationMap[c.Sender.ID].HistoryMsg, "askAdminPasswordMsg")
					//把该用户加到管理员中
					addAdmin(db, c.Sender.ID)
					editSubscriber(db, c.Sender, true)
					m, _ = bot.Send(c.Sender, "接着，请输入客户端连接密码")
					idConversationMap[c.Sender.ID].CurConversation = "askConnectPassword"
					idConversationMap[c.Sender.ID].HistoryMsg["askConnectPasswordMsg"] = m
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
				idConversationMap[msg.Sender.ID].HistoryMsg["askConfirmPasswordMsg"] = m

				bot.Handle(&btnCancel, func(c *tb.Callback) {
					bot.Delete(m)
					deleteHistoryMsg(bot, idConversationMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					bot.Respond(c, &tb.CallbackResponse{
						Text: "请重新输入连接密码",
					})
				})
				bot.Handle(&btnConfirm, func(c *tb.Callback) {
					//设置连接
					v.WriteConfig()
					deleteHistoryMsg(bot, idConversationMap[c.Sender.ID].HistoryMsg, "userSendMsg")
					deleteHistoryMsg(bot, idConversationMap[msg.Sender.ID].HistoryMsg, "askConfirmPasswordMsg")
					deleteHistoryMsg(bot, idConversationMap[msg.Sender.ID].HistoryMsg, "askConnectPasswordMsg")
					deleteHistoryMsg(bot, idConversationMap[msg.Sender.ID].HistoryMsg, "helloMsg")
					bot.Respond(c, &tb.CallbackResponse{
						Text: "成功设置连接密码，接着添加服务器吧。",
					})
				})
			}
		case "addServer":
			{

				//设置服务器表单，三个阶段
				switch idConversationMap[msg.Sender.ID].CurAddServerStep {
				case "设置服务器名称":
					{
						idConversationMap[msg.Sender.ID].AddServer.ServerName = msg.Text
						idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器简介"
						showAddServerForm(bot, db, msg.Sender, idConversationMap)
						bot.Delete(msg)
					}
				case "设置服务器简介":
					{
						idConversationMap[msg.Sender.ID].AddServer.ServerDescription = msg.Text
						idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器地点"
						showAddServerForm(bot, db, msg.Sender, idConversationMap)
						bot.Delete(msg)
					}
				case "设置服务器地点":
					{
						idConversationMap[msg.Sender.ID].AddServer.ServerLocation = msg.Text
						idConversationMap[msg.Sender.ID].CurAddServerStep = "完成填写"
						showAddServerForm(bot, db, msg.Sender, idConversationMap)
						bot.Delete(msg)
					}
				}
			}
		}
	})
}

//展示添加服务器表单的方法，展示当前表单状态，同时提供内联按钮切换下一句输入的内容匹配的表单项
//如果原来有了信息的话，改为编辑
func showAddServerForm(bot *tb.Bot, db *leveldb.DB, user *tb.User, idConversationMap map[int]*Conversation) {
	//展示当前服务器表单
	c := idConversationMap[user.ID]
	showForm := "当前待添加服务器表单:\n" + "服务器名称:" + c.AddServer.ServerName + "\n服务器简介:" + c.AddServer.ServerDescription + "\n服务器地点:" + c.AddServer.ServerLocation + "\n当前:" + c.CurAddServerStep
	//构建内联菜单
	var (
		inlineMenu        = &tb.ReplyMarkup{}
		btnSetName        = inlineMenu.Data("设置名称", "setServerName", "")
		btnSetDescription = inlineMenu.Data("设置简介", "setServerDescription", "")
		btnSetLocation    = inlineMenu.Data("设置地点", "setServerLocation", "")
		btnConfirm        = inlineMenu.Data("确认添加", "confirmAddServer", "")
		btnCancel         = inlineMenu.Data("取消添加", "cancelAddServer", "")
	)
	inlineMenu.Inline(
		inlineMenu.Row(btnSetName, btnSetDescription, btnSetLocation),
		inlineMenu.Row(btnConfirm, btnCancel),
	)

	if fMsg, exist := idConversationMap[user.ID].HistoryMsg["showFormMsg"]; exist {
		m, _ := bot.Edit(fMsg, showForm, inlineMenu)
		idConversationMap[user.ID].HistoryMsg["showFormMsg"] = m
	} else {
		m, _ := bot.Send(user, showForm, inlineMenu)
		idConversationMap[user.ID].HistoryMsg["showFormMsg"] = m
	}

	//按钮监听处理部分
	bot.Handle(&btnCancel, func(c *tb.Callback) {
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请想好之后再操作吧",
		})
		//清空当前用户列表中的表单
		idConversationMap[user.ID].AddServer = &AddServerForm{}
		bot.Delete(idConversationMap[user.ID].HistoryMsg["showFormMsg"])
		delete(idConversationMap[user.ID].HistoryMsg, "showFormMsg")

	})

}

//检查map中是否已经有了用户（没有的话新建一个，不然会出错

func checkIfMapNil(id int, conversationMap map[int]*Conversation) {
	if _, exist := conversationMap[id]; !exist {
		//当前用户id还不存在map中，新建一个
		var c Conversation
		c.AddServer = &AddServerForm{}
		c.Permission = "visitor"
		//主要是这，要新建一个map的map
		c.HistoryMsg = make(map[string]*tb.Message)
		conversationMap[id] = &c
	}
}
