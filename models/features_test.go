package models

import (
	"fmt"
	"testing"

	"gitee.com/quant1x/data/exchange"
	"xquant/cache"
	"xquant/factors"
	"gitee.com/quant1x/gox/api"
)

func TestFeatureToSnapshot(t *testing.T) {
	code := "300410"
	securityCode := exchange.CorrectSecurityCode(code)
	filename := cache.WideFilename(securityCode)
	features := []factors.SecurityFeature{}
	err := api.CsvToSlices(filename, &features)
	if err != nil {
		fmt.Println(err)
		return
	}
	length := len(features)
	feature := features[length-1]
	v := FeatureToSnapshot(feature, securityCode)
	fmt.Println(v)
}
