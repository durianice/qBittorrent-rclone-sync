package main

import (
	"fmt"
	"qbittorrentRcloneSync/http"
	"qbittorrentRcloneSync/util"
	"strings"
	"sync"
	"time"
)


const (
	RCLONE_NAME = "gdrive:"
	RCLONE_LOCAL_DIR = "/opt/GoogleDrive"
	RCLONE_REMOTE_DIR = "/media/tv/"

	MULTI_THREAD_STREAMS = "10"

	TEMP_PATH = "/opt/qbittorrent/temp_dir"
	LOG_FILE = TEMP_PATH + "/qbit_sync_rclone.log"

	TAG_1 = "脚本控制"
	TAG_2 = "保种"

	THREAD = 10
)

var runTask bool = true

func rcloneTask(sourceFile string, targetFile string, keepSourceFile bool, wg *sync.WaitGroup, ch chan struct{}) error {
	defer wg.Done()
	option := "moveto"
	if keepSourceFile {
		option = "copyto"
	}
	command := fmt.Sprintf("/usr/bin/rclone -v -P %s --multi-thread-streams %s --log-file %s %s %s", option,
	MULTI_THREAD_STREAMS, LOG_FILE, sourceFile, targetFile)
	_, err := util.RunShellCommand(command)
	<-ch
	if err != nil {
		return err
	}
	return nil
	
}

func getList() ([]map[string]interface{}) {
	result := []map[string]interface{}{}
	http.Login()
	list := http.GetInfo()
	inCtrlList := util.Filter(list, func(obj map[string]interface{}) bool {
		return strings.Contains(obj["tags"].(string), TAG_1)
	})
	for _, obj := range inCtrlList {
		name, _ := obj["name"].(string)
		hash, _ := obj["hash"].(string)
		tags, _ := obj["tags"].(string)
		seqDl, _ := obj["seq_dl"].(bool)
		state, _ := obj["state"].(string)
		// 临时下载目录
		downloadPath, _ := obj["download_path"].(string)
		// 下载完成保存目录
		savePath, _ := obj["save_path"].(string)

		subList := http.GetDetail(hash)

		subListDownloaded := util.Filter(subList, func(obj map[string]interface{}) bool {
			return obj["progress"].(float64) == 1
		})
		for _, subObj := range subListDownloaded {
			subName, _ := subObj["name"].(string)
			newObj := map[string]interface{}{
				"name": name,
				"hash": hash,
				"tags": tags,
				"seqDl": seqDl,
				"state": state,
				"downloadPath": downloadPath,
				"savePath": savePath,
				"subName": subName,
			}
			result = append(result, newObj)
		}
	}
	return result
}

func memoryControl() {
	used := util.GetUsedSpacePercentage()
	res, _ := util.PercentageToDecimal(used)
	if res >= 0.90 {
		runTask = false
	}
	if res < 0.45 {
		runTask = true
	}
}

func mainTask() {
	var wg sync.WaitGroup
	ch := make(chan struct{}, THREAD)
	util.CreateDirIfNotExist(TEMP_PATH)
	http.Login()
	list := http.GetInfo()

	inCtrlList := util.Filter(list, func(obj map[string]interface{}) bool {
		return strings.Contains(obj["tags"].(string), TAG_1)
	})


	util.SendByTelegramBot(fmt.Sprintf("总数量：%v 脚本管理数量：%v", len(list), len(inCtrlList)))
	for _, obj := range inCtrlList {
		name, _ := obj["name"].(string)
		hash, _ := obj["hash"].(string)
		tags, _ := obj["tags"].(string)
		seqDl, _ := obj["seq_dl"].(bool)
		state, _ := obj["state"].(string)
		// 临时下载目录
		downloadPath, _ := obj["download_path"].(string)
		// 下载完成保存目录
		savePath, _ := obj["save_path"].(string)
		// 资源路径 文件：完整路径 目录：父级目录
		// contentPath, _ := obj["content_path"].(string)

		if state == "pausedDL" {
			http.Resume(hash)
		}
		if !seqDl {
			http.ToggleSequentialDownload(hash)
		}

		subList := http.GetDetail(hash)

		subListDownloaded := util.Filter(subList, func(obj map[string]interface{}) bool {
			return obj["progress"].(float64) == 1
		})

		// total := len(subList)
		downloadedLen := len(subListDownloaded)

		// util.SendByTelegramBot(fmt.Sprintf("名称：%s \n资源总数：%s 已下载：%s \n标签：%s", name, total, downloadedLen, tags))
		// if downloadedLen == total && !strings.Contains(tags, TAG_2) {
		// 	util.SendByTelegramBot("已全部下载且不需要保种，跳过")
		// 	continue
		// }

		for index, subObj := range subListDownloaded {
			subName, _ := subObj["name"].(string)
			sourcePath := downloadPath + "/" + subName
			targetPath := RCLONE_NAME + RCLONE_REMOTE_DIR + subName
			localTargetPath := RCLONE_LOCAL_DIR + RCLONE_REMOTE_DIR + subName
			if !util.FileExists(sourcePath) {
				sourcePath = savePath + "/" + subName
				if !util.FileExists(sourcePath) {
					continue
				}
			}
			if util.FileExists(localTargetPath) && util.FileExists(sourcePath) && !strings.Contains(tags, TAG_2) {
				command := fmt.Sprintf("sudo rm -r %s", subName)
				util.RunShellCommand(command)
				continue
			}
			ch <- struct{}{}
			wg.Add(1)
			go func(a string, b string, c string, wg *sync.WaitGroup, ch chan struct{}) {
				util.SendByTelegramBot(fmt.Sprintf("名称 %s\n开始同步 (%v/%v)", name, index + 1, downloadedLen))
				err := rcloneTask(a, b, strings.Contains(c, TAG_2), wg, ch)
				if err == nil {
					util.SendByTelegramBot(fmt.Sprintf("名称 %s\n同步完成 (%v/%v) \n已用空间 %s", name, index + 1, downloadedLen, util.GetUsedSpacePercentage()))
				} else {
					util.SendByTelegramBot(fmt.Sprintf("名称 %s\n同步错误 (%v/%v)\n错误原因：%s", name, index + 1, downloadedLen, err))
				}
			}(sourcePath, targetPath, tags, &wg, ch)
			
		}
	}
	wg.Wait()
	close(ch)
}

func main() {
	util.Env()
	// getList()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				memoryControl()
			}
		}
	}()
	for {
		if runTask {
			util.SendByTelegramBot("开始运行")
			util.SendByTelegramBot(fmt.Sprintf("已用空间：%s\n ", util.GetUsedSpacePercentage()))
			mainTask()
			util.SendByTelegramBot("结束运行")
		}
		time.Sleep(60 * time.Second)
	}
}