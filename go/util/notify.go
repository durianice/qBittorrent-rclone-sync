package util

import "fmt"

const (
	CHAT_ID = "111"
	BOT_TOKEN = "111:aaa"
)


func SendByTelegramBot(msg string) {
	fmt.Printf("%v\n", msg)
	url := "https://api.telegram.org/bot" + BOT_TOKEN + "/sendMessage"
	h := make(map[string]string)
	p := make(map[string]interface{})
	p["chat_id"] = CHAT_ID
	p["text"] = msg
	Post(url, h, p)
}