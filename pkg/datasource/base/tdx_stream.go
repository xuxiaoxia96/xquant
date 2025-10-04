package base

import (
	"sync"
	"time"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx"
	"gitee.com/quant1x/gotdx/proto"
	"gitee.com/quant1x/gotdx/quotes"
)

// 流式处理器结构
type KLineStreamer struct {
	securityCode string       // 证券代码
	freq         string       // 频率（如 "1min"）
	lastDatetime string       // 上次拉取的最后时间戳
	ticker       *time.Ticker // 定时器
	running      bool         // 运行状态
	mu           sync.Mutex   // 并发锁
	consumer     func(*KLine) // 流式消费函数
}

// 初始化流式处理器
func NewKLineStreamer(securityCode, freq string, consumer func(*KLine)) *KLineStreamer {
	return &KLineStreamer{
		securityCode: exchange.CorrectSecurityCode(securityCode),
		freq:         freq,
		consumer:     consumer,
		//lastDatetime: getLastDatetimeFromCache(securityCode, freq), // 从缓存读取上次最后时间戳
	}
}

// 启动流式处理（定时增量拉取+消费）
func (s *KLineStreamer) Start(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	s.ticker = time.NewTicker(interval)
	go func() {
		for range s.ticker.C {
			s.fetchAndConsume() // 增量拉取并消费
		}
	}()
}

// 增量拉取并流式消费
func (s *KLineStreamer) fetchAndConsume() {
	// 1. 确定增量拉取的起始时间（上次最后时间戳）
	startDate := s.lastDatetime
	if startDate == "" {
		startDate = exchange.MARKET_CH_FIRST_LISTTIME // 首次拉取从最早日期开始
	}

	// 2. 拉取增量数据（复用原有 UpdateAllKLine 逻辑，但仅拉取增量）
	kType := getKTypeByFreq(s.freq) // 根据频率获取 K 线类型（如 1min 对应 proto.KLINE_TYPE_1MIN）
	tdxApi := gotdx.GetTdxApi()
	data, err := tdxApi.GetKLine(s.securityCode, kType, 0, 100) // 拉取最新的 100 条数据（可调整）
	if err != nil || data == nil || len(data.List) == 0 {
		return
	}

	// 3. 过滤增量数据（仅保留上次最后时间戳之后的数据）
	var incrementalBars []quotes.SecurityBar
	for _, bar := range data.List {
		if bar.DateTime > s.lastDatetime {
			incrementalBars = append(incrementalBars, bar)
		}
	}
	if len(incrementalBars) == 0 {
		return
	}

	// 4. 流式消费：逐条处理增量数据（边转换边消费）
	for _, bar := range incrementalBars {
		// 转换为自定义 KLine 结构
		kline := KLine{
			Date:     exchange.FixTradeDate(bar.DateTime),
			Open:     bar.Open,
			Close:    bar.Close,
			High:     bar.High,
			Low:      bar.Low,
			Volume:   bar.Vol * 100, // 手转股
			Amount:   bar.Amount,
			Up:       int(bar.UpCount),
			Down:     int(bar.DownCount),
			Datetime: bar.DateTime,
		}
		// 前复权（按需调用）
		calculatePreAdjustedStockPrice(s.securityCode, []KLine{kline}, startDate)
		// 流式消费：调用自定义消费逻辑（如策略分析、存储到消息队列等）
		s.consumer(&kline)

		// 更新上次最后时间戳
		s.lastDatetime = bar.DateTime
	}
}

// 辅助函数：根据频率获取 K 线类型
func getKTypeByFreq(freq string) uint16 {
	switch freq {
	case "1min":
		return uint16(proto.KLINE_TYPE_1MIN)
	case "5min":
		return uint16(proto.KLINE_TYPE_5MIN)
	case "15min":
		return uint16(proto.KLINE_TYPE_15MIN)
	default:
		return uint16(proto.KLINE_TYPE_RI_K) // 默认日线
	}
}
