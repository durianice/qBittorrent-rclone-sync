package main

import (
	"fmt"
	"qbittorrentRcloneSync/http"
	"qbittorrentRcloneSync/util"
	"strings"
)


const (
	RCLONE_NAME = "gdrive:"
	RCLONE_LOCAL_DIR = "/opt/GoogleDrive"
	RCLONE_REMOTE_DIR = "/media/tv/"

	MULTI_THREAD_STREAMS = "10"

	TEMP_PATH = "/opt/qbittorrent/temp_dir"
	LOG_FILE = TEMP_PATH + "/qbit_sync_rclone.log"

	TAG_1 = "脚本控制"
	TAG_2 = "启停控制"
	TAG_3 = "保种"

	MAX_DIST = 10
	MIN_DIST = 3
	THREAD = 2
)

func syncTask(sourceFile string, targetFile string, keepSourceFile bool) {
	option := "moveto"
	if keepSourceFile {
		option = "copyto"
	}
	command := fmt.Sprintf("/usr/bin/rclone -v -P %q --multi-thread-streams %s --log-file %q %q %q", option,
	MULTI_THREAD_STREAMS, LOG_FILE, sourceFile, targetFile)

	util.SendByTelegramBot(fmt.Sprintf("开始同步\n%v ", sourceFile))
	_, err := util.RunShellCommand(command)
	if err != nil {
		util.SendByTelegramBot(fmt.Sprintf("同步出错\n%v ", err))
		return
	}
	util.SendByTelegramBot(fmt.Sprintf("同步完成\n%v ", sourceFile))
}

func main() {
	// util.Env()
	// util.CreateDirIfNotExist(TEMP_PATH)
	http.Login()
	list := http.GetInfo()

	inCtrlList := util.Filter(list, func(obj map[string]interface{}) bool {
		return strings.Contains(obj["tags"].(string), TAG_1)
	})

	for _, obj := range inCtrlList {
		name, _ := obj["name"].(string)
		hash, _ := obj["hash"].(string)
		tags, _ := obj["tags"].(string)
		// 临时下载目录
		downloadPath, _ := obj["download_path"].(string)
		// 下载完成保存目录
		savePath, _ := obj["save_path"].(string)
		// 资源路径 文件：完整路径 目录：父级目录
		// contentPath, _ := obj["content_path"].(string)

		
		subList := http.GetDetail(hash)

		subListDownloaded := util.Filter(subList, func(obj map[string]interface{}) bool {
			return obj["progress"].(float64) == 1
		})

		total := len(subList)
		downloadedLen := len(subListDownloaded)

		util.SendByTelegramBot(fmt.Sprintf("名称：%v \n资源总数：%v 已下载：%v \n标签：%v", name, total, downloadedLen, tags))
		// if downloadedLen == total && !strings.Contains(tags, TAG_3) {
		// 	util.SendByTelegramBot("已全部下载且不需要保种，跳过")
		// 	continue
		// }

		for _, subObj := range subListDownloaded {
			name, _ := subObj["name"].(string)

			sourcePath := downloadPath + "/" + name
			if !util.FileExists(sourcePath) {
				sourcePath = savePath + "/" + name
				if !util.FileExists(sourcePath) {
					continue
				}
			}
			targetPath := RCLONE_NAME + RCLONE_REMOTE_DIR + name
			syncTask(sourcePath, targetPath, strings.Contains(tags, TAG_3))
		}
	}
}