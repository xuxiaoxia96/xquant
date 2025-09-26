package tracker

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gox/tags"
	"gitee.com/quant1x/num"
	"gitee.com/quant1x/pkg/tablewriter"

	"xquant/pkg/cache"
	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/models"
	"xquant/pkg/storages"
	"xquant/pkg/trader"
)

// HandleTrackerResult 跟踪结果处理器：协调“数据转换→表格输出→股票池→交易检查”流程
// 作为 snapshotTracker 的“输出结果”环节入口，职责仅为流程调度
func HandleTrackerResult(model models.Strategy, sortedSnapshots []factors.QuoteSnapshot) {
	// 1. 第一步：将快照转换为统计模型（Statistics）
	stats, currentDate, updateTime, err := buildStatistics(sortedSnapshots)
	if err != nil {
		log.Errorf("策略[%s]：统计数据构建失败：%v", model.Name(), err)
		return
	}
	if len(stats) == 0 {
		log.Debugf("策略[%s]：无有效统计数据，跳过后续处理", model.Name())
		return
	}

	// 2. 第二步：控制台输出表格（仅负责渲染，不掺杂其他逻辑）
	renderConsoleTable(stats, currentDate, updateTime)

	// 3. 第三步：处理股票池（合并数据+更新缓存）
	if err := processStockPool(model, currentDate, stats); err != nil {
		log.Errorf("策略[%s]：股票池处理失败：%v", model.Name(), err)
		return
	}

	// 4. 第四步：检查并触发买入订单（仅交易相关逻辑）
	if success := triggerBuyCheck(model, currentDate, stats); !success {
		log.Warnf("策略[%s]：买入检查未完成或未满足条件", model.Name())
	}
}

// buildStatistics 将股票快照转换为统计模型（Statistics）
// 同时计算当前交易日、更新时间，纯数据处理，无副作用
func buildStatistics(snapshots []factors.QuoteSnapshot) ([]models.Statistics, string, string, error) {
	// 1. 计算基础时间信息（当前交易日、更新时间）
	today := time.Now().Format("2006-01-02")
	dates := exchange.TradingDateRange(exchange.MARKET_CN_FIRST_DATE, today)
	if len(dates) == 0 {
		return nil, "", "", errors.New("无有效交易日数据")
	}
	currentDate := dates[len(dates)-1]
	updateTime := "15:00:59"

	// 处理当日交易时段逻辑
	now := time.Now()
	nowTime := now.Format(exchange.CN_SERVERTIME_FORMAT)
	if today == currentDate {
		if nowTime < exchange.CN_TradingStartTime {
			// 未开盘：取前一交易日
			if len(dates) < 2 {
				return nil, "", "", errors.New("无历史交易日数据")
			}
			currentDate = dates[len(dates)-2]
		} else if nowTime >= exchange.CN_TradingStartTime && nowTime <= exchange.CN_TradingStopTime {
			// 盘中：用当前时间作为更新时间
			updateTime = now.Format(time.TimeOnly)
		}
	}

	// 2. 快照转换为统计模型
	var stats []models.Statistics
	for _, snap := range snapshots {
		// 基础字段赋值
		stat := models.Statistics{
			Date:                 currentDate,
			Code:                 snap.SecurityCode,
			Name:                 snap.Name,
			Active:               int(snap.Active),
			LastClose:            snap.LastClose,
			Open:                 snap.Open,
			OpenRaise:            snap.OpeningChangeRate,
			Price:                snap.Price,
			UpRate:               snap.ChangeRate,
			OpenPremiumRate:      snap.PremiumRate,
			OpenVolume:           snap.OpenVolume,
			TurnZ:                snap.OpenTurnZ,
			QuantityRatio:        snap.OpenQuantityRatio,
			AveragePrice:         snap.Amount / float64(snap.Vol), // 计算均价
			Speed:                snap.Rate,
			ChangePower:          snap.ChangePower,
			AverageBiddingVolume: snap.AverageBiddingVolume,
			UpdateTime:           factors.GetTimestamp(),
		}

		// 计算趋势描述（低开/平开/高开 + 回落/拉升 + 强势/弱势）
		if snap.Open < snap.LastClose {
			stat.Tendency += "低开"
		} else if snap.Open == snap.LastClose {
			stat.Tendency += "平开"
		} else {
			stat.Tendency += "高开"
		}
		if stat.AveragePrice < snap.Open {
			stat.Tendency += ",回落"
		} else {
			stat.Tendency += ",拉升"
		}
		if snap.Price > stat.AveragePrice {
			stat.Tendency += ",强势"
		} else {
			stat.Tendency += ",弱势"
		}

		// 补充板块信息（从全局缓存获取）
		if bs, ok := __stock2Block[stat.Code]; ok && len(bs) > 0 {
			if block, blockOk := __mapBlockData[bs[0].Code]; blockOk {
				stat.BlockName = block.Name
				stat.BlockRate = block.ChangeRate
				stat.BlockTop = block.Rank
				if shot, shotOk := __stock2Rank[stat.Code]; shotOk {
					stat.BlockRank = shot.TopNo
				}
			}
		}

		stats = append(stats, stat)
	}

	return stats, currentDate, updateTime, nil
}

