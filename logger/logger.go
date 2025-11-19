package logger

import "github.com/urie96/xiaozhi-server-go/core/utils"

var logger *utils.Logger

func init() {
	// 初始化日志系统
	var err error
	logger, err = utils.NewLogger("debug")
	if err != nil {
		panic(err)
	}
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}
