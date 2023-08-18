package util

import (
	"fmt"
	"os"
)


func SendByTelegramBot(msg string) {
	fmt.Printf("%v\n", msg)
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage"
	h := make(map[string]string)
	p := make(map[string]interface{})
	p["chat_id"] = os.Getenv("CHAT_ID")
	p["text"] = msg
	Post(url, h, p)
}