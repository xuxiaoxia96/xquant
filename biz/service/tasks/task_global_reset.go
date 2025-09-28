package services

import (
	"runtime/debug"

	"gitee.com/quant1x/gotdx"

	"xquant/pkg/cache"
	"xquant/pkg/factors"
	"xquant/pkg/log"
)

// 任务 - 交易日数据缓存重置
func jobGlobalReset() {
	defer func() {
		// 1. 捕获可能的 panic
		if err := recover(); err != nil {
			// 2. 记录 panic 详情（错误信息 + 调用栈），便于调试
			log.Errorf("jobGlobalReset 执行过程中发生异常，已安全捕获: %v", err)
			log.Errorf("异常调用栈:\n%s", debug.Stack()) // 打印调用栈，定位具体出错代码行
			// 3. （可选）若需触发后续降级逻辑，可在此处添加（如标记缓存重置失败）
			// 例如：log.Println("缓存重置任务失败，触发降级流程...")
		}
	}()

	log.Infof("系统初始化...")
	log.Infof("清理过期的更新状态文件...")
	//_ = cleanExpiredStateFiles()
	log.Infof("清理过期的更新状态文件...OK")
	gotdx.ReOpen()
	log.Infof("重置系统缓存...")
	factors.SwitchDate(cache.DefaultCanReadDate())
	log.Infof("重置系统缓存...OK")

	log.Infof("系统初始化...OK")
}
