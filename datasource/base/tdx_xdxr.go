package base

import (
	"gitee.com/quant1x/data/exchange"
	"gitee.com/quant1x/data/level1"
	"gitee.com/quant1x/data/level1/quotes"
	"xquant/cache"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
)

// UpdateXdxrInfo 除权除息数据（优化后：避免重复写入）
//
//	先读取缓存数据，与服务器数据对比，如果数据相同则跳过写入，避免重复IO操作
func UpdateXdxrInfo(securityCode string) {
	securityCode = exchange.CorrectSecurityCode(securityCode)
	
	// 1. 先读取缓存数据
	cachedData := GetCacheXdxrList(securityCode)
	
	// 2. 获取服务器数据
	tdxApi := level1.GetApi()
	serverData, err := tdxApi.GetXdxrInfo(securityCode)
	if err != nil {
		logger.Errorf("获取除权除息数据失败: %v", err)
		return
	}
	
	if len(serverData) == 0 {
		return
	}
	
	// 3. 比较数据是否相同（比较数据条数和关键字段）
	if len(cachedData) == len(serverData) && len(cachedData) > 0 {
		// 比较最后一条数据的关键字段（除权除息数据通常按日期排序）
		// 如果最后一条数据相同，且条数相同，认为数据未变化
		lastCached := cachedData[len(cachedData)-1]
		lastServer := serverData[len(serverData)-1]
		
		// 比较最后一条数据的关键字段
		if lastCached.Date == lastServer.Date &&
			lastCached.HouZongGuBen == lastServer.HouZongGuBen &&
			lastCached.HouLiuTong == lastServer.HouLiuTong {
			// 数据可能未变化，跳过写入
			// 注意：这里只检查最后一条，如果需要更严格的检查，可以比较所有数据
			return
		}
	}
	
	// 4. 数据不同或缓存不存在，写入文件
	filename := cache.XdxrFilename(securityCode)
	_ = api.SlicesToCsv(filename, serverData)
}

// GetCacheXdxrList 获取除权除息的数据列表
func GetCacheXdxrList(securityCode string) []quotes.XdxrInfo {
	securityCode = exchange.CorrectSecurityCode(securityCode)
	filename := cache.XdxrFilename(securityCode)
	var list []quotes.XdxrInfo
	_ = api.CsvToSlices(filename, &list)
	return list
}
