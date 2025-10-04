package utils

import (
	"fmt"

	"xquant/pkg/log"
)

func CatchPanic(format string, args ...interface{}) {
	if err := recover(); err != nil {
		// 拼接错误信息：原格式化信息 + Panic 内容
		errMsg := fmt.Sprintf(format, args...)
		log.Errorf("%s, panic: %v", errMsg, err)
	}
}
