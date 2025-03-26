package factors

import (
	"fmt"
	"gitee.com/quant1x/engine/datasource/base"
	"testing"
)

func TestTransactionOld(t *testing.T) {
	code := "sh881200"
	date := "2025-02-21"
	list := base.GetHistoricalTradingData(code, date)
	v := CountInflow(list, code, date)
	fmt.Printf("%+v\n", v)
}

func TestTransaction(t *testing.T) {
	code := "sh881200"
	date := "2025-02-21"
	list := base.CheckoutTransactionData(code, date, true)
	v := CountInflow(list, code, date)
	fmt.Printf("%+v\n", v)
}
