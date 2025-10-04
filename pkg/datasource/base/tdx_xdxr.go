package base

import (
	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx"
	"gitee.com/quant1x/gotdx/quotes"
	"gitee.com/quant1x/gox/api"

	"xquant/pkg/cache"
	"xquant/pkg/log"
)

// UpdateXdxrInfo 除权除息数据
func UpdateXdxrInfo(securityCode string) {
	securityCode = exchange.CorrectSecurityCode(securityCode)
	xdxrInfos, err := gotdx.GetTdxApi().GetXdxrInfo(securityCode)
	if err != nil {
		log.Errorf("获取除权除息数据失败 %s", err)
		return
	}

	if len(xdxrInfos) > 0 {
		filename := cache.XdxrFilename(securityCode)
		_ = api.SlicesToCsv(filename, xdxrInfos)
	}
}

// GetCacheXdxrList 获取除权除息的数据列表
func GetCacheXdxrList(securityCode string) []quotes.XdxrInfo {
	filename := cache.XdxrFilename(exchange.CorrectSecurityCode(securityCode))
	var list []quotes.XdxrInfo
	_ = api.CsvToSlices(filename, &list)
	return list
}
