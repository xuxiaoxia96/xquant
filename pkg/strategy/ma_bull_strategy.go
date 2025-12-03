package strategy

import (
	"gitee.com/quant1x/gotdx/securities"
	"gitee.com/quant1x/gox/concurrent"
	"gitee.com/quant1x/gox/logger"
	. "gitee.com/quant1x/pandas/formula"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/models"
	"xquant/pkg/realtime"
	"xquant/pkg/utils"
)

func init() {
	err := models.Register(ModelMABull{})
	if err != nil {
		logger.Fatalf("注册均线多头排列策略失败: %+v", err)
	}
}

// ModelMABull 3号策略：均线多头排列策略
//
//	策略逻辑：
//	1. 均线多头排列（MA5 > MA10 > MA20）
//	2. 价格在均线上方（Price > MA5）
//	3. 均线向上发散（MA5 持续上升）
type ModelMABull struct {
}

func (m ModelMABull) Code() models.ModelKind {
	return models.ModelNo3
}

func (m ModelMABull) Name() string {
	return "均线多头排列策略"
}

func (m ModelMABull) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelMABull) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return GeneralFilter(ruleParameter, snapshot)
}

func (m ModelMABull) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelMABull) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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

	// 3. 计算增量均线
	ma5 := realtime.IncrementalMovingAverage(history.MA4, 5, snapshot.Price)
	ma10 := realtime.IncrementalMovingAverage(history.MA9, 10, snapshot.Price)
	ma20 := realtime.IncrementalMovingAverage(history.MA19, 20, snapshot.Price)

	// 4. 判断均线多头排列：MA5 > MA10 > MA20
	isBullAlignment := ma5 > ma10 && ma10 > ma20

	// 5. 判断价格在均线上方：Price > MA5
	isPriceAboveMA5 := snapshot.Price > ma5

	// 6. 获取 K 线数据，判断均线是否向上发散
	df := factors.BasicKLine(securityCode)
	if df.Nrow() < 6 {
		return
	}

	CLOSE := df.ColAsNDArray("close")
	if CLOSE.Len() < 6 {
		return
	}

	// 7. 计算前一日均线，判断是否上升趋势
	prevCLOSE := REF(CLOSE, 1)
	if prevCLOSE.Len() < 5 {
		return
	}
	prevMA5 := MA(prevCLOSE, 5)
	prevMA5Value := utils.Float64IndexOf(prevMA5, -1)

	// 8. 判断 MA5 是否上升
	isMA5Rising := ma5 > prevMA5Value

	// 9. 如果满足所有条件，加入结果
	if isBullAlignment && isPriceAboveMA5 && isMA5Rising {
		price := snapshot.Price
		date := snapshot.Date
		result.Put(securityCode, models.ResultInfo{
			Code:         securityCode,
			Name:         securities.GetStockName(securityCode),
			Date:         date,
			Rate:         0.00,
			Buy:          price,
			Sell:         price * 1.10, // 目标涨幅 10%
			StrategyCode: m.Code(),
			StrategyName: m.Name(),
		})
	}
}

