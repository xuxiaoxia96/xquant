package base

import (
	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx"
	"gitee.com/quant1x/gotdx/quotes"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/runtime"
	"xquant/pkg/cache"
)

// GetMinutes 获取分时数据
func GetMinutes(securityCode, date string) (list []quotes.MinuteTime) {
	tdxApi := gotdx.GetTdxApi()
	u32Date := exchange.ToUint32Date(date)
	hs, err := tdxApi.GetHistoryMinuteTimeData(securityCode, u32Date)
	if err != nil || hs.Count == 0 {
		return
	}
	list = append(list, hs.List...)
	_ = hs
	return
}

// UpdateMinutes 更新指定日期的个股分时数据
func UpdateMinutes(securityCode, date string) {
	defer runtime.IgnorePanic("update-minutes: code=%s, date=%s", securityCode, date)
	list := GetMinutes(securityCode, date)
	if len(list) > 0 {
		filename := cache.MinuteFilename(securityCode, date)
		_ = api.SlicesToCsv(filename, list)
	}
}
