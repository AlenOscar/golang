/**
* @Author: Alen
* @Date: 2019-03-10 00:22
* @Description: TODO
 */

package logger

import "testing"

func TestLogger(t *testing.T) {
	logFile = LogInstance()
	logFile.SetLogOutFile("", "logger.log")
	logFile.SetLogLevel(InfoLevel)
	DEBUG("我是调试日志", 1, "abc")
	INFO("我是输出日志", 1, "abc")
	logFile.SetLogLevel(ErrorLevel)
	WARN("我是警告日志", 1, "abc")
	ERROR("我是错误日志", 1, "abc")
	FATAL("我是严重错误日志", 1, "abc")
}
