package services

import (
	"testing"

	"gitee.com/quant1x/data/level1"
	"xquant/cache"
	"xquant/factors"
)

func TestGlobalReset(t *testing.T) {
	_ = cleanExpiredStateFiles()
	level1.ReOpen()
	date := cache.DefaultCanUpdateDate()
	factors.SwitchDate(date)
}
