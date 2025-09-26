package rules

import (
	"xquant/pkg/config"
	"xquant/pkg/factors"
)

type RuleImpl struct {
	kind Kind
	name string
	exec func(rules config.RuleParameter, snapshot factors.QuoteSnapshot) error
}

func (r RuleImpl) Kind() Kind {
	return r.kind
}

func (r RuleImpl) Name() string {
	return r.name
}

func (r RuleImpl) Exec(rules config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return r.exec(rules, snapshot)
}

func (r RuleImpl) RuleMethod() func(rules config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	return r.exec
}
