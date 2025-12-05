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
	err := models.Register(ModelBlackThree{})
	if err != nil {
		logger.Fatalf("注册黑色三乌鸦策略失败: %+v", err)
	}
}

// ModelBlackThree 5号策略：黑色三乌鸦策略（卖出信号）
//
//	策略逻辑：
//	1. 连续3根阴线（收盘价 < 开盘价）
//	2. 每根K线都是光头光脚（开盘价=最高价，收盘价=最低价）
//	3. 连续下跌趋势（收盘价逐日降低）
type ModelBlackThree struct {
}

func (m ModelBlackThree) Code() models.ModelKind {
	return models.ModelNo5
}

func (m ModelBlackThree) Name() string {
	return "黑色三乌鸦策略"
}

func (m ModelBlackThree) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelBlackThree) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelBlackThree) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelBlackThree) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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
	if df.Nrow() < 3 {
		return
	}

	OPEN := df.ColAsNDArray("open")
	CLOSE := df.ColAsNDArray("close")
	HIGH := df.ColAsNDArray("high")
	LOW := df.ColAsNDArray("low")

	if OPEN.Len() < 3 || CLOSE.Len() < 3 || HIGH.Len() < 3 || LOW.Len() < 3 {
		return
	}

	// 4. 判断是否是黑色三乌鸦形态
	if !m.isBlackThree(OPEN, CLOSE, HIGH, LOW) {
		return
	}

	// 5. 如果满足条件，加入结果（卖出信号）
	price := snapshot.Price
	date := snapshot.Date
	result.Put(securityCode, models.ResultInfo{
		Code:         securityCode,
		Name:         securities.GetStockName(securityCode),
		Date:         date,
		Rate:         0.00,
		Buy:          price,
		Sell:         price * 0.95, // 目标跌幅 5%（卖出信号）
		StrategyCode: m.Code(),
		StrategyName: m.Name(),
	})
}

// isBlackThree 判断是否是黑色三乌鸦形态
// 黑色三乌鸦：连续3根阴线，每根都是光头光脚，连续下跌
func (m ModelBlackThree) isBlackThree(OPEN, CLOSE, HIGH, LOW pandas.Series) bool {
	// 获取最近3根K线的数据（从最新到最旧）
	open0 := utils.Float64IndexOf(OPEN, -1) // 最新
	open1 := utils.Float64IndexOf(OPEN, -2)
	open2 := utils.Float64IndexOf(OPEN, -3) // 最旧

	close0 := utils.Float64IndexOf(CLOSE, -1)
	close1 := utils.Float64IndexOf(CLOSE, -2)
	close2 := utils.Float64IndexOf(CLOSE, -3)

	high0 := utils.Float64IndexOf(HIGH, -1)
	high1 := utils.Float64IndexOf(HIGH, -2)
	high2 := utils.Float64IndexOf(HIGH, -3)

	low0 := utils.Float64IndexOf(LOW, -1)
	low1 := utils.Float64IndexOf(LOW, -2)
	low2 := utils.Float64IndexOf(LOW, -3)

	// 1. 判断最近3根都是阴线（收盘价 < 开盘价）
	if close0 >= open0 || close1 >= open1 || close2 >= open2 {
		return false
	}

	// 2. 判断每根K线都是光头光脚（开盘价=最高价，收盘价=最低价）
	// 允许小的误差（0.01元）
	epsilon := 0.01
	if !(abs(open0-high0) < epsilon && abs(close0-low0) < epsilon) {
		return false
	}
	if !(abs(open1-high1) < epsilon && abs(close1-low1) < epsilon) {
		return false
	}
	if !(abs(open2-high2) < epsilon && abs(close2-low2) < epsilon) {
		return false
	}

	// 3. 判断连续下跌趋势（收盘价逐日降低）
	if close0 >= close1 || close1 >= close2 {
		return false
	}

	return true
}

// abs 返回浮点数的绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
