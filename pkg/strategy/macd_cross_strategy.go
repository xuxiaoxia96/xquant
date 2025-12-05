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
)

func init() {
	err := models.Register(ModelMacdCross{})
	if err != nil {
		logger.Fatalf("注册MACD金叉策略失败: %+v", err)
	}
}

// ModelMacdCross 2号策略：MACD 金叉策略
//
//	策略逻辑：
//	1. MACD 金叉（DIF 上穿 DEA）
//	2. MACD 柱状图转正（MACD > 0）
//	3. 价格在均线上方（Price > MA20）
type ModelMacdCross struct {
}

func (m ModelMacdCross) Code() models.ModelKind {
	return models.ModelNo2
}

func (m ModelMacdCross) Name() string {
	return "MACD金叉策略"
}

func (m ModelMacdCross) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelMacdCross) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelMacdCross) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelMacdCross) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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

	// 3. 获取 K 线数据计算 MACD
	df := factors.BasicKLine(securityCode)
	if df.Nrow() < 30 {
		return
	}

	CLOSE := df.ColAsNDArray("close")
	if CLOSE.Len() < 30 {
		return
	}

	// 4. 计算 MACD（12, 26, 9）
	_, _, dif, dea, macd := realtime.MovingAverageConvergenceDivergence(CLOSE, 12, 26, 9)

	// 5. 获取前一日 MACD 值（用于判断金叉）
	// 使用 REF 函数获取前一日收盘价序列
	prevCLOSE := REF(CLOSE, 1)
	if prevCLOSE.Len() < 30 {
		return
	}
	// 计算前一日MACD
	_, _, prevDIF, prevDEA, _ := realtime.MovingAverageConvergenceDivergence(prevCLOSE, 12, 26, 9)

	// 6. 判断 MACD 金叉：DIF 上穿 DEA
	isGoldenCross := prevDIF <= prevDEA && dif > dea

	// 7. 判断 MACD 柱状图转正：MACD > 0
	isMacdPositive := macd > 0

	// 8. 判断价格在均线上方：Price > MA20
	ma20 := realtime.IncrementalMovingAverage(history.MA19, 20, snapshot.Price)
	isPriceAboveMA20 := snapshot.Price > ma20

	// 9. 如果满足所有条件，加入结果
	if isGoldenCross && isMacdPositive && isPriceAboveMA20 {
		price := snapshot.Price
		date := snapshot.Date
		result.Put(securityCode, models.ResultInfo{
			Code:         securityCode,
			Name:         securities.GetStockName(securityCode),
			Date:         date,
			Rate:         0.00,
			Buy:          price,
			Sell:         price * 1.08, // 目标涨幅 8%
			StrategyCode: m.Code(),
			StrategyName: m.Name(),
		})
	}
}
