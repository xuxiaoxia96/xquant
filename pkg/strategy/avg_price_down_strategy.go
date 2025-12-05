package strategy

import (
	"gitee.com/quant1x/gotdx/securities"
	"gitee.com/quant1x/gox/concurrent"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/pandas"
	. "gitee.com/quant1x/pandas/formula"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/models"
	"xquant/pkg/utils"
)

func init() {
	err := models.Register(ModelAvgPriceDown{})
	if err != nil {
		logger.Fatalf("注册均线向下策略失败: %+v", err)
	}
}

// ModelAvgPriceDown 9号策略：均线向下策略（卖出信号）
//
//	策略逻辑：
//	1. 7日均线向下穿过25日均线
//	2. 7日均线向下穿过99日均线
//	3. 这是一个卖出信号，表示趋势转弱
type ModelAvgPriceDown struct {
}

func (m ModelAvgPriceDown) Code() models.ModelKind {
	return models.ModelNo9
}

func (m ModelAvgPriceDown) Name() string {
	return "均线向下策略"
}

func (m ModelAvgPriceDown) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelAvgPriceDown) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelAvgPriceDown) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelAvgPriceDown) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
	// 1. 获取历史数据
	history := factors.GetL5History(securityCode)
	if history == nil {
		return
	}

	// 2. 获取策略快照
	snapshot := models.SnapshotMgr.GetStrategySnapshot(securityCode)
	if snapshot == nil {
		return
	}

	// 3. 获取 K 线数据
	df := factors.BasicKLine(securityCode)
	if df.Nrow() < 100 {
		return
	}

	CLOSE := df.ColAsNDArray("close")
	if CLOSE.Len() < 100 {
		return
	}

	// 4. 计算均线
	// 7日均线
	ma7 := MA(CLOSE, 7)
	// 25日均线
	ma25 := MA(CLOSE, 25)
	// 99日均线
	ma99 := MA(CLOSE, 99)

	if ma7.Len() < 2 || ma25.Len() < 2 || ma99.Len() < 2 {
		return
	}

	// 5. 判断是否满足均线向下形态
	if !m.isAvgPriceDown(ma7, ma25, ma99) {
		return
	}

	// 6. 如果满足条件，加入结果（卖出信号）
	price := snapshot.Price
	date := snapshot.Date
	result.Put(securityCode, models.ResultInfo{
		Code:         securityCode,
		Name:         securities.GetStockName(securityCode),
		Date:         date,
		Rate:         0.00,
		Buy:          price,
		Sell:         price * 0.92, // 目标跌幅 8%（卖出信号）
		StrategyCode: m.Code(),
		StrategyName: m.Name(),
	})
}

// isAvgPriceDown 判断是否满足均线向下形态
// 7日均线向下穿过25日均线和99日均线
func (m ModelAvgPriceDown) isAvgPriceDown(ma7, ma25, ma99 pandas.Series) bool {
	// 获取当前和前一日均线值
	ma7Current := utils.Float64IndexOf(ma7, -1)
	ma7Prev := utils.Float64IndexOf(ma7, -2)

	ma25Current := utils.Float64IndexOf(ma25, -1)
	ma25Prev := utils.Float64IndexOf(ma25, -2)

	ma99Current := utils.Float64IndexOf(ma99, -1)
	ma99Prev := utils.Float64IndexOf(ma99, -2)

	// 1. 判断7日均线向下穿过25日均线
	// 前一日：MA7 >= MA25
	// 当前：MA7 < MA25
	crossDown25 := ma7Prev >= ma25Prev && ma7Current < ma25Current

	// 2. 判断7日均线向下穿过99日均线
	// 前一日：MA7 >= MA99
	// 当前：MA7 < MA99
	crossDown99 := ma7Prev >= ma99Prev && ma7Current < ma99Current

	// 3. 两个条件都满足，表示均线向下
	return crossDown25 && crossDown99
}
