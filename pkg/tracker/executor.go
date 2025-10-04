package tracker

import (
	"context"
	"fmt"
	"xquant/pkg/utils"

	"xquant/pkg/config"
	"xquant/pkg/models"
)

// ExecuteStrategy 执行策略
func ExecuteStrategy(model models.Strategy, barIndex *int) {
	tradeRule := config.GetStrategyParameterByCode(model.Code())
	if tradeRule == nil {
		fmt.Printf("strategy[%d]: trade rule not found\n", model.Code())
		return
	}
	// 加载快照数据
	models.SnapshotMgr.SyncAllSnapshots(context.Background(), utils.Ptr(1))
	// 计算市场情绪
	MarketSentiment()
	// 扫描板块
	ScanAllSectors(barIndex, model)
	// 扫描个股
	AllScan(barIndex, model)
}
