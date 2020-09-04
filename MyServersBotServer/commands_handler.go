package main

//机器人命令处理
import (
	"log"
	"time"

	"github.com/spf13/viper"

	"github.com/syndtr/goleveldb/leveldb"
	tb "gopkg.in/tucnak/telebot.v2"
)

func registerCommandHandler(bot *tb.Bot, db *leveldb.DB, v *viper.Viper) {
	var (
		//是否需要初始化密码
		initPasswd = false
	)
	bot.Handle("/start", func(m *tb.Message) {
		log.Printf("接收到start请求:\n%+v\n", m.Sender)
		bot.Send(m.Sender, "你好，我是服务器探针机器人")
		//先看看数据库中有管理密码不，没有的话让第一个用户新建
		// password, err := db.Get([]byte("password"), nil)
		password := v.GetString("adminPassword")
		if password == "" {
			//要求用户输入密码
			bot.Send(m.Sender, "这是机器人首次运行，请设置管理密码")
			initPasswd = true
		}
		//先看看用户都有啥信息
		//start之后储存用户信息，订阅，之后的消息推送，第一个start的人作为管理？还是说第一个start的人来设置管理密码？
	})

	//处理用户输入的任何文本
	bot.Handle(tb.OnText, func(m *tb.Message) {
		if initPasswd {
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
				initPasswd = false
				bot.Delete(promptMsg)
				sendAutoDeleteMsg(bot, c.Sender, "成功设置管理密码，之后也可以在配置文件bot.yaml中修改", 10*time.Second)
				bot.Respond(c, &tb.CallbackResponse{})
			})
		}
	})
}
