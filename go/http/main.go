package http

import (
	"fmt"
	"log"
	"qbittorrentRcloneSync/util"
)

const host = "http://127.0.0.1:8080"

func Login() {
	url := host + "/api/v2/auth/login"
	h := make(map[string]string)
	h["Referer"] = host
	p := make(map[string]string)
	p["username"] = "admin"
	p["password"] = "adminadmin"
	res, _ := util.PostForm(url, h, p)
	if res == "Fails." {
		log.Fatal("登录失败") 
	} 
	fmt.Println("登录成功")
}

func GetInfo() ([]map[string]interface{}) {
	url := host + "/api/v2/torrents/info"
	h := make(map[string]string)
	h["Referer"] = host
	p := make(map[string]interface{})
	res, _ := util.Get(url, h, p)
	list := util.ParseJsonStr(res)
	return list
}

func GetDetail(hash string) ([]map[string]interface{}) {
	url := host + "/api/v2/torrents/files"
	h := make(map[string]string)
	h["Referer"] = host
	p := make(map[string]interface{})
	p["hash"] = hash
	res, _ := util.Get(url, h, p)
	list := util.ParseJsonStr(res)
	return list
}

func Resume(hash string) {
	url := host + "/api/v2/torrents/resume"
	h := make(map[string]string)
	h["Referer"] = host
	p := make(map[string]string)
	p["hashes"] = hash
	_, err := util.PostForm(url, h, p)
	if err != nil {
		log.Fatal("恢复下载失败") 
	} 
}

func Pause(hash string) {
	url := host + "/api/v2/torrents/pause"
	h := make(map[string]string)
	h["Referer"] = host
	p := make(map[string]string)
	p["hashes"] = hash
	_, err := util.PostForm(url, h, p)
	if err != nil {
		log.Fatal("暂停下载失败") 
	} 
}