package storages

import (
	"context"
	"strings"
	"sync"
	"xquant/pkg/utils"

	"gitee.com/quant1x/gox/progressbar"
	"gitee.com/quant1x/gox/text/runewidth"

	"xquant/pkg/cache"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/market"
)

// DataSetUpdate 修复数据
func DataSetUpdate(barIndex int, date string, plugins []cache.DataAdapter, op cache.OpKind) {
	moduleName := "基础数据"
	if op == cache.OpRepair {
		moduleName = "修复" + moduleName
	} else {
		moduleName = "更新" + moduleName
	}
	var dataSetList []factors.DataSet
	// 1.1 缓存数据集名称的最大宽度
	maxWidth := 0
	for _, plugin := range plugins {
		dataSet, ok := plugin.(factors.DataSet)
		if ok {
			dataSetList = append(dataSetList, dataSet)
			width := runewidth.StringWidth(dataSet.Name())
			if width > maxWidth {
				maxWidth = width
			}
		}
	}
	log.Infof("%s: all, begin", moduleName)
	// 2. 遍历全部数据插件
	dataSetCount := len(dataSetList)
	barCache := progressbar.NewBar(barIndex, "执行["+date+":"+moduleName+"]", dataSetCount)

	allCodes := market.GetCodeList()
	codeCount := len(allCodes)
	var wg sync.WaitGroup

	parent := context.Background()
	ctx := context.WithValue(parent, cache.KBarIndex, barIndex)
	for sequence, dataSet := range dataSetList {
		_ = dataSet.Init(ctx, date)
		desc := dataSet.Name()
		width := runewidth.StringWidth(desc)
		title := strings.Repeat(" ", maxWidth-width) + desc
		barNo := barIndex + 1
		if cache.UseGoroutine {
			barNo += sequence
		}
		barCode := progressbar.NewBar(barNo, "执行["+title+"]", codeCount)
		wg.Add(1)
		if cache.UseGoroutine {
			go updateOneDataSet(&wg, barCache, barCode, dataSet, date, op, allCodes)
		} else {
			updateOneDataSet(&wg, barCache, barCode, dataSet, date, op, allCodes)
		}
		barCode.Wait()
	}
	barCache.Wait()
	wg.Wait()
	log.Infof("%s: all, end", moduleName)
}

func syncDataSetByDate(data factors.DataSet, date string, operation cache.OpKind) {
	defer utils.CatchPanic("%s[%s]: date=%s", data.Name(), data.GetSecurityCode(), date)
	if operation == cache.OpUpdate {
		data.Update(date)
	} else if operation == cache.OpRepair {
		data.Repair(date)
	}
}

// 更新单个数据集
func updateOneDataSet(wg *sync.WaitGroup, parent, bar *progressbar.Bar, dataSet factors.DataSet, date string, operation cache.OpKind, allCodes []string) {
	moduleName := "基础数据"
	if operation == cache.OpRepair {
		moduleName = "修复" + moduleName
	} else {
		moduleName = "更新" + moduleName
	}
	log.Infof("%s: %s, begin", moduleName, dataSet.Name())

	for _, code := range allCodes {
		data := dataSet.Clone(date, code).(factors.DataSet)
		syncDataSetByDate(data, date, operation)
		bar.Add(1)
	}

	parent.Add(1)
	wg.Done()
	log.Infof("%s: %s, end", moduleName, dataSet.Name())
}
