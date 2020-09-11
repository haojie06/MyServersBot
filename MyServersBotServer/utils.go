package main

import (
	"log"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func checkError(err error, info ...string) {
	if err != nil && len(info) > 0 {
		log.Panic(info[0] + err.Error())
	} else if err != nil {
		log.Panic(err.Error())
	}
}

func sendAutoDeleteMsg(bot *tb.Bot, recv *tb.User, msg string, dur time.Duration) {
	go func() {
		m, _ := bot.Send(recv, msg)
		time.Sleep(dur)
		bot.Delete(m)
	}()
}

//发送信息，并记录到配置文件中
func sendMessage(msg *tb.Message, err error, historyMsgMap map[string]*tb.Message, key string) {
	checkError(err, "发送信息")
	//没错的话记录
	historyMsgMap[key] = msg
}

func deleteHistoryMsg(bot *tb.Bot, historyMsg map[string]*tb.Message, key string) {
	if _, exist := historyMsg[key]; exist {
		bot.Delete(historyMsg[key])
		delete(historyMsg, key)
	}
}

//检查map中是否已经有了用户（没有的话新建一个，不然会出错
func checkIfMapNil(id int, conversationMap map[int]*Conversation) {
	if _, exist := conversationMap[id]; !exist {
		//当前用户id还不存在map中，新建一个
		var c Conversation

		c.Permission = "visitor"
		//主要是这，要新建一个map的map 以及一个表单 ！结构体或者map这类数据都要进行初始化分配空间
		c.HistoryMsg = make(map[string]*tb.Message)
		c.AddServer = &AddServerForm{}
		conversationMap[id] = &c
	}
}
