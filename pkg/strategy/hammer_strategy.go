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
	err := models.Register(ModelHammer{})
	if err != nil {
		logger.Fatalf("注册锤子线策略失败: %+v", err)
	}
}

// ModelHammer 6号策略：锤子线策略（买入信号）
//
//	策略逻辑：
//	1. 最后一根K线是上涨的（收盘价 > 开盘价）
//	2. 倒数第二根K线是锤子线形态（下影线很长，实体很小）
//	3. 之前是下跌趋势（收盘价逐日降低）
type ModelHammer struct {
}

func (m ModelHammer) Code() models.ModelKind {
	return models.ModelNo6
}

func (m ModelHammer) Name() string {
	return "锤子线策略"
}

func (m ModelHammer) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelHammer) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelHammer) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelHammer) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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
	if df.Nrow() < 5 {
		return
	}

	OPEN := df.ColAsNDArray("open")
	CLOSE := df.ColAsNDArray("close")
	HIGH := df.ColAsNDArray("high")
	LOW := df.ColAsNDArray("low")

	if OPEN.Len() < 5 || CLOSE.Len() < 5 || HIGH.Len() < 5 || LOW.Len() < 5 {
		return
	}

	// 4. 判断是否满足锤子线形态
	if !m.isHammerPattern(OPEN, CLOSE, HIGH, LOW) {
		return
	}

	// 5. 如果满足条件，加入结果（买入信号）
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

// isHammerPattern 判断是否满足锤子线形态
// 锤子线：最后一根上涨，倒数第二根是锤子线，之前是下跌趋势
func (m ModelHammer) isHammerPattern(OPEN, CLOSE, HIGH, LOW pandas.Series) bool {
	// 获取最近几根K线的数据
	open0 := utils.Float64IndexOf(OPEN, -1) // 最新（最后一根）
	open1 := utils.Float64IndexOf(OPEN, -2) // 倒数第二根（锤子线）
	open2 := utils.Float64IndexOf(OPEN, -3)

	close0 := utils.Float64IndexOf(CLOSE, -1)
	close1 := utils.Float64IndexOf(CLOSE, -2)
	close2 := utils.Float64IndexOf(CLOSE, -3)

	high0 := utils.Float64IndexOf(HIGH, -1)
	high1 := utils.Float64IndexOf(HIGH, -2)
	high2 := utils.Float64IndexOf(HIGH, -3)

	low0 := utils.Float64IndexOf(LOW, -1)
	low1 := utils.Float64IndexOf(LOW, -2)
	low2 := utils.Float64IndexOf(LOW, -3)

	// 1. 判断最后一根K线是上涨的（收盘价 > 开盘价）
	if close0 <= open0 {
		return false
	}

	// 2. 判断倒数第二根K线是锤子线形态
	// 锤子线特征：
	// - 下影线长度 > 实体长度的2倍
	// - 上影线很短或没有（上影线 < 实体长度）
	// - 实体较小（实体 < K线总高度的1/3）
	body1 := abs(close1 - open1)               // 实体长度
	upperShadow1 := high1 - max(open1, close1) // 上影线长度
	lowerShadow1 := min(open1, close1) - low1  // 下影线长度
	totalHeight1 := high1 - low1               // K线总高度

	if totalHeight1 <= 0 {
		return false
	}

	// 下影线长度 > 实体长度的2倍
	if lowerShadow1 <= body1*2 {
		return false
	}

	// 上影线很短（上影线 < 实体长度）
	if upperShadow1 >= body1 {
		return false
	}

	// 实体较小（实体 < K线总高度的1/3）
	if body1 >= totalHeight1/3 {
		return false
	}

	// 3. 判断之前是下跌趋势（收盘价逐日降低）
	// 从倒数第三根开始，检查是否连续下跌
	if close1 >= close2 {
		return false
	}

	// 可以检查更多根K线，确保是下跌趋势
	if OPEN.Len() >= 5 {
		close3 := utils.Float64IndexOf(CLOSE, -4)
		if close2 >= close3 {
			return false
		}
	}

	return true
}

// max 返回两个浮点数的最大值
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min 返回两个浮点数的最小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
