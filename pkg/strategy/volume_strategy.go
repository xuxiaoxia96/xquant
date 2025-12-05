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
	err := models.Register(ModelVolume{})
	if err != nil {
		logger.Fatalf("注册放量上涨策略失败: %+v", err)
	}
}

// ModelVolume 8号策略：放量上涨策略（买入信号）
//
//	策略逻辑：
//	1. 之前是下跌趋势（收盘价逐日降低）
//	2. 最近一根K线是上涨的（收盘价 > 开盘价）
//	3. 成交量异常放大（当前成交量 > 之前平均成交量的5倍）
type ModelVolume struct {
}

func (m ModelVolume) Code() models.ModelKind {
	return models.ModelNo8
}

func (m ModelVolume) Name() string {
	return "放量上涨策略"
}

func (m ModelVolume) OrderFlag() string {
	return models.OrderFlagTail
}

func (m ModelVolume) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelVolume) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelVolume) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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
	if df.Nrow() < 10 {
		return
	}

	OPEN := df.ColAsNDArray("open")
	CLOSE := df.ColAsNDArray("close")
	VOL := df.ColAsNDArray("volume")

	if OPEN.Len() < 10 || CLOSE.Len() < 10 || VOL.Len() < 10 {
		return
	}

	// 4. 判断是否满足放量上涨形态
	if !m.isVolumePattern(OPEN, CLOSE, VOL) {
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
		Sell:         price * 1.12, // 目标涨幅 12%
		StrategyCode: m.Code(),
		StrategyName: m.Name(),
	})
}

// isVolumePattern 判断是否满足放量上涨形态
// 放量上涨：之前下跌，最近上涨，成交量放大5倍以上
func (m ModelVolume) isVolumePattern(OPEN, CLOSE, VOL pandas.Series) bool {
	// 获取最近几根K线的数据
	open0 := utils.Float64IndexOf(OPEN, -1) // 最新（最后一根）
	close0 := utils.Float64IndexOf(CLOSE, -1)
	close1 := utils.Float64IndexOf(CLOSE, -2)
	close2 := utils.Float64IndexOf(CLOSE, -3)

	vol0 := utils.Float64IndexOf(VOL, -1) // 最新成交量

	// 1. 判断最近一根K线是上涨的（收盘价 > 开盘价）
	if close0 <= open0 {
		return false
	}

	// 2. 判断之前是下跌趋势（收盘价逐日降低）
	// 检查最近3-5根K线是否连续下跌
	downCount := 0
	for i := 2; i <= 5 && i < CLOSE.Len(); i++ {
		prevClose := utils.Float64IndexOf(CLOSE, -i)
		currClose := utils.Float64IndexOf(CLOSE, -(i - 1))
		if currClose < prevClose {
			downCount++
		}
	}
	// 至少要有2根K线是下跌的
	if downCount < 2 {
		return false
	}

	// 3. 判断成交量异常放大
	// 计算之前5根K线的平均成交量（不包括最新一根）
	if VOL.Len() < 6 {
		return false
	}

	var sumVolume float64
	count := 0
	for i := 2; i <= 6 && i < VOL.Len(); i++ {
		vol := utils.Float64IndexOf(VOL, -i)
		sumVolume += vol
		count++
	}

	if count == 0 {
		return false
	}

	avgVolume := sumVolume / float64(count)

	// 当前成交量 > 平均成交量的5倍
	if avgVolume <= 0 || vol0/avgVolume < 5.0 {
		return false
	}

	return true
}
