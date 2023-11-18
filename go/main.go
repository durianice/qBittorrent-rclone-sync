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
	RCLONE_NAME          string
	RCLONE_LOCAL_DIR     string
	RCLONE_REMOTE_DIR    string
	MULTI_THREAD_STREAMS string
	LOG_FILE             string
	THREAD               string
	DISK_LOCAL           string
	MAX_MEM              string
	MIN_MEM              string
	QBIT_DIR             string
)

const CATEGORY_1 = "_电影"
const CATEGORY_2 = "_电视节目"
const STAY_TAG = "保种"
const CTRL_TAG = "脚本控制"

const currentVersion = "v1.2.9"

var qBitList []map[string]interface{}

func rcloneTask(sourceFile string, targetFile string, keepSourceFile bool, syncMsg string) error {
	option := "moveto"
	if keepSourceFile {
		option = "copyto"
	}
	log_level := "ERROR"
	// %s%s%s 防止路径中有全角字符，使用%q会转换为Unicode
	command := fmt.Sprintf("/usr/bin/rclone -P %s --multi-thread-streams %s --log-file %q --log-level %q %s%s%s %s%s%s", option, MULTI_THREAD_STREAMS, LOG_FILE, log_level, "\"", sourceFile, "\"", "\"", targetFile, "\"")
	util.Notify(fmt.Sprintf("🍪 正在你的小鸡上执行\n%s\n", command), "")
	err := util.RunRcloneCommand(command, syncMsg, sourceFile)
	if err != nil {
		return err
	}
	return nil

}

func memoryControl() string {
	used := util.GetUsedSpacePercentage(DISK_LOCAL)
	res, _ := util.PercentageToDecimal(used)
	MAX, _ := util.PercentageToDecimal(MAX_MEM)
	MIN, _ := util.PercentageToDecimal(MIN_MEM)
	if res >= MAX {
		return "P"
	}
	if res < MIN {
		return "D"
	}
	return "N"
}

func getList() []map[string]interface{} {
	http.Login()
	list := http.GetInfo()
	// 按标签过滤
	inCtrlList := util.Filter(list, func(obj map[string]interface{}) bool {
		dir := obj["content_path"].(string)
		progress := obj["progress"].(float64)
		isEmpty, err := util.IsDirectoryEmpty(dir)
		if err != nil {
			fmt.Println(err)
		}
		if isEmpty && progress == 1 {
			http.DeleteTorrents(obj["hash"].(string))
			util.Notify(fmt.Sprintf("%v\n😁 这个同步完了，删除本地空目录和torrents任务\n", dir), "")
		}
		return strings.Contains(obj["tags"].(string), CTRL_TAG) || strings.Contains(obj["category"].(string), CATEGORY_1) || strings.Contains(obj["category"].(string), CATEGORY_2)
	})
	res := util.Map(inCtrlList, func(obj map[string]interface{}) map[string]interface{} {
		name, _ := obj["name"].(string)
		hash, _ := obj["hash"].(string)
		tags, _ := obj["tags"].(string)
		category, _ := obj["category"].(string)
		seqDl, _ := obj["seq_dl"].(bool)
		state, _ := obj["state"].(string)
		downloadPath, _ := obj["download_path"].(string)
		contentPath, _ := obj["content_path"].(string)
		savePath, _ := obj["save_path"].(string)
		// 过滤已下载完成的子内容
		subListDownloaded := util.Filter(http.GetDetail(hash), func(obj map[string]interface{}) bool {
			return obj["progress"].(float64) == 1
		})
		subListDownloaded = util.Map(subListDownloaded, func(subObj map[string]interface{}) map[string]interface{} {
			subName, _ := subObj["name"].(string)
			return map[string]interface{}{
				"name":         name,
				"subName":      subName,
				"hash":         hash,
				"tags":         tags,
				"category":     category,
				"seqDl":        seqDl,
				"state":        state,
				"downloadPath": downloadPath,
				"savePath":     savePath,
				"contentPath":  contentPath,
			}
		})
		memState := memoryControl()
		if memState == "P" && state == "downloading" {
			util.Notify("🤢 内存不够了暂停一下先", "")
			http.Pause(hash)
		}
		if memState == "D" && state != "downloading" {
			util.Notify("😸 元气满满，恢复下载", "")
			http.Resume(hash)
		}
		if !seqDl {
			http.ToggleSequentialDownload(hash)
			util.Notify("🥶 已强制按顺序下载，不然鸡爆了", "")
		}
		return map[string]interface{}{
			"subListDownloaded": subListDownloaded,
		}
	})
	var r []map[string]interface{}
	for _, obj := range res {
		subListDownloaded, _ := obj["subListDownloaded"].([]map[string]interface{})
		for _, subObj := range subListDownloaded {
			r = append(r, subObj)
		}
	}
	return r
}

