package strategy

import (
	"gitee.com/quant1x/gox/concurrent"
	"quantity/common"
	"time"
)

// AvgPriceDownStrategy   代表7日均线向下穿过25日均线和99日均线
type AvgPriceDownStrategy struct {
	name string
}

func (down *AvgPriceDownStrategy) Code() ModelKind {
	//TODO implement me
	panic("implement me")
}

func (down *AvgPriceDownStrategy) OrderFlag() string {
	//TODO implement me
	panic("implement me")
}

func (down *AvgPriceDownStrategy) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	//TODO implement me
	panic("implement me")
}

func (down *AvgPriceDownStrategy) Sort(snapshots []factors.QuoteSnapshot) SortedStatus {
	//TODO implement me
	panic("implement me")
}

func (down *AvgPriceDownStrategy) Evaluate(securityCode string, result *concurrent.TreeMap[string, ResultInfo]) {
	//TODO implement me
	panic("implement me")
}

func (down *AvgPriceDownStrategy) Name() string {
	return down.name
}

func NewAvgPriceDownStrategy() Strategy {
	return &AvgPriceDownStrategy{
		name: "down",
	}
}

func (down *AvgPriceDownStrategy) Analysis(symbol string, kLines []*common.KLine) (action *common.SubmitOrder, err error) {
	// 获取当前价格
	now := time.Now()
	price, err := getCurrentPrice()
	if err != nil {
		return
	}

	currentPrice := price[symbol]
	action = &common.SubmitOrder{
		Symbol:       symbol,
		Price:        currentPrice,
		Action:       common.Hold,
		Timestamp:    now,
		StrategyName: down.name,
	}

	//sum, e := common.SymbolOrderSumAction(symbol)
	//if e != nil {
	//	return nil, e
	//}

	// 当前价格下跌,7日均线价格小于25日均线价格,即刻止损卖出
	//if sum >= 1 {
	//action.Action = common.Sell
	//}

	return
}
