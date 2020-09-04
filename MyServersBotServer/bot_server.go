package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	//读取配置文件
	viper.SetConfigFile("./bot.yaml")
	err := viper.ReadInConfig()
	checkError(err, "读取配置文件")

	bot, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		// URL: "",
		Token:  viper.GetString("token"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	//移除之前的webhook
	if err := bot.RemoveWebhook(); err != nil {
		log.Panic(err.Error())
	}
	//开启数据库
	db := startDB()
	initDB(db)
	defer closeDB(db)
	registerCommandHandler(bot, db, viper.GetViper())

	//开启状态监控服务器
	go startStatusServer(db, viper.GetString("listenPort"))
	go checkServers()
	logo := `
 __  __          _____                                     ____          _   
|  \/  |        / ____|                                   |  _ \        | |  
| \  / | _   _ | (___    ___  _ __ __   __ ___  _ __  ___ | |_) |  ___  | |_ 
| |\/| || | | | \___ \  / _ \| '__|\ \ / // _ \| '__|/ __||  _ <  / _ \ | __|
| |  | || |_| | ____) ||  __/| |    \ V /|  __/| |   \__ \| |_) || (_) || |_ 
|_|  |_| \__, ||_____/  \___||_|     \_/  \___||_|   |___/|____/  \___/  \__|
          __/ |                                                              
         |___/                                                               

	`
	fmt.Println(logo)
	log.Println("MyServersBot started...")
	bot.Start()
}
