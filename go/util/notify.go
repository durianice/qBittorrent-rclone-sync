package util

import (
	"fmt"
	"os"
	"time"
)

type notifyMap map[string]interface{}
var notify notifyMap

func init() {
	notify = make(notifyMap)
}

func Notify(msg string, _type string) {
	
	// fmt.Printf("%v\n", msg)
	msg = "[" + time.Now().Format("2006-01-02 15:04:05") + "]\n\n" + msg
	if notify[_type] != nil && notify[_type] != "" {
		editTgBotMessage(msg, notify[_type])
		return
	}
	res := sendTgBotMessage(msg)
	parser := JSONParser{}
	err := parser.Parse(res)
	if err != nil {
		fmt.Println("解析 TG MSG JSON 失败：", err)
		return
	}
	message_id, err := parser.Get("result", "message_id")
	if err != nil {
		fmt.Println("获取message_id失败：", err)
	} else {
		// fmt.Println("message_id:", message_id)
		notify[_type] = message_id
	}
}

func DeleteMsg(_type string) {
	if notify[_type] != nil && notify[_type] != "" {
		deleteTgBotMessage(notify[_type])
	}
}

func sendTgBotMessage(msg string) string {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage"
	h := make(map[string]string)
	p := make(map[string]interface{})
	p["chat_id"] = os.Getenv("CHAT_ID")
	p["text"] = msg
	res, err := Post(url, h, p)
	if err != nil {
		return ""
	} 
	return res
}

func editTgBotMessage(msg string, id interface{}) bool {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/editMessageText"
	h := make(map[string]string)
	p := make(map[string]interface{})
	p["chat_id"] = os.Getenv("CHAT_ID")
	p["message_id"] = id
	p["text"] = msg
	res, err := Post(url, h, p)
	if err != nil {
		return false
	} 
	parser := JSONParser{}
	parseErr := parser.Parse(res)
	if parseErr != nil {
		fmt.Println("解析 TG MSG JSON 失败：", err)
		return false
	}
	ok, msgErr := parser.Get("ok")
	if msgErr != nil {
		return false
	}
	return ok.(bool)
}

func deleteTgBotMessage(id interface{}) bool {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/deletemessage"
	h := make(map[string]string)
	p := make(map[string]interface{})
	p["chat_id"] = os.Getenv("CHAT_ID")
	p["message_id"] = id
	res, err := Post(url, h, p)
	if err != nil {
		return false
	}
	parser := JSONParser{}
	parseErr := parser.Parse(res)
	if parseErr != nil {
		fmt.Println("解析 TG MSG JSON 失败：", err)
		return false
	}
	ok, msgErr := parser.Get("ok")
	if msgErr != nil {
		return false
	}
	return ok.(bool)
}