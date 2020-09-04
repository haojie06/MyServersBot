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

// bot.Delete(idConversionMap[m.Sender.ID].HistoryMsg["askConfirmPassword"])
// delete(idConversionMap[m.Sender.ID].HistoryMsg, "askConfirmPassword")
func deleteHistoryMsg(bot *tb.Bot, historyMsg map[string]*tb.Message, key string) {
	bot.Delete(historyMsg[key])
	delete(historyMsg, key)
}
