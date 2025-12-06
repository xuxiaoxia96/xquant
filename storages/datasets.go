package storages

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"xquant/cache"
	"xquant/factors"
	"xquant/market"
	"xquant/pkg/progressbar"

	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/text/runewidth"
)

// calculateConcurrencyLimit 动态计算并发限制
//
//	根据CPU核心数动态计算并发限制，IO密集型任务可以设置更高的并发数
//	返回范围：20-100
func calculateConcurrencyLimit() int {
	cpuCount := runtime.NumCPU()
	limit := cpuCount * 3 // IO密集型，可以设置更高

	// 设置合理范围
	if limit < 20 {
		limit = 20 // 最小20
	}
	if limit > 100 {
		limit = 100 // 最大100，避免资源耗尽
	}
	return limit
}

var (
	// datasetConcurrencyLimit 每个数据集处理股票时的并发数限制（动态计算）
	datasetConcurrencyLimit = calculateConcurrencyLimit()
)

// contextKey 用于 context 的键类型，避免使用 string 作为 key
type contextKey string

const (
	barIndexKey contextKey = "barIndex"
)

// getModuleName 获取模块名称
func getModuleName(op cache.OpKind) string {
	if op == cache.OpRepair {
		return "修复基础数据"
	}
	return "更新基础数据"
}

// catchPanic 捕获 panic 并记录错误信息（替代 runtime.CatchPanic）
func catchPanic(format string, args ...interface{}) {
	if r := recover(); r != nil {
		message := fmt.Sprintf(format, args...)
		logger.Errorf("Panic recovered in %s: %v", message, r)
	}
}

// syncDataSetByDate 同步单个数据集
func syncDataSetByDate(data factors.DataSet, date string, op cache.OpKind) {
	defer catchPanic("%s[%s]: date=%s", data.Name(), data.GetSecurityCode(), date)
	if op == cache.OpUpdate {
		data.Update(date)
	} else if op == cache.OpRepair {
		data.Repair(date)
	}
}

// updateOneDataSet 更新单个数据集（优化后：使用并发限制）
func updateOneDataSet(wg *sync.WaitGroup, parent, bar progressbar.Bar, dataSet factors.DataSet, date string, op cache.OpKind, allCodes []string) {
	moduleName := getModuleName(op)
	logger.Infof("%s: %s, begin", moduleName, dataSet.Name())

	semaphore := make(chan struct{}, datasetConcurrencyLimit)
	var codeWg sync.WaitGroup

	// 并行处理所有股票代码
	for _, code := range allCodes {
		codeWg.Add(1)
		go func(securityCode string) {
			// 获取信号量（限制并发数）
			semaphore <- struct{}{}
			defer func() {
				<-semaphore // 释放信号量
				codeWg.Done()
			}()

			defer catchPanic("%s[%s]: date=%s", dataSet.Name(), securityCode, date)

			data := dataSet.Clone(date, securityCode)
			syncDataSetByDate(data, date, op)
			bar.Add(1)
		}(code)
	}

	codeWg.Wait() // 等待所有股票处理完成
	parent.Add(1)
	wg.Done()
	logger.Infof("%s: %s, end", moduleName, dataSet.Name())
}

// DataSetUpdate 更新或修复基础数据（优化后）
func DataSetUpdate(barIndex int, date string, plugins []cache.DataAdapter, op cache.OpKind) {
	moduleName := getModuleName(op)

	// 1. 预分配切片容量
	dataSetList := make([]factors.DataSet, 0, len(plugins))
	maxWidth := 0

	// 2. 提取 DataSet 并计算最大宽度
	for _, plugin := range plugins {
		dataSet, ok := plugin.(factors.DataSet)
		if !ok {
			continue
		}
		dataSetList = append(dataSetList, dataSet)
		width := runewidth.StringWidth(dataSet.Name())
		if width > maxWidth {
			maxWidth = width
		}
	}

	if len(dataSetList) == 0 {
		logger.Warnf("%s: 没有找到可用的数据集插件", moduleName)
		return
	}

	logger.Infof("%s: all, begin", moduleName)

	// 3. 获取股票列表
	allCodes := market.GetCodeList()
	codeCount := len(allCodes)

	// 4. 创建进度条
	dataSetCount := len(dataSetList)
	barCache := progressbar.NewBar(barIndex, "执行["+date+":"+moduleName+"]", dataSetCount)

	// 5. 准备上下文（移除 coroutine.Context 依赖）
	// 同时设置两个 key 以保持兼容性：类型安全的 key 和 cache.KBarIndex
	ctx := context.WithValue(context.Background(), barIndexKey, barIndex)
	//nolint:SA1029 // 使用 cache.KBarIndex 保持与现有代码的兼容性
	ctx = context.WithValue(ctx, cache.KBarIndex, barIndex)

	// 6. 并行处理所有数据集
	var wg sync.WaitGroup
	bars := make([]progressbar.Bar, len(dataSetList))

	for sequence, dataSet := range dataSetList {
		// 初始化数据集
		_ = dataSet.Init(ctx, date)

		// 构建标题（优化字符串拼接）
		desc := dataSet.Name()
		width := runewidth.StringWidth(desc)
		var titleBuilder strings.Builder
		titleBuilder.Grow(maxWidth)
		titleBuilder.WriteString(strings.Repeat(" ", maxWidth-width))
		titleBuilder.WriteString(desc)
		title := titleBuilder.String()

		// 创建进度条
		barNo := barIndex + 1 + sequence
		barCode := progressbar.NewBar(barNo, "执行["+title+"]", codeCount)
		bars[sequence] = barCode

		// 启动协程处理数据集（始终使用协程）
		wg.Add(1)
		go updateOneDataSet(&wg, barCache, barCode, dataSet, date, op, allCodes)
	}

	// 7. 等待所有数据集处理完成
	wg.Wait()
	barCache.Wait()

	// 8. 等待所有进度条完成
	for _, bar := range bars {
		bar.Wait()
	}

	logger.Infof("%s: all, end", moduleName)
}
