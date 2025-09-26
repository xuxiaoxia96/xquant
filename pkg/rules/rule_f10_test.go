package rules

import (
	"fmt"
	"testing"

	"xquant/pkg/config"
	"xquant/pkg/models"
)

func Test_baseFilter(t *testing.T) {
	code := "601868"
	code = "sh600622"
	code = "sh601188"
	code = "sz002682"
	snapshot := models.GetTick(code)
	strategyParameter := config.GetStrategyParameterByCode(0)
	passed, failKind, err := Filter(strategyParameter.Rules, *snapshot)
	fmt.Println(passed, failKind, err)
}
