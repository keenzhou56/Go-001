// 自定义函数库
package common

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 调试输出
func Vd(debug bool, argv ...interface{}) {
	if debug {
		Println(argv)
	}
}

// 输出数据
func Println(argv ...interface{}) {
	fmt.Println(GetTimestamp(), argv)
}

// 获取当前时间戳
func GetTime() int64 {
	return time.Now().Unix()
}

// 获取当前格式化时间
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 格式化运行时间
func FormateRunTime(second int64) time.Duration {
	return time.Duration(second * 1000 * 1000 * 1000)
}

// 格式化unixtime
func FormatUnixTime(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}

// 检查文件夹是否存在
func CheckPathExist(pathName string) bool {
	if fi, err := os.Stat(pathName); err == nil {
		return fi.IsDir()
	}

	return false
}

// 检查文件是否存在
func CheckFileExist(filename string) bool {
	if fi, err := os.Stat(filename); err == nil {
		return !fi.IsDir()
	}

	return false
}

// 遍历文件夹，读取所有文件
func ReadDir(pathname string) ([]string, error) {
	files := make([]string, 0)

	if !CheckPathExist(pathname) {
		return files, errors.New("path is not exists:" + pathname)
	}

	filepath.Walk(pathname,
		func(path string, f os.FileInfo, err error) error {
			if f == nil || f.IsDir() {
				return err
			}
			files = append(files, path)
			return nil
		})

	return files, nil
}

// 执行一个shell命令，不关心返回
func CommandStart(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Start()
	return err
}

// 执行shell命令
func Command(command string, argv []string) ([]byte, error) {
	cmd := exec.Command(command, argv...)
	output, err := cmd.CombinedOutput()

	return output, err
}

// 获取进程的pids
func GetPids(command string) []int {
	pids := make([]int, 0)

	cmd := exec.Command("/bin/sh", "-c", command)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return pids
	}

	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")

		ft := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				ft = append(ft, t)
			}
		}

		pid, err := strconv.Atoi(ft[1])
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	return pids
}

// 获取当前路径
func GetDirPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)

	return filepath.Dir(path)
}

// 获取im的根目录
func GetImPath() string {
	return filepath.Dir(GetDirPath())
}

// 读取配置文件
func LoadConf(filename string) (map[string]string, error) {
	configs := make(map[string]string)

	if !CheckFileExist(filename) {
		return configs, errors.New("config file not exists:" + filename)
	}

	// 打开配置文件
	f, _ := os.Open(filename)
	defer f.Close()
	// 读取文件到buffer里边
	buf := bufio.NewReader(f)
	for {
		// 按照换行读取每一行
		l, err := buf.ReadString('\n')

		// 跳过空行
		if l == "\n" {
			continue
		}

		// 相当于PHP的trim
		line := strings.TrimSpace(l)
		// 判断退出循环
		if err != nil {
			if err != io.EOF {
				//return err
				panic(err)
				//continue
			}
			if len(line) == 0 {
				break
			}
		}

		lineSplit := strings.SplitN(line, "=", 3)
		// 跳过错误的行
		if len(lineSplit) != 2 {
			continue
		}

		configs[strings.TrimSpace(lineSplit[0])] = strings.TrimSpace(lineSplit[1])
	}

	return configs, nil
}

// 去读配置文件的某个配置
func LoadConfValue(filename string, key string) (string, error) {

	configs, err := LoadConf(filename)
	if err != nil {
		return "", errors.New("load config file error.")
	}

	value, exits := configs[key]
	if !exits {
		return "", errors.New("config key not found:" + key)
	}

	return value, nil
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// 截取字符串函数
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

// 获取token值
func GetToken(encryptKey string, userId int, time int64) string {
	tokenStr := fmt.Sprintf("%s#%d#%d", encryptKey, userId, time)
	md5Str := GetMd5String(tokenStr)
	return Substr(md5Str, 0, 4)
}
