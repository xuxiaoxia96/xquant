package tracker

import (
	"fmt"
	"testing"

	"xquant/pkg/config"
)

func TestConfig(t *testing.T) {
	strategyCode := 82
	rule := config.GetStrategyParameterByCode(uint64(strategyCode))
	fmt.Println(rule)
	list := rule.StockList()
	fmt.Println(list)
}
