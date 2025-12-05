package storages

import (
	"path/filepath"
	"sync"
	"time"

	"xquant/cache"
	"xquant/config"
	"xquant/models"

	"gitee.com/quant1x/data/exchange"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
)

const (
	filenameStockPool = "stock_pool.csv"
)

var (
	poolMutex sync.Mutex
)

// 股票池文件
func getStockPoolFilename() string {
	filename := filepath.Join(cache.GetQmtCachePath(), filenameStockPool)
	return filename
}

// 从本地缓存加载股票池
func getStockPoolFromCache() (list []StockPool) {
	filename := getStockPoolFilename()
	err := api.CsvToSlices(filename, &list)
	_ = err
	return
}

// 刷新本地股票池缓存
func saveStockPoolToCache(list []StockPool) {
	filename := getStockPoolFilename()
	// 强制刷新股票池
	err := api.SlicesToCsv(filename, list, true)
	_ = err
}

// UpdateStockPoolAndExecuteTrading 更新股票池并执行交易
// 这是盘内跟踪交易的核心入口函数：
// 1. 将策略扫描结果合并到股票池
// 2. 如果有新增标的，执行买入交易
func UpdateStockPoolAndExecuteTrading(model models.Strategy, date string, statistics []models.Statistics) {
	tradeRule := config.GetStrategyParameterByCode(model.Code())
	if tradeRule == nil || !tradeRule.Enable() || tradeRule.Total == 0 {
		// 配置不存在, 或者规则无效, 不执行交易
		return
	}
	topN := tradeRule.Total
	mergeStockPoolAndExecuteTrading(model, date, statistics, topN)
}

// mergeStockPoolAndExecuteTrading 合并股票池并执行交易
// 这是盘内跟踪交易的核心函数，包含两个主要步骤：
// 1. 合并策略扫描结果到股票池（标记新增、召回等）
// 2. 如果有新增标的，执行买入交易
func mergeStockPoolAndExecuteTrading(model models.Strategy, date string, statistics []models.Statistics, maximumNumberOfAvailablePurchases int) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// 1. 合并股票池：将策略扫描结果合并到本地股票池
	localStockPool, newStocks := mergeStockPool(model, date, statistics, maximumNumberOfAvailablePurchases)

	// 2. 如果有新增标的，执行交易并保存
	if len(newStocks) > 0 {
		localStockPool = append(localStockPool, newStocks...)
		logger.Infof("检查是否需要委托下单...")
		checkOrderForBuy(localStockPool, model, date)
		logger.Infof("检查是否需要委托下单...OK")
		saveStockPoolToCache(localStockPool)
	}
}

// mergeStockPool 合并股票池
// 将策略扫描结果合并到本地股票池，返回更新后的股票池和新增的标的列表
func mergeStockPool(model models.Strategy, date string, statistics []models.Statistics, maximumNumberOfAvailablePurchases int) ([]StockPool, []StockPool) {
	localStockPool := getStockPoolFromCache()
	tradeDate := exchange.FixTradeDate(date)
	now := time.Now()
	updateTime := now.Format(cache.TimeStampMilli)

	// 1. 将策略扫描结果转换为股票池格式
	newStockPoolMap := buildStockPoolMapFromStatistics(model, statistics, maximumNumberOfAvailablePurchases, updateTime)

	// 2. 处理本地股票池：标记已存在的标的，召回不再出现的标的
	processExistingStockPool(localStockPool, newStockPoolMap, tradeDate, updateTime)

	// 3. 提取新增的标的（未在本地股票池中存在的）
	newStocks := extractNewStocks(newStockPoolMap, model, updateTime)

	return localStockPool, newStocks
}

// buildStockPoolMapFromStatistics 从统计数据构建股票池映射
func buildStockPoolMapFromStatistics(model models.Strategy, statistics []models.Statistics, maximumNumberOfAvailablePurchases int, updateTime string) map[string]*StockPool {
	stockPoolMap := make(map[string]*StockPool, len(statistics))

	for i, v := range statistics {
		sp := StockPool{
			Status:       StrategyHit,
			Date:         v.Date,
			Code:         v.Code,
			Name:         v.Name,
			Buy:          v.Price,
			StrategyCode: model.Code(),
			StrategyName: model.Name(),
			OrderStatus:  0, // 股票池订单状态默认是0
			Active:       v.Active,
			Speed:        v.Speed,
			CreateTime:   v.UpdateTime,
			UpdateTime:   updateTime,
		}

		// 如果是前排个股，标记为可买入
		if i < maximumNumberOfAvailablePurchases {
			sp.OrderStatus = 1
		}

		stockPoolMap[sp.Key()] = &sp
	}

	return stockPoolMap
}

// processExistingStockPool 处理本地已存在的股票池
// 标记已存在的标的，召回不再出现的标的
func processExistingStockPool(localStockPool []StockPool, newStockPoolMap map[string]*StockPool, tradeDate string, updateTime string) {
	for i := range localStockPool {
		local := &localStockPool[i]

		// 非当日的跳过
		if local.Date != tradeDate {
			continue
		}

		// 检查是否在新的扫描结果中存在
		newStock, found := newStockPoolMap[local.Key()]
		if found {
			// 找到了，标记为已存在（避免重复添加）
			newStock.Status = StrategyAlreadyExists
		} else {
			// 没找到，做召回处理
			local.Status.Set(StrategyCancel, true)
			local.UpdateTime = updateTime
		}
	}
}

// extractNewStocks 提取新增的标的（未在本地股票池中存在的）
func extractNewStocks(newStockPoolMap map[string]*StockPool, model models.Strategy, updateTime string) []StockPool {
	var newStocks []StockPool

	for _, stock := range newStockPoolMap {
		// 跳过已存在的标的
		if stock.Status == StrategyAlreadyExists {
			continue
		}

		stock.UpdateTime = updateTime
		logger.Infof("%s[%d]: buy queue append %s", model.Name(), model.Code(), stock.Code)
		newStocks = append(newStocks, *stock)
	}

	return newStocks
}
