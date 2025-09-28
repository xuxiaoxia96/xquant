package services

import (
	"sync"

	"gitee.com/quant1x/exchange"

	"xquant/pkg/datasource/base"
	"xquant/pkg/log"
	"xquant/pkg/market"
	"xquant/pkg/models"
)

// 任务 - 实时更新K线
func jobRealtimeKLine() {
	funcName := "jobRealtimeKLine"
	updateInRealTime, status := exchange.CanUpdateInRealtime()
	// 14:30:00~15:01:00之间更新数据
	if updateInRealTime && IsTrading(status) {
		realtimeUpdateOfKLine()
	} else {
		// TODO: 调试阶段也可以更新
		log.Infof("%s, 非尾盘交易时段: %d", funcName, status)
	}
}

// 更新K线
func realtimeUpdateOfKLine() {
	// 原生panic捕获，替代runtime.IgnorePanic
	defer func() {
		if err := recover(); err != nil {
			// 可根据需要添加错误日志
			// log.Printf("K线更新发生异常: %v", err)
		}
	}()

	allCodes := market.GetCodeList()
	var wg sync.WaitGroup

	// 控制并发数量为5（模拟原RollingWaitGroup(5)的效果）
	concurrencyLimit := make(chan struct{}, 5)

	for _, code := range allCodes {
		wg.Add(1)
		concurrencyLimit <- struct{}{} // 占用一个并发名额

		go func(securityCode string) {
			defer wg.Done()
			defer func() { <-concurrencyLimit }() // 释放并发名额

			// 捕获单个goroutine中的panic，避免影响其他任务
			defer func() {
				if err := recover(); err != nil {
					// 可记录单个股票更新失败的日志
					// log.Printf("更新股票 %s K线失败: %v", securityCode, err)
				}
			}()

			snapshot := models.SnapshotMgr.GetTickFromMemory(securityCode)
			if snapshot != nil {
				base.BasicKLineForSnapshot(*snapshot)
			}
		}(code) // 注意此处传参，避免循环变量闭包陷阱
	}

	wg.Wait() // 等待所有任务完成
}
