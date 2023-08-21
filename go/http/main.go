package http

import (
	"log"
	"os"
	"qbittorrentRcloneSync/util"
)

var (
	host string
	user string
	password string
)

func getHost() string {
	if host == "" {
		host = os.Getenv("QBIT_URL")
	}
    return host
}

func getUser() string {
	if user == "" {
		user = os.Getenv("QBIT_USER")
	}
    return user
}

func getPasswd() string {
	if password == "" {
		password = os.Getenv("QBIT_PASSWD")
	}
    return password
}

func Login() {
	url := getHost() + "/api/v2/auth/login"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]string)
	p["username"] = getUser()
	p["password"] = getPasswd()
	res, _ := util.PostForm(url, h, p)
	if res == "Fails." {
		log.Fatal("登录失败") 
	}
}

func GetInfo() ([]map[string]interface{}) {
	url := getHost() + "/api/v2/torrents/info"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]interface{})
	res, _ := util.Get(url, h, p)
	list := util.ParseJsonStr(res)
	return list
}

func GetDetail(hash string) ([]map[string]interface{}) {
	url := getHost() + "/api/v2/torrents/files"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]interface{})
	p["hash"] = hash
	res, _ := util.Get(url, h, p)
	list := util.ParseJsonStr(res)
	return list
}

func Resume(hash string) {
	url := getHost() + "/api/v2/torrents/resume"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]string)
	p["hashes"] = hash
	_, err := util.PostForm(url, h, p)
	if err != nil {
		log.Fatal("恢复下载失败") 
	} 
}

func Pause(hash string) {
	url := getHost() + "/api/v2/torrents/pause"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]string)
	p["hashes"] = hash
	_, err := util.PostForm(url, h, p)
	if err != nil {
		log.Fatal("暂停下载失败") 
	} 
}

func ToggleSequentialDownload(hash string) {
	url := getHost() + "/api/v2/torrents/toggleSequentialDownload"
	h := make(map[string]string)
	h["Referer"] = getHost()
	p := make(map[string]string)
	p["hashes"] = hash
	_, err := util.PostForm(url, h, p)
	if err != nil {
		log.Fatal("切换下载类型失败") 
	} 
}
