package storages

import (
	"path/filepath"

	"xquant/cache"
	"xquant/config"
)

var (
	traderConfig = config.TraderConfig()
)

const (
	StrategiesPath = "quant" // 策略结果数据文件存储路径
)

// GetResultCachePath 获取结果缓存路径
func GetResultCachePath() string {
	path := filepath.Join(cache.GetRootPath(), StrategiesPath)
	return path
}
