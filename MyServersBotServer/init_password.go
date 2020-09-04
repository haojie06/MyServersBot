package main

import (
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

func InitialPassword(bot *tb.Bot, db *leveldb.DB, v *viper.Viper, initPassword *bool) {

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(m *tb.Message) {
		if *initPassword {
			//构建inline keyboards
			var (
				inlineMenu = &tb.ReplyMarkup{}
				btnConfirm = inlineMenu.Data("确认", "confirmSetPassword", "")
				btnCancel  = inlineMenu.Data("取消", "cancelSetPassword", "")
			)
			inlineMenu.Inline(
				inlineMenu.Row(btnConfirm, btnCancel),
			)
			promptMsg, _ := bot.Send(m.Sender, "确认设置初始密码为"+m.Text+"吗", inlineMenu)
			bot.Handle(&btnCancel, func(c *tb.Callback) {
				bot.Send(c.Sender, "请重新输入初始密码")
				bot.Respond(c, &tb.CallbackResponse{})
			})
			bot.Handle(&btnConfirm, func(c *tb.Callback) {
				//设置管理密码
				v.Set("adminPassword", m.Text)
				v.WriteConfig()
				*initPassword = false
				bot.Delete(promptMsg)
				bot.Respond(c, &tb.CallbackResponse{
					Text: "成功设置管理密码，之后也可以在配置文件bot.yaml中修改",
				})
			})
		}
	})
}
