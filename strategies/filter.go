package strategies

import (
	"xquant/config"
	"xquant/factors"
	"xquant/rules"
)

// GeneralFilter 过滤条件
//
//	执行所有在册的规则
func GeneralFilter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	passed, failed, err := rules.Filter(ruleParameter, snapshot)
	_ = passed
	_ = failed
	return err
}
