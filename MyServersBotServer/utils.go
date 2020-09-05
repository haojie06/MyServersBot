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
