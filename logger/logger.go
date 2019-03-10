/*
===========================================
 1. 支持归档输出，一个小时压缩归档一份
 2. 最多保留三天的日志 // TODO
 3. 支持日志级别自定义
 4. 如果没有指定输出文件默认输出到控制台。
 5. 支持输出文件名行号，以及时间、日志界别
===========================================
*/

package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*=============*
 *  log color  *
 *=============*/
const (
	ColorRed = uint8(iota + 91)
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta //洋红
)

/*=============*
 *  log level  *
 *=============*/
const (
	FatalLevel int = iota + 1
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

/*=============*
 *  log func   *
 *=============*/
var (
	DEBUG func(format string, v ...interface{})
	INFO  func(format string, v ...interface{})
	WARN  func(format string, v ...interface{})
	ERROR func(format string, v ...interface{})
	FATAL func(format string, v ...interface{})
)

type Logger struct {
	innerLogger    *log.Logger //
	level          int         // 日志级别[1, 2, 3, 4, 5]
	lastRecordTime int64       // 最后一次归档时间
	logDir         string      // 日志路径
	fileName       string      // 日志文件名
	runtimeCaller  int         // 文件深度
	openColor      bool        // 颜色开关
	fileFd         *os.File    // 文件句柄
}

var logFile *Logger

func init() {
	setLogFunc(logDebug, logInfo, logWarning, logError, logFatal)
	log.SetFlags(log.Lshortfile)
	logFile = NewLog()
}

/***********************************【公有函数】***********************************/

// 创建
func NewLog() *Logger {
	logFile := new(Logger)
	logFile.level = DebugLevel
	logFile.lastRecordTime = int64(0)
	logFile.logDir = ""
	logFile.fileName = ""
	logFile.runtimeCaller = 2
	logFile.openColor = true
	//Logger.fileFd = nil
	return logFile
}

// 获得实例
func LogInstance() *Logger {
	if logFile == nil {
		logFile = NewLog()
	}
	return logFile
}

// 设置日志级别
func (lf *Logger) SetLogLevel(logLevel int) {
	lf.level = logLevel
}

// 获得日志级别
func (lf *Logger) GetLogLevel() int {
	return lf.level
}

// 日志颜色 - 开启/关闭
func (lf *Logger) SetOpenColor(open bool) {
	lf.openColor = open
}

// 设置日志文件
func (lf *Logger) SetLogOutFile(logDir string, fileName string) {
	projectPath := os.Getenv("GOPATH")
	if projectPath != "" {
		if len(strings.Split(projectPath, ":")) >= 0 {
			projectPath = strings.Split(projectPath, ":")[0]
		}
	}

	DEBUG("projectPath =", projectPath)

	lf.logDir = projectPath + "/logs/" // 默认日志路径
	lf.fileName = "log.log"

	if logDir != "" {
		lf.logDir = logDir
	}
	if fileName != "" {
		lf.fileName = fileName
	}

	if index := strings.LastIndex(lf.fileName, "/"); index != -1 {
		lf.logDir += lf.fileName[0:index]
		lf.fileName = lf.fileName[index:]
	}
	lf.fileName = strings.Split(lf.fileName, ".")[0]

	err := os.MkdirAll(lf.logDir, os.ModePerm) // 生成多级目录
	if err != nil {
		// TODO
		ERROR("creat log dir failed !")
		return
	}

	now := time.Now()
	filename := fmt.Sprintf("%s_%04d-%02d-%02d_%02d.log", lf.fileName, now.Year(), now.Month(), now.Day(), now.Hour())
	//if err := os.Rename(lf.logDir + lf.fileName, filename); err == nil {
	//	go func() {
	//		tarCmd := exec.Command("tar", "-zcf", filename+".tar.gz", filename, "--remove-files")
	//		tarCmd.Run()
	//
	//		rmCmd := exec.Command("/bin/sh", "-c", "find "+lf.logDir+` -type f -mtime +2 -exec rm {} \;`)
	//		rmCmd.Run()
	//	}()
	//}

	for index := 0; index < 10; index++ {
		if fd, err := os.OpenFile(lf.logDir+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeExclusive); err == nil {
			lf.fileFd = fd
			log.SetOutput(lf.fileFd)
			//lf.fileFd.Sync()
			//lf.fileFd.Close()
			break
		}
		lf.fileFd = nil
	}
}

// 重构Write函数
func (lf *Logger) Write(buf []byte) (n int, err error) {
	DEBUG("Write called ...")
	if lf.fileName == "" {
		return len(buf), nil
	}

	// TODO
	if lf.lastRecordTime+3600 < time.Now().Unix() { // 时间戳
		lf.SetLogOutFile(lf.logDir, lf.fileName)
		lf.lastRecordTime = time.Now().Unix()
	}

	if lf.fileFd == nil {
		return len(buf), nil
	}

	return lf.fileFd.Write(buf)
}

func (lf *Logger) deleteHistory() {
	// 尝试删除5天前的日志
	DEBUG("deleteHistory run")
	nowTime := time.Now()
	time5dAgo := nowTime.Add(-1 * time.Hour * 24 * 5)

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filePath := dir + "/logs/" + lf.fileName + "." + time5dAgo.Format("2006-01-02")

	_, err := os.Stat(filePath)
	if err == nil {
		os.Remove(filePath)
	}
}

/***********************************【私有函数】***********************************/

// 设置日志函数
func setLogFunc(debugFunc, infoFunc, warnFunc,
	errFunc, fatalFunc func(format string, v ...interface{})) {
	DEBUG = debugFunc
	INFO = infoFunc
	WARN = warnFunc
	ERROR = errFunc
	FATAL = fatalFunc
}

// 格式化日志前缀
func formatLogPrefix(prefix string, level int) string {
	if logFile.openColor {
		switch level {
		case DebugLevel:
			prefix = green(prefix)
		case InfoLevel:
			prefix = blue(prefix)
		case WarnLevel:
			prefix = yellow(prefix)
		case ErrorLevel:
			prefix = magenta(prefix)
		case FatalLevel:
			prefix = red(prefix)
		default:
			prefix = green(prefix) // TODO: 默认颜色
		}
	}
	return time.Now().Format("[2006-01-02 15:04:05.000000]") + " " + prefix + " "
}

// debug日志
func logDebug(format string, args ...interface{}) {
	if logFile.level >= DebugLevel {
		log.SetPrefix(formatLogPrefix("[DEBUG]", DebugLevel))
		msg := fmt.Sprintf(format+" %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
		log.Output(logFile.runtimeCaller, msg)
	}
}

// info日志
func logInfo(format string, args ...interface{}) {
	if logFile.level >= InfoLevel {
		log.SetPrefix(formatLogPrefix(" [INFO]", InfoLevel))
		msg := fmt.Sprintf(format+" %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
		log.Output(logFile.runtimeCaller, msg)
	}
}

// warn日志
func logWarning(format string, args ...interface{}) {
	if logFile.level >= WarnLevel {
		log.SetPrefix(formatLogPrefix(" [WARN]", WarnLevel))
		msg := fmt.Sprintf(format+" %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
		log.Output(logFile.runtimeCaller, msg)
	}
}

// err日志
func logError(format string, args ...interface{}) {
	if logFile.level >= ErrorLevel {
		log.SetPrefix(formatLogPrefix("[ERROR]", ErrorLevel))
		msg := fmt.Sprintf(format+" %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
		log.Output(logFile.runtimeCaller, msg)
	}
}

// fatal日志
func logFatal(format string, args ...interface{}) {
	if logFile.level >= FatalLevel {
		log.SetPrefix(formatLogPrefix("[FATAL]", FatalLevel))
		msg := fmt.Sprintf(format+" %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
		log.Output(logFile.runtimeCaller, msg)
	}
}

// 红色
func red(str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorRed, str)
}

// 绿色
func green(str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorGreen, str)
}

// 黄色
func yellow(str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorYellow, str)
}

// 蓝色
func blue(str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorBlue, str)
}

// 洋红
func magenta(str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorMagenta, str)
}

func _getCallerLine() (string, int) {
	skip := 2
	// 获取调用者的信息
	_, file, line, _ := runtime.Caller(skip)
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	// DEBUG 文件名输出固定长度
	var fileBytes [10]byte
	for i := 0; i < 10; i++ {
		if i < len(short) {
			fileBytes[i] = short[i]
		} else {
			fileBytes[i] = byte(32)
		}
	}
	file = string(fileBytes[:])
	return file, line
}

func _toString(args ...interface{}) string {
	var data string
	for _, arg := range args {
		msgJson, err := json.Marshal(arg)
		if err != nil {
			continue
		}
		data += " " + string(msgJson)
	}
	return data
}
