package tracker

import (
	"path/filepath"
	"sync"
	"time"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gox/api"

	"xquant/pkg/cache"
	"xquant/pkg/log"
	"xquant/pkg/models"
	"xquant/pkg/storages"
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
func saveStockPoolToCache(list []StockPool) error {
	filename := getStockPoolFilename()
	// 强制刷新股票池
	err := api.SlicesToCsv(filename, list, true)
	log.Errorf("saveStockPoolToCache error: %s", err)
	_ = err
	return nil
}

// 股票池合并
func stockPoolMerge(model models.Strategy, date string, orders []models.Statistics, maximumNumberOfAvailablePurchases int) {
	poolMutex.Lock()
	defer poolMutex.Unlock()
	localStockPool := storages.GetStockPoolFromCache()
	cacheStatistics := map[string]*StockPool{}
	tradeDate := exchange.FixTradeDate(date)
	for i, v := range orders {
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
			UpdateTime:   v.UpdateTime,
		}
		if i < maximumNumberOfAvailablePurchases {
			//  如果是前排个股标志可以买入
			sp.OrderStatus = 1
		}
		cacheStatistics[sp.Key()] = &sp
	}
	count := len(localStockPool)
	now := time.Now()
	updateTime := now.Format(cache.TimeStampMilli)
	for i := 0; i < count; i++ {
		local := &(localStockPool[i])
		// 1. 非当日的跳过
		if local.Date != tradeDate {
			continue
		}
		v, found := cacheStatistics[local.Key()]
		if found {
			// 相同日期, 策略和证券代码, 视为重复
			// 找到了, 标记为已存在
			v.Status = StrategyAlreadyExists
			//local.OrderStatus = v.OrderStatus
			continue
		}
		// 没找到, 做召回处理
		local.Status.Set(StrategyCancel, true)
		local.UpdateTime = updateTime
	}
	var newList []StockPool
	for _, v := range cacheStatistics {
		if v.Status == StrategyAlreadyExists {
			continue
		}
		v.UpdateTime = updateTime
		log.Infof("%s[%d]: buy queue append %s", model.Name(), model.Code(), v.Code)
		newList = append(newList, *v)
	}
	// 如果有新增标的, 则执行交易指令
	if len(newList) > 0 {
		localStockPool = append(localStockPool, newList...)
		log.Infof("检查是否需要委托下单...")
		storages.CheckOrderForBuy(localStockPool, model, date)
		log.Infof("检查是否需要委托下单...OK")
		err := saveStockPoolToCache(localStockPool)
		if err != nil {
			log.Errorf("saveStockPoolToCache error: %s", err)
		}
	}
}
