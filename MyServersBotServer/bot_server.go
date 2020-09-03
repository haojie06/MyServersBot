package main

import (
	"log"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	bot, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		// URL: "",
		Token:  "1343712816:AAFLRDa4DYZ_ryHMEF5vJrlO-gHsHxf68GA",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	registerCommandHandler(bot)
	//开启状态监控服务器
	go startStatusServer()
	log.Println("MyServersBot started...")
	bot.Start()
}
