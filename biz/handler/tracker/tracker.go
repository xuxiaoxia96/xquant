package tracker

import (
	"context"
	"sort"
	"time"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/runtime"
	"github.com/cloudwego/hertz/pkg/app"

	"xquant/biz/handler"
	"xquant/biz/model/tracker"
	"xquant/biz/models"
	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/openapi_error"
)

// Tracker 实时跟踪策略在当前市场的表现，输出表格
func Tracker(ctx context.Context, c *app.RequestContext) {
	var err error
	var req tracker.TrackerRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		log.CtxErrorf(ctx, "[Tracker] error: %s", err)
		handler.OpenAPIFail(ctx, c, openapi_error.NewInvalidParameterError(ctx, "", err.Error()))
		return
	}

	trackerStrategyCodes := req.GetTrackerStrategyCodes()
	for {
		// 前置判断
		// 能否实时更新, 收盘时间前停止更新
		updateInRealTime, status := exchange.CanUpdateInRealtime()
		isTrading := updateInRealTime && (status == exchange.ExchangeTrading || status == exchange.ExchangeSuspend)
		// 休市中交易暂停
		if !isTrading {
			// 非调试且非交易时段返回
			return
		}
		if status == exchange.ExchangeSuspend {
			time.Sleep(time.Second * 1)
			continue
		}

		// 进度条
		// 收集信息到缓存，__cacheTicks
		models.SyncAllSnapshots()

		// 策略列表
		for _, trackerStrategyCode := range trackerStrategyCodes {
			// 获取策略对象
			strategy, err := models.CheckoutStrategy(trackerStrategyCode)
			if err != nil || strategy == nil {
				continue
			}

			// 获取策略参数
			strategyParameter := config.GetStrategyParameterByCode(trackerStrategyCode)
			if strategyParameter == nil {
				continue
			}

			if strategyParameter.Session.IsTrading() {
				// 核心代码
				SnapshotTracker(strategy, strategyParameter)
			} else {
				if runtime.Debug() {
					SnapshotTracker(strategy, strategyParameter)
				} else {
					break
				}
			}
		}
		time.Sleep(time.Second * 1)
	}
}

// SnapshotTracker 策略配置 quant-engine/engine/config/resources/quant1x.yaml
func SnapshotTracker(model models.Strategy, tradeRule *config.StrategyParameter) {
	if tradeRule == nil {
		return
	}

	// 获取股票代码列表（策略作用于的股票）
	stockCodes := tradeRule.StockList()
	if len(stockCodes) == 0 {
		return
	}

	// 即时行情快照(副本)
	var stockSnapshots []factors.QuoteSnapshot
	stockCount := len(stockCodes)
	//bar := progressbar.NewBar(*barIndex, "执行["+model.Name()+"全市场扫描]", stockCount)
	for start := 0; start < stockCount; start++ {
		//bar.Add(1)
		code := stockCodes[start]
		securityCode := exchange.CorrectSecurityCode(code)

		if exchange.AssertIndexBySecurityCode(securityCode) {
			continue
		}

		v := models.GetTickFromMemory(securityCode)
		if v != nil {
			// 转换
			snapshot := models.QuoteSnapshotFromProtocol(*v)
			// 加入即使行情
			stockSnapshots = append(stockSnapshots, snapshot)
		}
	}

	if len(stockSnapshots) == 0 {
		return
	}

	// 过滤不符合条件的个股
	stockSnapshots = api.Filter(stockSnapshots, func(snapshot factors.QuoteSnapshot) bool {
		err := model.Filter(tradeRule.Rules, snapshot)
		return err == nil
	})

	// 结果集排序
	sortedStatus := model.Sort(stockSnapshots)
	if sortedStatus == models.SortDefault || sortedStatus == models.SortNotExecuted {
		// 默认排序或者排序未执行, 使用默认排序
		sort.Slice(stockSnapshots, func(i, j int) bool {
			a := stockSnapshots[i]
			b := stockSnapshots[j]
			if a.OpenTurnZ > b.OpenTurnZ {
				return true
			}
			return a.OpenTurnZ == b.OpenTurnZ && a.OpeningChangeRate > b.OpeningChangeRate
		})
	}

	// 输出表格，利用该策略对行情的结果
	//OutputTable(model, stockSnapshots)
}
