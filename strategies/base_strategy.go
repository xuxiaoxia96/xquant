package strategies

import (
	"xquant/config"
	"xquant/factors"
	"xquant/rules"
)

// BaseStrategy 基础策略实现，提供默认的 Filter 和 Sort 方法
// 具体策略可以嵌入此结构体以复用默认实现
type BaseStrategy struct{}

// Filter 默认过滤实现，使用通用过滤规则
func (b BaseStrategy) Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return rules.GeneralFilter(ruleParameter, snapshot)
}

// Sort 默认排序实现，返回默认排序状态
func (b BaseStrategy) Sort(snapshots []factors.QuoteSnapshot) SortedStatus {
	return SortDefault
}

