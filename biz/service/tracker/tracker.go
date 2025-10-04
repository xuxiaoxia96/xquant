package update

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"gitee.com/quant1x/exchange"
	"golang.org/x/exp/slices"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/models"
	"xquant/pkg/tracker"
)

// --------------------------
// 1. 定义核心参数结构体（统一 HTTP 和命令行输入）
// --------------------------
// TrackerCoreParams 跟踪核心的输入参数（HTTP 和命令行入口均需构造此参数）
type TrackerCoreParams struct {
	TrackerStrategyCodes []uint64 // 待跟踪的策略代码列表
	IsDebug              bool     // 是否开启调试模式（非交易时段也执行）
}

// RunTrackerCore 跟踪核心逻辑：创建可取消上下文，传递给单次任务
// 该函数可被 HTTP 入口、命令行入口直接调用，逻辑完全一致
func RunTrackerCore(ctx context.Context, params TrackerCoreParams) {
	// 1. 基于外部上下文，创建可取消上下文（保留取消函数）
	coreCtx, coreCancel := context.WithCancel(ctx)
	defer coreCancel() // 确保退出时释放资源，避免内存泄漏

	// 校验核心参数（空策略列表直接返回）
	if len(params.TrackerStrategyCodes) == 0 {
		log.CtxWarnf(coreCtx, "[TrackerCore] 未指定跟踪的策略代码，终止跟踪")
		return
	}

	// 初始化定时器（1秒间隔）
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// 核心循环：使用 coreCtx（可取消）而非外部 ctx
	for {
		select {
		case <-coreCtx.Done(): // 监听可取消上下文的终止信号
			log.CtxWarnf(coreCtx, "[TrackerCore] 收到退出信号，停止跟踪")
			return
		case <-ticker.C:
			// 2. 将 coreCtx 和 coreCancel 传入单次任务（允许单次任务终止循环）
			if err := executeSingleTrack(coreCtx, coreCancel, params); err != nil {
				log.CtxErrorf(coreCtx, "[TrackerCore] 单次跟踪任务执行失败: %v", err)
			}
		}
	}
}

// executeSingleTrack 执行单次跟踪任务（接收取消函数，支持终止循环）
// 参数新增：cancel func() —— 用于终止外层 coreCtx 循环
func executeSingleTrack(ctx context.Context, cancel func(), params TrackerCoreParams) error {
	// 1. 检查当前是否允许实时更新（交易时段/调试模式判断）
	updateInRealTime, exchangeStatus := exchange.CanUpdateInRealtime()
	isAllowed := updateInRealTime &&
		(exchangeStatus == exchange.ExchangeTrading || exchangeStatus == exchange.ExchangeSuspend)

	// 非交易时段且非调试模式：调用取消函数，终止外层循环
	if !isAllowed && !params.IsDebug {
		log.CtxInfof(ctx, "[TrackerCore] 非交易时段（状态：%v）且未开启调试模式，停止跟踪", exchangeStatus)
		cancel() // 调用取消函数，触发 coreCtx.Done()，终止外层循环
		return nil
	}

	// 休市暂停状态：跳过本次跟踪（不终止循环）
	if exchangeStatus == exchange.ExchangeSuspend {
		log.CtxDebugf(ctx, "[TrackerCore] 当前为休市暂停状态，跳过本次跟踪")
		return nil
	}

	// 2. 同步所有股票快照（准备数据）
	barIndex := 1
	models.SnapshotMgr.SyncAllSnapshots(ctx, &barIndex)
	// 3. 并行处理所有策略
	var wg sync.WaitGroup
	for _, strategyCode := range params.TrackerStrategyCodes {
		wg.Add(1)
		go func(code uint64) {
			defer wg.Done()
			if err := processSingleStrategy(ctx, code, params.IsDebug); err != nil {
				log.CtxErrorf(ctx, "[TrackerCore] 处理策略 %d 失败: %v", code, err)
			}
		}(strategyCode)
	}
	wg.Wait()

	return nil
}

