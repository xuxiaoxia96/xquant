package trader

import (
	"xquant/config"
	"xquant/factors"
	"xquant/strategies"

	"gitee.com/quant1x/gox/concurrent"
)

type TestModel struct{}

func (TestModel) Code() strategies.ModelKind {
	return 82
}

func (s TestModel) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s TestModel) OrderFlag() string {
	//TODO implement me
	panic("implement me")
}

func (s TestModel) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	//TODO implement me
	panic("implement me")
}

func (s TestModel) Sort(snapshots []factors.QuoteSnapshot) strategies.SortedStatus {
	//TODO implement me
	panic("implement me")
}

func (s TestModel) Evaluate(securityCode string, result *concurrent.TreeMap[string, strategies.Signal]) {
	//TODO implement me
	panic("implement me")
}
