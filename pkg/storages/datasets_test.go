package storages

import (
	"testing"
	"xquant/pkg/cache"
)

func TestBaseDataUpdate(t *testing.T) {
	barIndex := 1
	date := "2024-01-31"
	plugins := cache.PluginsWithName(cache.PluginMaskBaseData, "wide")
	BaseDataUpdate(barIndex, date, plugins, cache.OpUpdate)

}
