package main

import (
	"fmt"
	"os"
	"qbittorrentRcloneSync/http"
	"qbittorrentRcloneSync/util"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	DISK_LOCAL string
	MAX_MEM string
	MIN_MEM string
)

const CATEGORY_1 = "_电影"
const CATEGORY_2 = "_电视节目"

var counter Counter

var qBitList []map[string]interface{}

func rcloneTask(sourceFile string, targetFile string, keepSourceFile bool, wg *sync.WaitGroup, ch chan struct{}, syncMsg string) error {
	defer wg.Done()
	defer counter.Decrease()
	option := "moveto"
	if keepSourceFile {
		option = "copyto"
	}
	command := fmt.Sprintf("/usr/bin/rclone -v -P %s --multi-thread-streams %s --log-file %q %q %q", option,
	MULTI_THREAD_STREAMS, LOG_FILE, sourceFile, targetFile)
	err := util.RunRcloneCommand(command, syncMsg, sourceFile)
	<-ch
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
		category, _ := obj["category"].(string)
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
			"category": category,
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

type Counter struct {
    val int32
}

func (c *Counter) Increase() {
    atomic.AddInt32(&c.val, 1)
}

func (c *Counter) Decrease() {
    atomic.AddInt32(&c.val, -1)
}

func (c *Counter) Value() int32 {
    return atomic.LoadInt32(&c.val)
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
		category, _ := obj["category"].(string)
		downloadPath, _ := obj["downloadPath"].(string)
		savePath, _ := obj["savePath"].(string)
		subListDownloaded, _ := obj["subListDownloaded"].([]map[string]interface{})
		downloadedLen := len(subListDownloaded)
		for index, subObj := range subListDownloaded {
			subName, _ := subObj["name"].(string)
			sourcePath := downloadPath + "/" + subName
			targetPath := RCLONE_NAME + RCLONE_REMOTE_DIR + category2Path(category) + subName
			localTargetPath := RCLONE_LOCAL_DIR + RCLONE_REMOTE_DIR + category2Path(category) + subName
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
			counter.Increase()
			go func(a string, b string, c string, wg *sync.WaitGroup, ch chan struct{}, index int) {
				// util.Notify(fmt.Sprintf("正在同步 (%v/%v)\n一级名称 %v\n二级名称 %v\n已用空间 %s", index + 1, downloadedLen, name, subName, util.GetUsedSpacePercentage(DISK_LOCAL)), subName)
				syncMsg := fmt.Sprintf("🔵同步 (%v/%v)\n一级名称 %v\n二级名称 %v\n已用空间 %s", index + 1, downloadedLen, name, subName, util.GetUsedSpacePercentage(DISK_LOCAL))
				err := rcloneTask(a, b, strings.Contains(c, TAG_2), wg, ch, syncMsg)
				if err != nil {
					util.Notify(fmt.Sprintf("❌同步错误 (%v/%v)\n一级名称 %v\n二级名称 %v \n错误原因 %v", index + 1, downloadedLen, name, subName, err), "")
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
	DISK_LOCAL = os.Getenv("DISK_LOCAL")
	MAX_MEM = os.Getenv("MAX_MEM")
	MIN_MEM = os.Getenv("MIN_MEM")
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

func main() {
	util.Env()
	getConfig()
	counter = Counter{}
	qBitList = getList()
	http.CreateCategory(CATEGORY_1, "")
	http.CreateCategory(CATEGORY_2, "")
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
				case <-ticker.C:
					qBitList = getList()
					util.Notify(fmt.Sprintf("查询到%v条信息", len(qBitList)), "查询")
					util.Notify(fmt.Sprintf("当前线程情况(%v/%v)", counter.Value(), THREAD), "线程")
				}
		}
	}()
	for {
		util.Notify(fmt.Sprintf("已用空间：%s ", util.GetUsedSpacePercentage(DISK_LOCAL)), "空间")
		sec := util.MeasureExecutionTime(mainTask)
		util.Notify(fmt.Sprintf("运行结束 本次耗时 %v", sec), "")
		time.Sleep(60 * time.Second)
	}
}