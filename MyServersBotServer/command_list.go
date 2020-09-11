//command handler里面的命代码太长了，新的命令独立放到其他的文件夹中
package main

import (
	"encoding/json"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

//列出数据库中已经添加的服务器
func ListServers(bot *tb.Bot, db *leveldb.DB, user *tb.User) {
	mServers, err := db.Get([]byte("servers"), nil)
	if err != nil {
		log.Panic(err.Error())
	}
	var serverMap map[string]Server
	if err := json.Unmarshal(mServers, &serverMap); err != nil {
		log.Panic(err.Error())
	}
	showMsg := "当前已添加的服务器列表:\n"
	for _, server := range serverMap {
		showMsg = showMsg + "服务器名称:" + server.ServerName + "\n"
		showMsg = showMsg + "服务器简介:" + server.ServerDescription + "\n"
		showMsg = showMsg + "服务器地点:" + server.ServerLocation + "\n"
		showMsg = showMsg + "-----------\n"
	}
	showMsg += "~~~~~~~~~~~~~~~~~~"
	bot.Send(user, showMsg)
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
	bot.Handle(&btnSetName, func(c *tb.Callback) {
		idConversationMap[user.ID].CurAddServerStep = "设置服务器名称"
		idConversationMap[user.ID].AddServer.ServerName = ""
		showAddServerForm(bot, db, user, idConversationMap)
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请重新输入服务器名称",
		})
	})

	bot.Handle(&btnSetDescription, func(c *tb.Callback) {
		idConversationMap[user.ID].CurAddServerStep = "设置服务器简介"
		idConversationMap[user.ID].AddServer.ServerDescription = ""
		showAddServerForm(bot, db, user, idConversationMap)
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
			db.Put([]byte(serverName), mServers, nil)
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
		showAddServerForm(bot, db, user, idConversationMap)
		bot.Respond(c, &tb.CallbackResponse{
			Text: "请重新输入服务器地点",
		})
	})

}
