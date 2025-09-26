package tracker

import (
	"context"
	"sort"
	"sync"
	"time"

	"gitee.com/quant1x/exchange"
	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/exp/slices"

	"xquant/biz/handler"
	trackermodel "xquant/biz/model/tracker"
	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/models"
	"xquant/pkg/openapi_error"
	"xquant/pkg/tracker"
)

// Tracker 实时跟踪策略在当前市场的表现，输出表格
func Tracker(ctx context.Context, c *app.RequestContext) {
	// 解析并验证请求参数
	var req trackermodel.TrackerRequest
	if err := c.BindAndValidate(&req); err != nil {
		log.CtxErrorf(ctx, "[Tracker] 参数绑定失败: %s", err)
		handler.OpenAPIFail(ctx, c, openapi_error.NewInvalidParameterError(ctx, "", err.Error()))
		return
	}

	trackerStrategyCodes := req.GetTrackerStrategyCodes()
	if len(trackerStrategyCodes) == 0 {
		log.CtxWarnf(ctx, "[Tracker] 未指定跟踪的策略代码")
		return // 无跟踪目标，直接返回
	}

	// 使用带退出条件的循环，避免无法终止的无限循环
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop() // 确保资源释放

	for {
		select {
		case <-ctx.Done(): // 上下文取消时退出
			log.CtxWarnf(ctx, "[Tracker] 收到退出信号，停止跟踪")
			return
		case <-ticker.C: // 按固定间隔执行
			// 检查是否可以实时更新
			updateInRealTime, status := exchange.CanUpdateInRealtime()

			isTrading := updateInRealTime &&
				(status == exchange.ExchangeTrading || status == exchange.ExchangeSuspend)
			// 非交易时段且非调试模式，退出跟踪
			if !isTrading && !req.IsDebug {
				log.CtxInfof(ctx, "[Tracker] 非交易时段，停止跟踪")
				return
			}

			// 休市暂停状态，跳过本次循环
			if status == exchange.ExchangeSuspend {
				continue
			}

			barIndex := 1
			models.SnapshotMgr.SyncAllSnapshots(ctx, &barIndex)

			// 并行处理多个策略（提高效率，根据实际情况调整并行度）
			var wg sync.WaitGroup
			for _, strategyCode := range trackerStrategyCodes {
				wg.Add(1)
				go func(code uint64) {
					defer wg.Done()
					processStrategy(ctx, code, req.IsDebug)
				}(strategyCode)
			}
			wg.Wait()
		}
	}
}

// processStrategy 处理单个策略的跟踪逻辑
func processStrategy(ctx context.Context, strategyCode uint64, isDebug bool) {
	// 获取策略对象
	strategy, err := models.CheckoutStrategy(strategyCode)
	if err != nil {
		log.CtxErrorf(ctx, "[Tracker] 获取策略 %d 失败: %s", strategyCode, err)
		return
	}
	if strategy == nil {
		log.CtxWarnf(ctx, "[Tracker] 策略 %d 不存在", strategyCode)
		return
	}

	// 获取策略参数
	strategyParam := config.GetStrategyParameterByCode(strategyCode)
	if strategyParam == nil {
		log.CtxWarnf(ctx, "[Tracker] 策略 %d 无参数配置", strategyCode)
		return
	}

	// 检查是否在交易时段或调试模式
	if strategyParam.Session.IsTrading() || isDebug {
		snapshotTracker(strategy, strategyParam)
	}
}

// SnapshotTracker 策略配置 quant-engine/engine/config/resources/quant1x.yaml
func snapshotTracker(model models.Strategy, tradeRule *config.StrategyParameter) {
	if tradeRule == nil {
		return
	}

	// 获取并验证股票代码
	stockCodes := getValidStockCodes(tradeRule)
	if len(stockCodes) == 0 {
		return
	}

	// 获取股票快照
	stockSnapshots := getStockSnapshots(stockCodes)
	if len(stockSnapshots) == 0 {
		return
	}

	// 过滤股票
	filteredSnapshots := filterStocks(model, tradeRule, stockSnapshots)

	// 排序股票
	sortedSnapshots := sortStocks(model, filteredSnapshots)

	// “数据转换→表格输出→股票池→交易检查”流程
	tracker.HandleTrackerResult(model, sortedSnapshots)
}

// 获取有效的股票代码列表
func getValidStockCodes(tradeRule *config.StrategyParameter) []string {
	stockCodes := tradeRule.StockList()
	if len(stockCodes) == 0 {
		return nil
	}

	validCodes := make([]string, 0, len(stockCodes))
	for _, code := range stockCodes {
		securityCode := exchange.CorrectSecurityCode(code)
		if !exchange.AssertIndexBySecurityCode(securityCode) {
			validCodes = append(validCodes, securityCode)
		}
	}

	return validCodes
}

// 获取股票快照数据
func getStockSnapshots(stockCodes []string) []factors.QuoteSnapshot {
	snapshots := make([]factors.QuoteSnapshot, 0, len(stockCodes))

	for _, code := range stockCodes {
		v := models.SnapshotMgr.GetTickFromMemory(code)
		if v != nil {
			snapshot := models.QuoteSnapshotFromProtocol(*v)
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots
}

// 过滤不符合条件的股票
func filterStocks(model models.Strategy, tradeRule *config.StrategyParameter, snapshots []factors.QuoteSnapshot) []factors.QuoteSnapshot {
	return slices.DeleteFunc(snapshots, func(snapshot factors.QuoteSnapshot) bool {
		err := model.Filter(tradeRule.Rules, snapshot)
		return err != nil // 条件为true时会被删除
	})
}

// 对股票进行排序
func sortStocks(model models.Strategy, snapshots []factors.QuoteSnapshot) []factors.QuoteSnapshot {
	sortedStatus := model.Sort(snapshots)

	if sortedStatus == models.SortDefault || sortedStatus == models.SortNotExecuted {
		// 使用默认排序
		sort.Slice(snapshots, func(i, j int) bool {
			a := snapshots[i]
			b := snapshots[j]
			if a.OpenTurnZ > b.OpenTurnZ {
				return true
			}
			return a.OpenTurnZ == b.OpenTurnZ && a.OpeningChangeRate > b.OpeningChangeRate
		})
	}

	return snapshots
}
