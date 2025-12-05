package strategy

import (
	"gitee.com/quant1x/gotdx/securities"
	"gitee.com/quant1x/gox/concurrent"
	"gitee.com/quant1x/gox/logger"
	. "gitee.com/quant1x/pandas/formula"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/models"
	"xquant/pkg/utils"
)

func init() {
	err := models.Register(ModelBreakthrough{})
	if err != nil {
		logger.Fatalf("注册突破策略失败: %+v", err)
	}
}

// ModelBreakthrough 4号策略：突破策略
//
//	策略逻辑：
//	1. 价格突破近期高点（20日内最高价）
//	2. 成交量放大（成交量 > 5日均量的 1.5 倍）
//	3. 价格在均线上方（Price > MA20）
type ModelBreakthrough struct {
}

func (m ModelBreakthrough) Code() models.ModelKind {
	return models.ModelNo4
}

func (m ModelBreakthrough) Name() string {
	return "突破策略"
}

func (m ModelBreakthrough) OrderFlag() string {
	return models.OrderFlagTick // 盘中实时突破
}

func (m ModelBreakthrough) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return ChainFilters(GeneralFilter)(ruleParameter, snapshot)
}

func (m ModelBreakthrough) Sort(snapshots []factors.QuoteSnapshot) models.SortedStatus {
	return models.SortDefault
}

func (m ModelBreakthrough) Evaluate(securityCode string, result *concurrent.TreeMap[string, models.ResultInfo]) {
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
	if df.Nrow() < 25 {
		return
	}

	HIGH := df.ColAsNDArray("high")
	VOL := df.ColAsNDArray("volume")
	CLOSE := df.ColAsNDArray("close")

	if HIGH.Len() < 25 || VOL.Len() < 25 {
		return
	}

	// 4. 计算近期高点（20日内最高价）
	recentHigh := HHV(HIGH, 20)
	maxHigh := utils.Float64IndexOf(recentHigh, -1)

	// 5. 判断价格是否突破近期高点
	isBreakthrough := snapshot.Price > maxHigh

	// 6. 计算 5 日均量
	avgVol5Series := MA(VOL, 5)
	avgVol5 := utils.Float64IndexOf(avgVol5Series, -1)
	currentVol := float64(snapshot.Vol)

	// 7. 判断成交量是否放大（当前成交量 > 5日均量的 1.5 倍）
	isVolumeAmplified := currentVol > avgVol5*1.5

	// 8. 计算 MA20，判断价格是否在均线上方
	ma20 := utils.Float64IndexOf(MA(CLOSE, 20), -1)
	isPriceAboveMA20 := snapshot.Price > ma20

	// 9. 如果满足所有条件，加入结果
	if isBreakthrough && isVolumeAmplified && isPriceAboveMA20 {
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
}
