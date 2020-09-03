package main

//机器人命令处理
import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func registerCommandHandler(bot *tb.Bot) {
	bot.Handle("/start", func(m *tb.Message) {
		bot.Send(m.Sender, "你好，我是服务器探针机器人，请修改好我的配置文件并建立连接吧！")
	})
	bot.Handle("/hello", func(m *tb.Message) {
		bot.Send(m.Sender, "Hello World")
	})
}