// renderConsoleTable 仅负责控制台表格输出
// 输入：统计数据、当前交易日、更新时间；输出：控制台表格+胜率统计
func renderConsoleTable(stats []models.Statistics, currentDate, updateTime string) {
	if len(stats) == 0 {
		return
	}

	// 1. 初始化表格并设置表头
	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.SetHeader(tags.GetHeadersByTags(models.Statistics{}))

	// 2. 填充表格数据
	for _, stat := range stats {
		tbl.Append(tags.GetValuesByTags(stat))
	}

	// 3. 计算胜率统计
	gtP1, gtP2, gtP3, gtP4, gtP5, avgYield := calculateWinRate(stats)
	count := len(stats)

	// 4. 渲染输出
	fmt.Println() // 换行分隔
	tbl.Render()
	fmt.Println()

	// 输出胜率统计
	fmt.Printf("%s %s, 胜率统计:\n", currentDate, updateTime)
	fmt.Printf("\t==> 胜    率: %d/%d, %.2f%%, 收益率: %.2f%%\n", gtP1, count, 100*float64(gtP1)/float64(count), avgYield)
	fmt.Printf("\t==> 溢价超1%%: %d/%d, %.2f%%\n", gtP2, count, 100*float64(gtP2)/float64(count))
	fmt.Printf("\t==> 溢价超2%%: %d/%d, %.2f%%\n", gtP3, count, 100*float64(gtP3)/float64(count))
	fmt.Printf("\t==> 溢价超3%%: %d/%d, %.2f%%\n", gtP4, count, 100*float64(gtP4)/float64(count))
	fmt.Printf("\t==> 溢价超5%%: %d/%d, %.2f%%\n", gtP5, count, 100*float64(gtP5)/float64(count))
	fmt.Println()
}

// calculateWinRate 辅助函数：计算胜率和平均收益率（纯数值计算）
func calculateWinRate(stats []models.Statistics) (int, int, int, int, int, float64) {
	var gtP1, gtP2, gtP3, gtP4, gtP5 int
	var totalYield float64

	for _, stat := range stats {
		// 计算开盘价到现价的收益率
		rate := num.NetChangeRate(stat.Open, stat.Price)
		totalYield += rate

		// 统计不同溢价区间的数量
		if rate > 0 {
			gtP1++
		}
		if rate >= 1.00 {
			gtP2++
		}
		if rate >= 2.00 {
			gtP3++
		}
		if rate >= 3.00 {
			gtP4++
		}
		if rate >= 5.00 {
			gtP5++
		}
	}

	// 计算平均收益率（避免除零）
	avgYield := 0.0
	if len(stats) > 0 {
		avgYield = totalYield / float64(len(stats))
	}

	return gtP1, gtP2, gtP3, gtP4, gtP5, avgYield
}

