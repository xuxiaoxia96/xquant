package tracker

import (
	"sort"
	"time"

	"fmt"
	"os"

	"xquant/factors"
	"xquant/models"
	"xquant/storages"

	"gitee.com/quant1x/data/exchange"
	"gitee.com/quant1x/gox/tags"
	"gitee.com/quant1x/num"
	"gitee.com/quant1x/pkg/tablewriter"

	"xquant/config"
	"xquant/permissions"

	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/progressbar"
	"gitee.com/quant1x/gox/runtime"
)

// Tracker 盘中跟踪
func Tracker(strategyNumbers ...uint64) {
	for {
		updateInRealTime, status := exchange.CanUpdateInRealtime()
		isTrading := updateInRealTime && (status == exchange.ExchangeTrading || status == exchange.ExchangeSuspend)
		if !runtime.Debug() && !isTrading {
			// 非调试且非交易时段返回
			return
		}
		if status == exchange.ExchangeSuspend {
			time.Sleep(time.Second * 1)
			continue
		}
		barIndex := 1
		models.SyncAllSnapshots(&barIndex)
		//stockCodes := radar.ScanSectorForTick(barIndex)
		for _, strategyNumber := range strategyNumbers {
			model, err := models.CheckoutStrategy(strategyNumber)
			if err != nil || model == nil {
				continue
			}
			err = permissions.CheckPermission(model)
			if err != nil {
				logger.Error(err)
				continue
			}
			strategyParameter := config.GetStrategyParameterByCode(strategyNumber)
			if strategyParameter == nil {
				continue
			}
			if strategyParameter.Session.IsTrading() {
				snapshotTracker(&barIndex, model, strategyParameter)
			} else {
				if runtime.Debug() {
					snapshotTracker(&barIndex, model, strategyParameter)
				} else {
					break
				}
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func snapshotTracker(barIndex *int, model models.Strategy, tradeRule *config.StrategyParameter) {
	if tradeRule == nil {
		return
	}
	stockCodes := tradeRule.StockList()
	if len(stockCodes) == 0 {
		return
	}
	var stockSnapshots []factors.QuoteSnapshot
	stockCount := len(stockCodes)
	bar := progressbar.NewBar(*barIndex, "执行["+model.Name()+"全市场扫描]", stockCount)
	for start := 0; start < stockCount; start++ {
		bar.Add(1)
		code := stockCodes[start]
		securityCode := exchange.CorrectSecurityCode(code)
		if exchange.AssertIndexBySecurityCode(securityCode) {
			continue
		}
		v := models.GetTickFromMemory(securityCode)
		if v != nil {
			snapshot := models.QuoteSnapshotFromProtocol(*v)
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
	// 处理策略扫描结果：构建统计、渲染表格、更新股票池并执行交易
	ProcessStrategyResults(model, stockSnapshots)
}

// ProcessStrategyResults 处理策略扫描结果
// 包括：构建统计数据、渲染表格、计算胜率、更新股票池并执行交易
func ProcessStrategyResults(model models.Strategy, stockSnapshots []factors.QuoteSnapshot) {
	// 1. 获取当前交易日期和时间
	currentlyDay, updateTime := getCurrentTradeDateAndTime()

	// 2. 构建统计数据
	statistics := buildStatisticsFromSnapshots(stockSnapshots, currentlyDay)

	// 3. 渲染表格并计算胜率
	renderTableAndWinRate(statistics, currentlyDay, updateTime)

	// 4. 更新股票池并执行交易（如果配置了交易）
	storages.UpdateStockPoolAndExecuteTrading(model, currentlyDay, statistics)
}

// getCurrentTradeDateAndTime 获取当前交易日期和时间
func getCurrentTradeDateAndTime() (string, string) {
	today := exchange.IndexToday()
	dates := exchange.TradeRange(exchange.MARKET_CN_FIRST_DATE, today)
	days := len(dates)
	currentlyDay := dates[days-1]
	updateTime := "15:00:59"

	if today == currentlyDay {
		now := time.Now()
		nowTime := now.Format(exchange.CN_SERVERTIME_FORMAT)
		if nowTime < exchange.CN_TradingStartTime {
			currentlyDay = dates[days-2]
		} else if nowTime >= exchange.CN_TradingStartTime && nowTime <= exchange.CN_TradingStopTime {
			updateTime = now.Format(exchange.TimeOnly)
		}
	}
	return currentlyDay, updateTime
}

// buildStatisticsFromSnapshots 从快照数据构建统计数据
func buildStatisticsFromSnapshots(stockSnapshots []factors.QuoteSnapshot, currentlyDay string) []models.Statistics {
	votingResults := make([]models.Statistics, 0, len(stockSnapshots))
	orderCreateTime := factors.GetTimestamp()

	for _, v := range stockSnapshots {
		ticket := models.Statistics{
			Date:                 currentlyDay,              // 日期
			Code:                 v.SecurityCode,            // 证券代码
			Name:                 v.Name,                    // 证券名称
			Active:               int(v.Active),             // 活跃度
			LastClose:            v.LastClose,               // 昨收
			Open:                 v.Open,                    // 开盘价
			OpenRaise:            v.OpeningChangeRate,       // 开盘涨幅
			Price:                v.Price,                   // 现价
			UpRate:               v.ChangeRate,              // 涨跌幅
			OpenPremiumRate:      v.PremiumRate,             // 集合竞价买入溢价率
			OpenVolume:           v.OpenVolume,              // 集合竞价-开盘量, 单位是股
			TurnZ:                v.OpenTurnZ,               // 开盘换手率z
			QuantityRatio:        v.OpenQuantityRatio,       // 开盘量比
			AveragePrice:         v.Amount / float64(v.Vol), // 均价
			Speed:                v.Rate,                    // 涨速
			ChangePower:          v.ChangePower,             // 涨跌力度
			AverageBiddingVolume: v.AverageBiddingVolume,    // 委比
			UpdateTime:           orderCreateTime,           // 更新时间
		}

		// 计算趋势
		ticket.Tendency = calculateTendency(v, ticket.AveragePrice)

		// 填充板块信息
		fillBlockInfo(&ticket)

		votingResults = append(votingResults, ticket)
	}
	return votingResults
}

// calculateTendency 计算趋势描述
func calculateTendency(snapshot factors.QuoteSnapshot, averagePrice float64) string {
	var tendency string

	// 开盘情况
	if snapshot.Open < snapshot.LastClose {
		tendency += "低开"
	} else if snapshot.Open == snapshot.LastClose {
		tendency += "平开"
	} else {
		tendency += "高开"
	}

	// 均价相对开盘
	if averagePrice < snapshot.Open {
		tendency += ",回落"
	} else {
		tendency += ",拉升"
	}

	// 现价相对均价
	if snapshot.Price > averagePrice {
		tendency += ",强势"
	} else {
		tendency += ",弱势"
	}

	return tendency
}

// fillBlockInfo 填充板块信息
func fillBlockInfo(ticket *models.Statistics) {
	bs, ok := __stock2Block[ticket.Code]
	if !ok {
		return
	}

	tb := bs[0]
	block, ok := __mapBlockData[tb.Code]
	if !ok {
		return
	}

	ticket.BlockName = block.Name
	ticket.BlockRate = block.ChangeRate
	ticket.BlockTop = block.Rank

	shot, ok := __stock2Rank[ticket.Code]
	if ok {
		ticket.BlockRank = shot.TopNo
	}
}

// renderTableAndWinRate 渲染表格并计算胜率统计
func renderTableAndWinRate(statistics []models.Statistics, currentlyDay, updateTime string) {
	// 渲染表格
	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.SetHeader(tags.GetHeadersByTags(models.Statistics{}))

	// 计算胜率统计
	winRateStats := calculateWinRateStatistics(statistics)

	// 填充表格数据
	for _, v := range statistics {
		tbl.Append(tags.GetValuesByTags(v))
	}

	// 输出表格
	fmt.Println() // 输出一个换行
	tbl.Render()
	fmt.Println()

	// 输出胜率统计
	printWinRateStatistics(winRateStats, currentlyDay, updateTime, len(statistics))
}

// WinRateStatistics 胜率统计结果
type WinRateStatistics struct {
	WinCount     int     // 胜率（存在溢价）
	Over1Percent int     // 超过1%
	Over2Percent int     // 超过2%
	Over3Percent int     // 超过3%
	Over5Percent int     // 超过5%
	AverageYield float64 // 平均收益率
}

// calculateWinRateStatistics 计算胜率统计
func calculateWinRateStatistics(statistics []models.Statistics) WinRateStatistics {
	stats := WinRateStatistics{}
	totalYield := 0.0

	for _, v := range statistics {
		rate := num.NetChangeRate(v.Open, v.Price)
		if rate > 0 {
			stats.WinCount++
		}
		if rate >= 1.00 {
			stats.Over1Percent++
		}
		if rate >= 2.00 {
			stats.Over2Percent++
		}
		if rate >= 3.00 {
			stats.Over3Percent++
		}
		if rate >= 5.00 {
			stats.Over5Percent++
		}
		totalYield += rate
	}

	if len(statistics) > 0 {
		stats.AverageYield = totalYield / float64(len(statistics))
	}

	return stats
}

// printWinRateStatistics 打印胜率统计
func printWinRateStatistics(stats WinRateStatistics, currentlyDay, updateTime string, totalCount int) {
	fmt.Println(currentlyDay + " " + updateTime + ", 胜率统计:")
	fmt.Printf("\t==> 胜    率: %d/%d, %.2f%%, 收益率: %.2f%%\n",
		stats.WinCount, totalCount, 100*float64(stats.WinCount)/float64(totalCount), stats.AverageYield)
	fmt.Printf("\t==> 溢价超1%%: %d/%d, %.2f%%\n", stats.Over1Percent, totalCount, 100*float64(stats.Over1Percent)/float64(totalCount))
	fmt.Printf("\t==> 溢价超2%%: %d/%d, %.2f%%\n", stats.Over2Percent, totalCount, 100*float64(stats.Over2Percent)/float64(totalCount))
	fmt.Printf("\t==> 溢价超3%%: %d/%d, %.2f%%\n", stats.Over3Percent, totalCount, 100*float64(stats.Over3Percent)/float64(totalCount))
	fmt.Printf("\t==> 溢价超5%%: %d/%d, %.2f%%\n", stats.Over5Percent, totalCount, 100*float64(stats.Over5Percent)/float64(totalCount))
	fmt.Println()
}
