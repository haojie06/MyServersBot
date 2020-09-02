package main

//机器人命令处理
import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func registerCommandHandler(bot *tb.Bot) {
	bot.Handle("/hello", func(m *tb.Message) {
		bot.Send(m.Sender, "Hello World")
	})
}