func mainTask() {
	var wg sync.WaitGroup
	THREAD, err := strconv.Atoi(THREAD)
	if err != nil {
		panic("Error converting THREAD to int")
	}
	ch := make(chan struct{}, THREAD)

	total := len(qBitList)
	for index, obj := range qBitList {
		contentPath, _ := obj["contentPath"].(string)
		isEmpty, _ := util.IsDirectoryEmpty(contentPath)
		if isEmpty {
			util.Notify(fmt.Sprintf("%v\n😓 这目录暂时没东西可以同步，下一个", contentPath), "")
			continue
		}

		name, _ := obj["name"].(string)
		tags, _ := obj["tags"].(string)
		category, _ := obj["category"].(string)
		downloadPath, _ := obj["downloadPath"].(string)
		if QBIT_DIR != "" {
			downloadPath = QBIT_DIR
		}
		savePath, _ := obj["savePath"].(string)
		subName, _ := obj["subName"].(string)
		sourcePath := downloadPath + "/" + subName
		targetPath := RCLONE_NAME + RCLONE_REMOTE_DIR + category2Path(category) + subName
		localTargetPath := RCLONE_LOCAL_DIR + RCLONE_REMOTE_DIR + category2Path(category) + subName
		if !util.FileExists(sourcePath) {
			sourcePath = savePath + "/" + subName
			if !util.FileExists(sourcePath) {
				// util.Notify(fmt.Sprintf("%v\n未找到或已同步该资源", sourcePath), "")
				continue
			}
		}
		if util.FileExists(localTargetPath) {
			if util.FileExists(sourcePath) {
				if strings.Contains(tags, STAY_TAG) {
					util.Notify(fmt.Sprintf("%v\n😵‍💫 同步过了，保下种", sourcePath), "")
				} else {
					command := fmt.Sprintf("sudo rm %q", sourcePath)
					util.RunShellCommand(command)
					util.Notify(fmt.Sprintf("%v\n😅 同步过了，不保种，删了", sourcePath), "")
				}
			}
			continue
		}
		ch <- struct{}{}
		wg.Add(1)
		util.Notify("🙀 准备启动Rclone", "")
		go func(ID int) {
			defer wg.Done()
			defer func() { <-ch }()
			syncMsg := fmt.Sprintf("🤡 在同步了 (%v/%v)\n%v\n%v", ID, total, name, subName)
			err := rcloneTask(sourcePath, targetPath, strings.Contains(tags, STAY_TAG), syncMsg)
			if err != nil {
				util.Notify(fmt.Sprintf("🥵 同步错误 (%v/%v)\n%v\n%v \n错误原因 %v", ID, total, name, subName, err), "")
				return
			}
		}(index + 1)
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
	THREAD = os.Getenv("THREAD")
	DISK_LOCAL = os.Getenv("DISK_LOCAL")
	MAX_MEM = os.Getenv("MAX_MEM")
	MIN_MEM = os.Getenv("MIN_MEM")
	QBIT_DIR = os.Getenv("QBIT_DIR")
}

func category2Path(category string) string {
	if category == CATEGORY_1 {
		return "movie/"
	}
	if category == CATEGORY_2 {
		return "tv/"
	}
	return ""
}

func checkVersion() {
	owner := "durianice"
	repo := "qBittorrent-rclone-sync"

	latestVersion, err := util.GetLatestRelease(owner, repo)
	if err != nil {
		fmt.Printf("获取版本信息失败: %s\n", err)
		os.Exit(1)
		return
	}

	outdated, err := util.IsVersionOutdated(currentVersion, latestVersion)
	if err != nil {
		fmt.Printf("版本信息比较失败: %s\n", err)
		return
	}
	if outdated {
		url := "https://github.com/durianice/qBittorrent-rclone-sync#%E5%AE%89%E8%A3%85%E6%9B%B4%E6%96%B0"
		util.Notify(fmt.Sprintf("😆 发现新的版本 %s\n\n当前版本 %s\n\n<a href='%s'>前往更新</a>", latestVersion, currentVersion, url), "")
		for _, obj := range qBitList {
			http.Pause(obj["hash"].(string))
		}
		util.Notify("😄 已暂停全部下载，脚本退出", "")
		os.Exit(1)
	} else {
		util.Notify(fmt.Sprintf("😄 当前为最新版本 %s", latestVersion), "")
	}
}

func main() {
	util.Env()
	util.Notify("🤠 欢迎使用", "")
	getConfig()
	util.CreateFileIfNotExist(LOG_FILE)
	qBitList = getList()
	http.CreateCategory(CATEGORY_1, "")
	http.CreateCategory(CATEGORY_2, "")
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				checkVersion()
				qBitList = getList()
				util.Notify(fmt.Sprintf("🤖 查询到 %v 个已下载文件", len(qBitList)), "")
				util.Notify(fmt.Sprintf("🫣 小鸡已用空间：%s ", util.GetUsedSpacePercentage(DISK_LOCAL)), "")
			}
		}
	}()
	for {
		sec := util.MeasureExecutionTime(mainTask)
		util.Notify(fmt.Sprintf("💩 跑完一遍了 花了 %v", sec), "")
		time.Sleep(60 * time.Second)
	}
}
