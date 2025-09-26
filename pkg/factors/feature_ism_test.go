package factors

import (
	"encoding/json"
	"fmt"
	"gitee.com/quant1x/gox/api"
	"testing"
	"xquant/pkg/cache"
)

func TestISM(t *testing.T) {
	code := "000737"
	date := "2024-03-21"
	cacheDate, featureDate := cache.CorrectDate(date)
	ism := NewInvestmentSentimentMaster(cacheDate, code)
	ism.Update(code, cacheDate, featureDate, true)
	data, _ := json.Marshal(ism)
	text := api.Bytes2String(data)
	fmt.Println(text)
}