// processStockPool 处理股票池：合并新数据、更新缓存
// 依赖全局变量 __stock2Block、__mapBlockData、__stock2Rank、poolMutex
func processStockPool(model models.Strategy, date string, stats []models.Statistics) error {
	// 1. 先获取策略参数，判断是否需要处理股票池
	tradeRule := config.GetStrategyParameterByCode(model.Code())
	if tradeRule == nil || !tradeRule.Enable() || tradeRule.Total == 0 {
		return errors.New("策略未启用或无配置，跳过股票池处理")
	}
	topN := tradeRule.Total
	tradeDate := exchange.FixTradeDate(date)

	// 2. 加锁处理股票池（避免并发冲突）
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// 3. 从缓存读取本地股票池
	localStockPool := getStockPoolFromCache()
	if localStockPool == nil {
		localStockPool = make([]StockPool, 0)
	}

	// 4. 转换统计数据为股票池格式，存入临时缓存
	cacheStats := make(map[string]*StockPool)
	updateTime := time.Now().Format(cache.TimeStampMilli)
	orderCreateTime := stats[0].UpdateTime // 复用统计数据的创建时间

	for i, stat := range stats {
		sp := StockPool{
			Status:       StrategyHit,
			Date:         tradeDate,
			Code:         stat.Code,
			Name:         stat.Name,
			Buy:          stat.Price,
			StrategyCode: model.Code(),
			StrategyName: model.Name(),
			OrderStatus:  0, // 默认不可买入
			Active:       stat.Active,
			Speed:        stat.Speed,
			CreateTime:   orderCreateTime,
			UpdateTime:   updateTime,
		}
		// 前排个股标记为可买入（不超过策略配置的topN）
		if i < topN {
			sp.OrderStatus = 1
		}
		cacheStats[sp.Key()] = &sp
	}

	// 5. 处理本地股票池中的旧数据（标记已存在/取消）
	for i := range localStockPool {
		local := &localStockPool[i]
		// 跳过非当日数据
		if local.Date != tradeDate {
			continue
		}
		// 检查是否在新数据中存在
		if v, exists := cacheStats[local.Key()]; exists {
			v.Status = StrategyAlreadyExists // 标记为已存在，后续跳过
		} else {
			// 新数据中不存在：标记为取消
			local.Status.Set(StrategyCancel, true)
			local.UpdateTime = updateTime
		}
	}

	// 6. 收集新增的股票池数据（排除已存在的）
	var newStockPool []StockPool
	for _, v := range cacheStats {
		if v.Status == StrategyAlreadyExists {
			continue
		}
		v.UpdateTime = updateTime
		log.Infof("%s[%d]: 股票池新增标的 %s", model.Name(), model.Code(), v.Code)
		newStockPool = append(newStockPool, *v)
	}

	// 7. 新增数据写入缓存
	if len(newStockPool) > 0 {
		localStockPool = append(localStockPool, newStockPool...)
		if err := saveStockPoolToCache(localStockPool); err != nil {
			return fmt.Errorf("股票池缓存保存失败：%w", err)
		}
		log.Infof("%s[%d]: 股票池更新完成，新增%d个标的", model.Name(), model.Code(), len(newStockPool))
	}

	return nil
}