// processSingleStrategy 处理单个策略的跟踪逻辑（并行任务内的核心）
func processSingleStrategy(ctx context.Context, strategyCode uint64, isDebug bool) error {
	// 1. 获取策略实例
	strategy, err := models.CheckoutStrategy(strategyCode)
	if err != nil {
		return fmt.Errorf("获取策略实例失败: %w", err)
	}
	if strategy == nil {
		return fmt.Errorf("策略 %d 不存在", strategyCode)
	}

	// 2. 获取策略参数配置
	strategyParam := config.GetStrategyParameterByCode(strategyCode)
	if strategyParam == nil {
		return fmt.Errorf("策略 %d 无参数配置", strategyCode)
	}

	// 3. 检查执行条件（交易时段 或 调试模式）
	if !strategyParam.Session.IsTrading() && !isDebug {
		log.CtxDebugf(ctx, "[TrackerCore] 策略 %d 非交易时段且未开启调试模式，跳过", strategyCode)
		return nil
	}

	// 4. 执行策略快照跟踪（数据过滤、排序、输出）
	snapshotTracker(strategy, strategyParam)
	return nil
}

// --------------------------
// 3. 原辅助函数保留（仅调整命名和参数，确保无 HTTP 依赖）
// --------------------------
// snapshotTracker 策略快照跟踪：处理股票筛选、排序、结果输出
func snapshotTracker(model models.Strategy, tradeRule *config.StrategyParameter) {
	if tradeRule == nil {
		log.CtxDebugf(context.Background(), "[TrackerCore] 策略参数为空，跳过快照跟踪")
		return
	}

	// 1. 获取有效股票代码
	stockCodes := getValidStockCodes(tradeRule)
	if len(stockCodes) == 0 {
		log.CtxDebugf(context.Background(), "[TrackerCore] 无有效股票代码，跳过快照跟踪")
		return
	}

	// 2. 获取股票快照数据
	stockSnapshots := getStockSnapshots(stockCodes)
	if len(stockSnapshots) == 0 {
		log.CtxDebugf(context.Background(), "[TrackerCore] 未获取到股票快照数据，跳过快照跟踪")
		return
	}

	// 3. 过滤不符合条件的股票
	filteredSnapshots := filterStocks(model, tradeRule, stockSnapshots)

	// 4. 对股票进行排序
	sortedSnapshots := sortStocks(model, filteredSnapshots)

	// 5. 最终结果处理（表格输出、股票池更新、交易检查）
	tracker.HandleTrackerResult(model, sortedSnapshots)
}

// getValidStockCodes 获取有效股票代码（过滤指数代码）
func getValidStockCodes(tradeRule *config.StrategyParameter) []string {
	stockCodes := tradeRule.StockList()
	if len(stockCodes) == 0 {
		return nil
	}

	validCodes := make([]string, 0, len(stockCodes))
	for _, code := range stockCodes {
		securityCode := exchange.CorrectSecurityCode(code)
		// 排除指数代码，保留股票代码
		if !exchange.AssertIndexBySecurityCode(securityCode) {
			validCodes = append(validCodes, securityCode)
		}
	}
	return validCodes
}

// getStockSnapshots 从内存获取股票快照数据
func getStockSnapshots(stockCodes []string) []factors.QuoteSnapshot {
	snapshots := make([]factors.QuoteSnapshot, 0, len(stockCodes))
	for _, code := range stockCodes {
		if tick := models.SnapshotMgr.GetTickFromMemory(code); tick != nil {
			snapshots = append(snapshots, models.QuoteSnapshotFromProtocol(*tick))
		}
	}
	return snapshots
}

// filterStocks 按策略规则过滤股票（保留符合条件的）
func filterStocks(model models.Strategy, tradeRule *config.StrategyParameter, snapshots []factors.QuoteSnapshot) []factors.QuoteSnapshot {
	// DeleteFunc：返回 true 表示删除该元素（即过滤掉不符合条件的）
	return slices.DeleteFunc(snapshots, func(snapshot factors.QuoteSnapshot) bool {
		return model.Filter(tradeRule.Rules, snapshot) != nil
	})
}

// sortStocks 按策略规则排序股票（无策略排序时用默认规则）
func sortStocks(model models.Strategy, snapshots []factors.QuoteSnapshot) []factors.QuoteSnapshot {
	sortedStatus := model.Sort(snapshots)
	// 策略未指定排序规则：使用默认排序（换手率降序 → 开盘涨跌幅降序）
	if sortedStatus == models.SortDefault || sortedStatus == models.SortNotExecuted {
		sort.Slice(snapshots, func(i, j int) bool {
			a, b := snapshots[i], snapshots[j]
			if a.OpenTurnZ != b.OpenTurnZ {
				return a.OpenTurnZ > b.OpenTurnZ
			}
			return a.OpeningChangeRate > b.OpeningChangeRate
		})
	}
	return snapshots
}
