package easy_money

import (
	"fmt"
	"testing"

	"gitee.com/quant1x/pandas"
)

func TestIndividualStocksFundFlow(t *testing.T) {
	code := "sz000701"
	list := IndividualStocksFundFlow(code, "2025-03-12")
	df := pandas.LoadStructs(list)
	fmt.Println(df)
}
