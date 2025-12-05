package tracker

import (
	"fmt"

	"xquant/config"
	"xquant/models"
	"xquant/permissions"
	"gitee.com/quant1x/gox/logger"
)

// ExecuteStrategy 执行策略
func ExecuteStrategy(model models.Strategy, barIndex *int) {
	// 策略权限验证
	err := permissions.CheckPermission(model)
	if err != nil {
		logger.Error(err)
		fmt.Println(err)
		return
	}
	tradeRule := config.GetStrategyParameterByCode(model.Code())
	if tradeRule == nil {
		fmt.Printf("strategy[%d]: trade rule not found\n", model.Code())
		return
	}
	// 加载快照数据
	models.SyncAllSnapshots(barIndex)
	// 计算市场情绪
	MarketSentiment()
	// 扫描板块
	ScanAllSectors(barIndex, model)
	// 扫描个股
	AllScan(barIndex, model)
}