// triggerBuyCheck 检查买入条件并触发下单
// 仅负责交易相关逻辑，不处理数据转换或缓存
func triggerBuyCheck(model models.Strategy, date string, stats []models.Statistics) bool {
	// 1. 基础条件校验：交易日、策略配置、交易时段
	if !exchange.DateIsTradingDay() {
		log.Errorf("%s[%d]: 非交易日，跳过买入检查", model.Name(), model.Code())
		return false
	}

	tradeRule := config.GetStrategyParameterByCode(model.Code())
	if tradeRule == nil || !tradeRule.BuyEnable() {
		log.Errorf("%s[%d]: 策略未配置买入或买入未启用，跳过", model.Name(), model.Code())
		return false
	}

	if !tradeRule.Session.IsTrading() {
		log.Errorf("%s[%d]: 非交易时段，跳过买入检查", model.Name(), model.Code())
		return false
	}

	// 2. 初始化变量
	tradeDate := exchange.FixTradeDate(date)
	direction := trader.BUY
	maxBuyTotal := tradeRule.Total
	// 统计已买入的标的数量
	boughtCount := storages.CountStrategyOrders(tradeDate, model, direction)
	if boughtCount >= maxBuyTotal {
		log.Infof("%s[%d]: 买入配额已用完（计划%d，已完成%d）", model.Name(), model.Code(), maxBuyTotal, boughtCount)
		return true // 已完成目标，视为成功
	}

	// 3. 筛选可买入的标的（符合策略+可买入状态）
	var buyTargets []*StockPool
	// 先从股票池缓存获取最新数据（确保包含新增标的）
	poolMutex.Lock()
	localPool := getStockPoolFromCache()
	poolMutex.Unlock()

	for _, v := range localPool {
		// 过滤条件：当日数据+当前策略+可买入状态+未买入
		if v.Date != tradeDate || v.StrategyCode != model.Code() || v.OrderStatus != 1 {
			continue
		}
		// 检查是否已买入
		if storages.CheckOrderState(tradeDate, model, v.Code, direction) {
			log.Infof("%s[%d]: 标的%s已买入，跳过", model.Name(), model.Code(), v.Code)
			continue
		}
		// 检查是否禁止买入
		if !trader.CheckForBuy(v.Code) {
			log.Infof("%s[%d]: 标的%s被禁止买入，跳过", model.Name(), model.Code(), v.Code)
			continue
		}
		buyTargets = append(buyTargets, &v)
	}

	// 4. 计算单标的可用资金
	// 非实时订单：按实际可买入数量核定资金；实时订单：按策略配额核定
	isTickOrder := tradeRule.Flag == models.OrderFlagTick
	quotaForTargets := maxBuyTotal
	if !isTickOrder {
		quotaForTargets = len(buyTargets)
	}
	if quotaForTargets < 1 {
		log.Errorf("%s[%d]: 无符合条件的买入标的，跳过", model.Name(), model.Code())
		return false
	}

	// 调用交易接口计算单标的可用资金
	singleFunds := trader.CalculateAvailableFundsForSingleTarget(
		quotaForTargets,
		tradeRule.Weight,
		tradeRule.FeeMax,
		tradeRule.FeeMin,
	)
	if singleFunds <= trader.InvalidFee {
		log.Errorf("%s[%d]: 单标的可用资金为0，跳过", model.Name(), model.Code())
		return false
	}

	// 5. 执行买入下单
	completedCount := boughtCount
	for _, target := range buyTargets {
		if completedCount >= maxBuyTotal {
			break // 已达到最大买入数量，停止
		}

		securityCode := target.Code
		// 再次校验买入状态（避免并发问题）
		if storages.CheckOrderState(tradeDate, model, securityCode, direction) {
			log.Errorf("%s[%d]: 标的%s已买入，跳过", model.Name(), model.Code(), securityCode)
			continue
		}

		// 标记订单状态（防止重复下单）
		if err := storages.PushOrderState(tradeDate, model, securityCode, direction); err != nil {
			log.Errorf("%s[%d]: 标的%s订单状态推送失败：%v，跳过", model.Name(), model.Code(), securityCode, err)
			continue
		}

		// 价格笼子计算（合规价格）
		buyPrice := trader.CalculatePriceCage(*tradeRule, direction, target.Buy)
		// 计算买入费用和数量
		tradeFee := trader.EvaluateFeeForBuy(securityCode, singleFunds, buyPrice)
		if tradeFee.Volume <= trader.InvalidVolume {
			log.Errorf("%s[%d]: 标的%s可买数量为0，跳过", model.Name(), model.Code(), securityCode)
			continue
		}

		// 执行下单
		orderID, err := trader.PlaceOrder(
			direction,
			model,
			securityCode,
			trader.FIX_PRICE,
			tradeFee.Price,
			tradeFee.Volume,
		)
		if err != nil || orderID < 0 {
			target.Status |= StrategyOrderFailed
			log.Errorf("%s[%d]: 标的%s下单失败：%v", model.Name(), model.Code(), securityCode, err)
			continue
		}

		// 下单成功：更新状态和订单ID
		target.Status |= StrategyOrderSucceeded | StrategyOrderPlaced
		target.OrderId = orderID
		completedCount++
		log.Infof("%s[%d]: 标的%s下单成功，订单ID：%d", model.Name(), model.Code(), securityCode, orderID)
	}

	// 返回是否完成全部买入配额
	return completedCount >= maxBuyTotal
}
