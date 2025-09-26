package trader

import (
	"fmt"
	"testing"
	"xquant/pkg/config"
)

func TestCalculateAvailableFund(t *testing.T) {
	id := 2
	tradeRule := config.GetStrategyParameterByCode(uint64(id))
	if tradeRule == nil {
		return
	}
	fmt.Println(tradeRule)
	fund := CalculateAvailableFund(tradeRule)
	fmt.Println(fund)
}
