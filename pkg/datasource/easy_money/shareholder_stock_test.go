package easy_money

import (
	"fmt"
	"testing"

	"gitee.com/quant1x/exchange"
)

func TestShareHolder(t *testing.T) {
	code := "sh600115"
	v := ShareHolder(code, exchange.Today(), 2)
	fmt.Println(v)
}

func TestGetCacheShareHolder(t *testing.T) {
	code := "sh600105"
	v := GetCacheShareHolder(code, exchange.Today(), 4)
	fmt.Println(v)
}
