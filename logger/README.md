安装:
``` bash
$ go get github.com/alenoscar/golang/logger
```

example:
``` bash
import (
    log "github.com/alenoscar/golang/logger"
)

func main() {
    // log.SetLogFile("test.log", log.GetLevel())
    log.DEBUG("我是调试日志", 1, "abc")
    log.INFO("我是输出日志", 1, "abc")
    log.SetLevel(log.ErrorLevel)
    log.WARN("我是警告日志", 1, "abc")
    log.ERROR("我是错误日志", 1, "abc")
    log.FATAL("我是严重错误日志", 1, "abc")
}
```