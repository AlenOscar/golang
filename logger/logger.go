/*
===========================================
 1. 支持归档输出，一个小时压缩归档一份
 2. 最多保留三天的日志
 3. 支持日志级别自定义
 4. 如果没有指定输出文件默认输出到控制台。
 5. 支持输出文件名行号，以及时间、日志界别
===========================================
*/

package logger

import (
        "fmt"
        "log"
        "os"
        "os/exec"
        "strings"
        "time"
        "encoding/json"
        "runtime"
)

/*=============*
 *  log color  *
 *=============*/
const (
        color_red = uint8(iota + 91)
        color_green
        color_yellow
        color_blue
        color_magenta //洋红
)
/*=============*
 *  log level  *
 *=============*/
const (
        DebugLevel int = iota
        InfoLevel
        WarnLevel
        ErrorLevel
        FatalLevel
)

type LogFile struct {
        level     int
        logTime   int64
        fileName  string
        openColor bool
        fileFd    *os.File
}

var logFile LogFile

func red(s string) string {
        return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color_red, s)
}

func green(s string) string {
        return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color_green, s)
}

func yellow(s string) string {
        return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color_yellow, s)
}

func blue(s string) string {
        return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color_blue, s)
}

func magenta(s string) string {
        return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color_magenta, s)
}

var (
        DEBUG    func(format string, v ...interface{})
        INFO     func(format string, v ...interface{})
        WARN     func(format string, v ...interface{})
        ERROR    func(format string, v ...interface{})
        FATAL    func(format string, v ...interface{})
)

func init() {
        SetLogFunc(LOG_DEBUG, LOG_INFO, LOG_WARN, LOG_ERROR, LOG_FATAL)
        log.SetFlags(log.Lshortfile)
        logFile.openColor = true
        // log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func _get_call_line() (string, int) {
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
        var file_bytes [10]byte
        for i := 0; i < 10; i++ {
                if i < len(short) {
                        file_bytes[i] = short[i]
                } else {
                        file_bytes[i] = byte(32)
                }
        }
        file = string(file_bytes[:])
        return file, line
}

func _toString(args ...interface{}) string {
        var data string;
        for _, arg := range args {
                msg_json, err := json.Marshal(arg)
                if err != nil {
                        continue
                }
                data += " " + string(msg_json);
        }
        return data
}

func SetLogFunc(debug, info, warn, error, fatal func(format string, v ...interface{})) {
        DEBUG = debug
        INFO  = info
        WARN  = warn
        ERROR = error
        FATAL = fatal
}

func SetLogFile(fileName string, level int) error {
        file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
        if err != nil {
                return err
        }
        logFile.fileName = fileName
        logFile.level = level
        logFile.openColor = false
        log.SetOutput(file)
        return nil
}

func SetLevel(level int) {
        logFile.level = level
}

func GetLevel() int {
        return logFile.level
}

func FormatLogPrefix(prefix string, level int) string {
        if (logFile.openColor) {
                switch (level) {
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
                        prefix = green(prefix)
                }
        }
        return time.Now().Format("[2006-01-02 15:04:05.000000]") + " " + prefix + " "
}

func LOG_DEBUG(format string, args ...interface{}) {
        if DebugLevel >= logFile.level {
                // file, line := _get_call_line()
                log.SetPrefix(FormatLogPrefix("[DEBUG]", DebugLevel))
                msg := fmt.Sprintf(format + " %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
                log.Output(2, msg)
        }
}

func LOG_INFO(format string, args ...interface{}) {
        if InfoLevel >= logFile.level {
                log.SetPrefix(FormatLogPrefix("[INFO]", InfoLevel))
                msg := fmt.Sprintf(format + " %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
                log.Output(2, msg)
        }
}

func LOG_WARN(format string, args ...interface{}) {
        if WarnLevel >= logFile.level {
                log.SetPrefix(FormatLogPrefix("[WARN]", WarnLevel))
                msg := fmt.Sprintf(format + " %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
                log.Output(2, msg)
        }
}

func LOG_ERROR(format string, args ...interface{}) {
        if ErrorLevel >= logFile.level {
                log.SetPrefix(FormatLogPrefix("[ERROR]", ErrorLevel))
                msg := fmt.Sprintf(format + " %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
                log.Output(2, msg)
        }
}

func LOG_FATAL(format string, args ...interface{}) {
        if FatalLevel >= logFile.level {
                log.SetPrefix(FormatLogPrefix("[FATAL]", FatalLevel))
                msg := fmt.Sprintf(format + " %s", strings.TrimRight(fmt.Sprintln(args...), "\n"))
                log.Output(2, msg)
        }
}

// 重构 Write 函数
func (file LogFile) Write(buf []byte) (n int, err error) {
        fmt.Println("Write func call")
        if file.fileName == "" {
                return len(buf), nil
        }

        if logFile.logTime + 3600 < time.Now().Unix() { // 时间戳
                logFile.createLogFile()
                logFile.logTime = time.Now().Unix()
        }

        if logFile.fileFd == nil {
                return len(buf), nil
        }

        return logFile.fileFd.Write(buf)
}

func (file *LogFile) createLogFile() {
        logdir := "./logs/" // 文件路径
        path := logdir
        if index := strings.LastIndex(file.fileName, "/"); index != -1 {
                path = path + file.fileName[0:index]
                err := os.MkdirAll(path, os.ModePerm) // 生成多级目录
                if err != nil {
                        LOG_ERROR("creat log dir failed !")
                        return
                }
        }

        now := time.Now()
        filename := logdir + fmt.Sprintf("%s_%04d%02d%02d_%02d%02d", file.fileName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
        if err := os.Rename(path, filename); err == nil {
                go func() {
                        tarCmd := exec.Command("tar", "-zcf", filename+".tar.gz", filename, "--remove-files")
                        tarCmd.Run()

                        rmCmd := exec.Command("/bin/sh", "-c", "find "+logdir+` -type f -mtime +2 -exec rm {} \;`)
                        rmCmd.Run()
                }()
        }

        for index := 0; index < 10; index++ {
                if fd, err := os.OpenFile(file.fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeExclusive); nil == err {
                        file.fileFd.Sync()
                        file.fileFd.Close()
                        file.fileFd = fd
                        break
                }
                file.fileFd = nil
        }
}
