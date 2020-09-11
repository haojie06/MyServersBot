package main

import (
	"encoding/json"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

func AddServer(bot *tb.Bot, db *leveldb.DB, msg *tb.Message, idConversationMap map[int]*Conversation) {
	checkIfMapNil(msg.Sender.ID, idConversationMap)
	m, _ := bot.Send(msg.Sender, "准备开始添加服务器，请在聊天栏中依次输入服务器的名字，描述，地点")
	idConversationMap[msg.Sender.ID].HistoryMsg["askAddServerMsg"] = m
	idConversationMap[msg.Sender.ID].CurConversation = "addServer"
	idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器名称"
	ShowAddServerForm(bot, db, msg.Sender, idConversationMap)
	bot.Delete(msg)
}

func HandleAddServer(bot *tb.Bot, db *leveldb.DB, msg *tb.Message, idConversationMap map[int]*Conversation) {
	checkIfMapNil(msg.Sender.ID, idConversationMap)
	switch idConversationMap[msg.Sender.ID].CurAddServerStep {
	case "设置服务器名称":
		{
			bot.Delete(idConversationMap[msg.Sender.ID].HistoryMsg["askAddServerMsg"])
			idConversationMap[msg.Sender.ID].AddServer.ServerName = msg.Text
			idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器简介"
			ShowAddServerForm(bot, db, msg.Sender, idConversationMap)
			bot.Delete(msg)
		}
	case "设置服务器简介":
		{
			idConversationMap[msg.Sender.ID].AddServer.ServerDescription = msg.Text
			idConversationMap[msg.Sender.ID].CurAddServerStep = "设置服务器地点"
			ShowAddServerForm(bot, db, msg.Sender, idConversationMap)
			bot.Delete(msg)
		}
	case "设置服务器地点":
		{
			idConversationMap[msg.Sender.ID].AddServer.ServerLocation = msg.Text
			idConversationMap[msg.Sender.ID].CurAddServerStep = "完成填写"
			ShowAddServerForm(bot, db, msg.Sender, idConversationMap)
			bot.Delete(msg)
		}
	}
}

//展示添加服务器表单的方法，展示当前表单状态，同时提供内联按钮切换下一句输入的内容匹配的表单项
//如果原来有了信息的话，改为编辑
func ShowAddServerForm(bot *tb.Bot, db *leveldb.DB, user *tb.User, idConversationMap map[int]*Conversation) {
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
	bot.Handle(&btnSetName, func(c *tb.Callback) {
		idConversationMap[user.ID].CurAddServerStep = "设置服务器名称"
		idConversationMap[user.ID].AddServer.ServerName = ""
		ShowAddServerForm(bot, db, user, idConversationMap)
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请重新输入服务器名称",
		})
	})

	bot.Handle(&btnSetDescription, func(c *tb.Callback) {
		idConversationMap[user.ID].CurAddServerStep = "设置服务器简介"
		idConversationMap[user.ID].AddServer.ServerDescription = ""
		ShowAddServerForm(bot, db, user, idConversationMap)
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请重新输入服务器简介",
		})
	})
	//确定提交表单
	bot.Handle(&btnConfirm, func(c *tb.Callback) {
		//先获取数据库中的服务器列表，查看是否重名
		mServers, err := db.Get([]byte("servers"), nil)
		if err != nil {
			log.Panic(err.Error())
		}
		var serverMap map[string]Server
		err = json.Unmarshal(mServers, &serverMap)
		if err != nil {
			log.Panic(err.Error())
		}
		//!检查表单是否都填完了
		serverName := idConversationMap[c.Sender.ID].AddServer.ServerName
		if _, exist := serverMap[serverName]; !exist {
			//server
			log.Println("确认添加服务器" + serverName)
			var server Server
			server.ServerName = serverName
			server.ServerDescription = idConversationMap[c.Sender.ID].AddServer.ServerDescription
			server.ServerLocation = idConversationMap[c.Sender.ID].AddServer.ServerLocation
			serverMap[serverName] = server
			mServers, _ = json.Marshal(serverMap)
			db.Put([]byte("servers"), mServers, nil)
			bot.Respond(c, &tb.CallbackResponse{
				Text: "已成功添加",
			})
			bot.Delete(idConversationMap[c.Sender.ID].HistoryMsg["showFormMsg"])
		} else {
			log.Println("添加服务器失败，已有重名服务器" + serverName)
			bot.Respond(c, &tb.CallbackResponse{
				Text: "添加服务器失败，已有重名服务器",
			})
			bot.Delete(idConversationMap[c.Sender.ID].HistoryMsg["showFormMsg"])
			idConversationMap[c.Sender.ID].AddServer = &AddServerForm{}
		}
	})

	bot.Handle(&btnCancel, func(c *tb.Callback) {
		bot.Delete(idConversationMap[c.Sender.ID].HistoryMsg["showFormMsg"])
		idConversationMap[c.Sender.ID].AddServer = &AddServerForm{}
		bot.Respond(c, &tb.CallbackResponse{
			Text: "取消添加",
		})
	})

	bot.Handle(&btnSetLocation, func(c *tb.Callback) {
		idConversationMap[user.ID].CurAddServerStep = "设置服务器地点"
		idConversationMap[user.ID].AddServer.ServerLocation = ""
		ShowAddServerForm(bot, db, user, idConversationMap)
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请重新输入服务器地点",
		})
	})

}
