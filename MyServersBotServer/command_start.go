package main

//机器人命令处理
import (
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

func Start(bot *tb.Bot, msg *tb.Message, idConversationMap map[int]*Conversation, v *viper.Viper) {
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
}
func AskConnectPassword(bot *tb.Bot, msg *tb.Message, db *leveldb.DB, idConversationMap map[int]*Conversation, v *viper.Viper) {
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

func AskAdminPassword(bot *tb.Bot, msg *tb.Message, db *leveldb.DB, idConversationMap map[int]*Conversation, v *viper.Viper) {
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
