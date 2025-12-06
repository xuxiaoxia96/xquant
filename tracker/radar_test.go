package tracker

import (
	"fmt"
	"testing"

	"xquant/config"
	"xquant/models"
)

func TestConfig(t *testing.T) {
	strategyCode := 82
	rule := config.GetStrategyParameterByCode(uint64(strategyCode))
	fmt.Println(rule)
	list := rule.StockList()
	fmt.Println(list)
}

// TestRadarScanSectorForTick 测试市场雷达扫描板块功能
// 这是一个示例，展示如何使用市场雷达功能
func TestRadarScanSectorForTick(t *testing.T) {
	// 注意：这个测试需要实时数据，在非交易时段可能无法正常运行
	// 可以通过设置 XQUANT_DEBUG=true 环境变量来在非交易时段运行

	// 初始化进度条索引
	barIndex := 1

	// 确保快照数据已同步
	models.SyncAllSnapshots(&barIndex)

	// 调用市场雷达扫描板块
	// 这会扫描所有概念板块，筛选出表现好的板块，并提取其中的个股
	stockCodes := ScanSectorForTick(&barIndex)

	// 输出结果
	fmt.Printf("\n市场雷达扫描完成，发现 %d 只股票\n", len(stockCodes))
	if len(stockCodes) > 0 {
		fmt.Printf("前10只股票代码: %v\n", stockCodes[:min(10, len(stockCodes))])
	}

	// 验证：至少应该返回一个空切片（即使没有符合条件的股票）
	if stockCodes == nil {
		t.Error("市场雷达应该返回一个切片，即使是空的")
	}
}

// min 辅助函数，返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
