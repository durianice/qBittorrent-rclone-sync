package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func ParseJsonStr(jsonStr string) []map[string]interface{} {

	var unknownObjects []json.RawMessage

	err := json.Unmarshal([]byte(jsonStr), &unknownObjects)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var array []map[string]interface{}

	for _, rawMsg := range unknownObjects {
		var obj map[string]interface{}
		err := json.Unmarshal(rawMsg, &obj)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		array = append(array, obj)
	}

	// fmt.Println("Parsed objects:", array)

	return array
}

func RunShellCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Error cmd:", command)
		return "", err
	}
	return string(output), nil
}

func RunRcloneCommand(command string, syncMsg string, flag string) error {
	cmd := exec.Command("bash", "-c", command)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败：%s", err)
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("启动命令失败：%s", err)
	}
	reader := bufio.NewReader(stdoutPipe)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取命令输出失败：%s", err)
		}
		if strings.Contains(strings.ToLower(line), "error") || strings.Contains(strings.ToLower(line), "fail") {
			Notify(fmt.Sprintf("🥵 同步发生错误 %v \n", line), line)
			return errors.New(line)
		}
		if !strings.Contains(line, "ETA") {
			continue
		}
		syncProcess := getFormatOutput(line)
		if strings.Contains(syncProcess, "100%") {
			syncMsg = strings.ReplaceAll(syncMsg, "🤡 在同步了", "✅ 完成")
		}
		Notify(fmt.Sprintf("%v", syncMsg+"\n\n🎃 实时进度\n"+syncProcess), flag)
		if err == io.EOF || strings.Contains(syncProcess, "100%") {
			go func() {
				time.Sleep(60 * time.Second)
				DeleteMsg(flag)
			}()
			break
		}
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("命令执行失败：%s", err)
	}
	return nil
}

func getFormatOutput(input string) string {
	parts := strings.Split(input, "Transferred:")
	if len(parts) != 0 {
		re := regexp.MustCompile(`\s+`)
		trimmed := strings.ReplaceAll(parts[1], " ", "")
		trimmed = re.ReplaceAllString(trimmed, "")
		trimmed = strings.ReplaceAll(trimmed, "ETA", "预计剩余时间 ")
		return strings.ReplaceAll(trimmed, ",", "\n")
	}
	return ""
}

func Env() {
	switch runtime.GOOS {
	case "windows":
		panic("Windows not support")
	case "linux":
		fmt.Println("Running on Linux")
	case "darwin":
		panic("MacOS not support")
	default:
		panic("Current OS not support")
	}
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false // 无法确定文件是否存在
}

func CreateDirIfNotExist(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Printf("Create '%v' Error : %v", dirPath, err)
			os.Exit(1)
		}
		fmt.Println("Directory created:", dirPath)
	} else {
		fmt.Println("Directory already exists:", dirPath)
	}
}

func CreateFileIfNotExist(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// 文件不存在
		file, err := os.Create(filepath)
		if err != nil {
			// 创建文件时出错
			return err
		}
		defer file.Close()
	}
	// 文件已存在或已成功创建
	return nil
}

func GetFreeSpace(dir string, unit string) (int, error) {
	command := fmt.Sprintf("df --output=avail %v | tail -n 1", dir)
	freeSpaceKBStr, _ := RunShellCommand(command)
	freeSpaceKB := 0
	fmt.Sscanf(freeSpaceKBStr, "%d", &freeSpaceKB)

	switch unit {
	case "KB":
		return freeSpaceKB, nil
	case "MB":
		return freeSpaceKB / 1024, nil
	case "GB":
		return freeSpaceKB / 1024 / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}
}

func GetUsedSpacePercentage(disk string) string {
	command := fmt.Sprintf("df --output=pcent %v | tail -n 1", disk)
	usedStr, _ := RunShellCommand(command)
	usedStr = strings.ReplaceAll(usedStr, " ", "")
	usedStr = strings.ReplaceAll(usedStr, "\n", "")
	return usedStr
}

func PercentageToDecimal(percentageStr string) (float64, error) {
	percentageStr = strings.ReplaceAll(percentageStr, " ", "")
	percentageStr = strings.ReplaceAll(percentageStr, "\n", "")
	percentageStr = strings.TrimRight(percentageStr, "%")
	percentage, err := strconv.ParseFloat(percentageStr, 64)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	decimal := percentage / 100
	return decimal, nil
}

func MeasureExecutionTime(function func()) time.Duration {
	startTime := time.Now()
	function()
	elapsed := time.Since(startTime)
	return elapsed
}

func GetRealAbsolutePath() string {
	res, _ := RunShellCommand("pwd")
	res = strings.ReplaceAll(res, "\n", "")
	return res
}

func Filter[T any](array []T, condition func(T) bool) []T {
	var result []T
	for _, item := range array {
		if condition(item) {
			result = append(result, item)
		}
	}
	return result
}

func Map[T any](array []T, mapper func(T) T) []T {
	var result []T
	for _, item := range array {
		result = append(result, mapper(item))
	}
	return result
}

func toString(i interface{}) (string, error) {
	i = indirectToStringerOrError(i)

	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(s), 10), nil
	case []byte:
		return string(s), nil
	case template.HTML:
		return string(s), nil
	case template.URL:
		return string(s), nil
	case template.JS:
		return string(s), nil
	case template.CSS:
		return string(s), nil
	case template.HTMLAttr:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}

func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

type Release struct {
	TagName string `json:"tag_name"`
}

func GetLatestRelease(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	header := map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}
	params := map[string]interface{}{} // No params needed for this request

	response, err := Get(url, header, params)
	if err != nil {
		return "", err
	}

	var release Release
	err = json.Unmarshal([]byte(response), &release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func IsVersionOutdated(currentVersion, latestVersion string) (bool, error) {
	currentParts := strings.Split(strings.TrimPrefix(currentVersion, "v"), ".")
	latestParts := strings.Split(strings.TrimPrefix(latestVersion, "v"), ".")

	minLength := len(currentParts)
	if len(latestParts) < minLength {
		minLength = len(latestParts)
	}

	for i := 0; i < minLength; i++ {
		currentInt, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return false, err
		}
		latestInt, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return false, err
		}
		if currentInt < latestInt {
			return true, nil
		} else if currentInt > latestInt {
			return false, nil
		}
	}

	// If all parts so far were equal, the shorter version is outdated if it has fewer parts.
	return len(currentParts) < len(latestParts), nil
}

// CheckPathStatus checks if a path is an empty directory or if a file exists.
func CheckPathStatus(path string) (bool, error) {
	// Check if the path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// The path does not exist
			return true, nil
		}
		// Other error
		return true, err
	}

	// Check if the path is a file
	if !info.IsDir() {
		// The path is a file and it exists
		return false, nil
	}

	// The path is a directory, now check if it's empty
	isEmpty := true
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Found a file, the directory is not empty
		if !d.IsDir() {
			isEmpty = false
			// Return a special error to stop walking the directory
			return fs.SkipDir
		}

		return nil
	})

	if err != nil && err != fs.SkipDir {
		// An error occurred that wasn't the special 'SkipDir' error
		return false, err
	}

	// If we got here, the directory is either empty or we encountered 'SkipDir'
	return isEmpty, nil
}

func Trim(originalString string) string {
	reg := regexp.MustCompile(`\s+`)
	noSpacesString := reg.ReplaceAllString(originalString, "")
	return noSpacesString
}
