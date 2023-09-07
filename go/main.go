package main

import (
	"fmt"
	"os"
	"qbittorrentRcloneSync/http"
	"qbittorrentRcloneSync/util"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/joho/godotenv"
)


var (
	RCLONE_NAME string
	RCLONE_LOCAL_DIR  string
	RCLONE_REMOTE_DIR string
	MULTI_THREAD_STREAMS string
	LOG_FILE string
	TAG_1 string
	TAG_2 string
	THREAD string
)

var qBitList []map[string]interface{}

func rcloneTask(sourceFile string, targetFile string, keepSourceFile bool, wg *sync.WaitGroup, ch chan struct{}) error {
	defer wg.Done()
	option := "moveto"
	if keepSourceFile {
		option = "copyto"
	}
	command := fmt.Sprintf("/usr/bin/rclone -v -P %s --multi-thread-streams %s --log-file %q %q %q", option,
	MULTI_THREAD_STREAMS, LOG_FILE, sourceFile, targetFile)
	_, err := util.RunShellCommand(command)
	<-ch
	if err != nil {
		return err
	}
	return nil
	
}

func memoryControl() string {
	used := util.GetUsedSpacePercentage()
	res, _ := util.PercentageToDecimal(used)
	if res >= 0.90 {
		return "P"
	}
	if res < 0.45 {
		return "D"
	}
	return "N"
}

func getList() ([]map[string]interface{}) {
	http.Login()
	list := http.GetInfo()
	// 按标签过滤
	inCtrlList := util.Filter(list, func(obj map[string]interface{}) bool {
		return strings.Contains(obj["tags"].(string), TAG_1)
	})
	return util.Map(inCtrlList, func(obj map[string]interface{}) map[string]interface{} {
		name, _ := obj["name"].(string)
		hash, _ := obj["hash"].(string)
		tags, _ := obj["tags"].(string)
		seqDl, _ := obj["seq_dl"].(bool)
		state, _ := obj["state"].(string)
		downloadPath, _ := obj["download_path"].(string)
		savePath, _ := obj["save_path"].(string)
		// 过滤已下载完成的子内容
		subListDownloaded := util.Filter(http.GetDetail(hash), func(obj map[string]interface{}) bool {
			return obj["progress"].(float64) == 1
		})
		newObj := map[string]interface{}{
			"name": name,
			"hash": hash,
			"tags": tags,
			"seqDl": seqDl,
			"state": state,
			"downloadPath": downloadPath,
			"savePath": savePath,
			"subListDownloaded": subListDownloaded,
		}
		memState := memoryControl()
		if memState == "P" {
			http.Pause(hash)
		}
		if memState == "D" {
			http.Resume(hash)
		}
		if !seqDl {
			http.ToggleSequentialDownload(hash)
		}
		return newObj
	})
}

func mainTask() {
	var wg sync.WaitGroup
	THREAD, err := strconv.Atoi(THREAD)
	if err != nil {
		panic("Error converting THREAD to int")
	}
	ch := make(chan struct{}, THREAD)

	for _, obj := range qBitList {
		name, _ := obj["name"].(string)
		tags, _ := obj["tags"].(string)
		downloadPath, _ := obj["downloadPath"].(string)
		savePath, _ := obj["savePath"].(string)
		subListDownloaded, _ := obj["subListDownloaded"].([]map[string]interface{})
		downloadedLen := len(subListDownloaded)
		for index, subObj := range subListDownloaded {
			subName, _ := subObj["name"].(string)
			sourcePath := downloadPath + "/" + subName
			targetPath := RCLONE_NAME + RCLONE_REMOTE_DIR + subName
			localTargetPath := RCLONE_LOCAL_DIR + RCLONE_REMOTE_DIR + subName
			if !util.FileExists(sourcePath) {
				sourcePath = savePath + "/" + subName
				if !util.FileExists(sourcePath) {
					fmt.Printf("%v\n未找到资源，跳过", sourcePath)
					continue
				}
			}
			if util.FileExists(localTargetPath) {
				if util.FileExists(sourcePath) && !strings.Contains(tags, TAG_2) {
					command := fmt.Sprintf("sudo rm %q", sourcePath)
					util.RunShellCommand(command)
				}
				fmt.Printf("%v\n云盘已有该资源，跳过", sourcePath)
				continue
			}
			ch <- struct{}{}
			wg.Add(1)
			go func(a string, b string, c string, wg *sync.WaitGroup, ch chan struct{}, index int) {
				err := rcloneTask(a, b, strings.Contains(c, TAG_2), wg, ch)
				if err == nil {
					util.SendByTelegramBot(fmt.Sprintf("名称 %v\n剧名 %v\n同步完成 (%v/%v)\n已用空间 %s", name, subName, index + 1, downloadedLen, util.GetUsedSpacePercentage()))
				} else {
					util.SendByTelegramBot(fmt.Sprintf("名称 %s\n同步错误 (%v/%v)\n错误原因：%s", name, index + 1, downloadedLen, err))
				}
			}(sourcePath, targetPath, tags, &wg, ch, index)
		}
	}
	wg.Wait()
	close(ch)
}

func getConfig() {
	err := godotenv.Load(util.GetRealAbsolutePath() + "/config.env")
	if err != nil {
		panic(err)
	}
	RCLONE_NAME = os.Getenv("RCLONE_NAME")
	RCLONE_LOCAL_DIR = os.Getenv("RCLONE_LOCAL_DIR")
	RCLONE_REMOTE_DIR = os.Getenv("RCLONE_REMOTE_DIR")
	MULTI_THREAD_STREAMS = os.Getenv("MULTI_THREAD_STREAMS")
	LOG_FILE = os.Getenv("LOG_FILE")
	TAG_1 = os.Getenv("TAG_1")
	TAG_2 = os.Getenv("TAG_2")
	THREAD = os.Getenv("THREAD")
}

func main() {
	util.Env()
	getConfig()
	qBitList = getList()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
				case <-ticker.C:
					qBitList = getList()
					util.SendByTelegramBot(fmt.Sprintf("查询到%v条信息", len(qBitList)))
				}
		}
	}()
	for {
		util.SendByTelegramBot(fmt.Sprintf("已用空间：%s ", util.GetUsedSpacePercentage()))
		sec := util.MeasureExecutionTime(mainTask)
		util.SendByTelegramBot(fmt.Sprintf("运行结束 本次耗时 %v", sec))
		time.Sleep(60 * time.Second)
	}
}