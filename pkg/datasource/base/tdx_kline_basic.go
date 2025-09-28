package base

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx"
	"gitee.com/quant1x/gotdx/proto"
	"gitee.com/quant1x/gotdx/quotes"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/num"
	"xquant/pkg/cache"
)

var (
	// DataDaysDiff 日期差异偏移量
	DataDaysDiff = 1
)

// KLine 日K线基础结构
type KLine struct {
	Date     string  `name:"日期" dataframe:"date"`       // 日期
	Open     float64 `name:"开盘" dataframe:"open"`       // 开盘价
	Close    float64 `name:"收盘" dataframe:"close"`      // 收盘价
	High     float64 `name:"最高" dataframe:"high"`       // 最高价
	Low      float64 `name:"最低" dataframe:"low"`        // 最低价
	Volume   float64 `name:"成交量(股)" dataframe:"volume"` // 成交量
	Amount   float64 `name:"成交额(元)" dataframe:"amount"` // 成交金额
	Up       int     `name:"上涨/外盘" dataframe:"up"`      // 上涨家数
	Down     int     `name:"下跌/内盘" dataframe:"down"`    // 下跌家数
	Datetime string  `name:"时间" dataframe:"datetime"`   // 时间
}

func (k *KLine) Apply(factor func(p float64) float64) {
	//TODO implement me
	panic("implement me")
}

func (k *KLine) GetDate() string {
	//TODO implement me
	panic("implement me")
}

func (k *KLine) GetAdjustmentCount() int {
	datetime, _ := api.ParseTime(k.Datetime)
	adjustTimes := int(datetime.UnixMilli() % 1000)
	return adjustTimes
}

// LoadBasicKline 加载基础K线
func LoadBasicKline(securityCode string) []KLine {
	filename := cache.KLineFilename(securityCode)
	var klines []KLine
	_ = api.CsvToSlices(filename, &klines)
	return klines
}

// UpdateAllBasicKLine 更新全部日K线基础数据并保存文件
func UpdateAllBasicKLine(securityCode string) []KLine {
	// 1. 确定本地有效数据最后1条数据作为拉取数据的开始日期
	startDate := exchange.MARKET_CH_FIRST_LISTTIME
	securityCode = exchange.CorrectSecurityCode(securityCode)
	isIndex := exchange.AssertIndexBySecurityCode(securityCode)
	cacheKLines := LoadBasicKline(securityCode)
	kLength := len(cacheKLines)
	var klineDaysOffset = DataDaysDiff
	adjustTimes := 0 // 除权除息的次数
	clearHistory := false
	if kLength > 0 {
		if klineDaysOffset > kLength {
			klineDaysOffset = kLength
		}

		timestampLength := len(api.Timestamp)
		klineFirst := cacheKLines[0]
		klineLast := cacheKLines[kLength-klineDaysOffset]
		totalTimes := TotalAdjustmentTimes(securityCode, klineFirst.Date, klineLast.Date)
		if totalTimes != klineFirst.GetAdjustmentCount() {
			clearHistory = true
		} else if len(klineLast.Datetime) == timestampLength && len(klineFirst.Datetime) == timestampLength {
			// 如果第一条数据和最后一条数据的datetime字段都包括毫秒
			startDate = klineLast.Date
			adjustTimes = klineLast.GetAdjustmentCount()
		} else {
			clearHistory = true
		}
	} else {
		//f10 := flash.GetL5F10(securityCode)
		//if f10 != nil && len(f10.IpoDate) > 0 {
		//	startDate = f10.IpoDate
		//	startDate = trading.FixTradeDate(startDate)
		//}
	}

	if clearHistory {
		// 时间戳格式不对, 清楚缓存清空
		cacheKLines = cacheKLines[0:0]
		kLength = len(cacheKLines)
		// 保持最早的开始日期, 1990-12-19
		adjustTimes = 0
	}

	// 2. 确定结束日期
	currentTradingDate := exchange.GetCurrentlyDay()
	endDate := exchange.Today()
	ts := exchange.TradingDateRange(startDate, endDate)
	history := make([]quotes.SecurityBar, 0)
	step := uint16(quotes.SECURITY_BARS_MAX)
	total := uint16(len(ts))
	start := uint16(0)
	hs := make([]quotes.SecurityBarsReply, 0)
	kType := uint16(proto.KLINE_TYPE_RI_K)
	tdxApi := gotdx.GetTdxApi()
	// 3. 拉取数据
	for {
		count := step
		if total-start >= step {
			count = step
		} else {
			count = total - start
		}
		var data *quotes.SecurityBarsReply
		var err error
		retryTimes := 0
		for retryTimes < quotes.DefaultRetryTimes {
			if isIndex {
				data, err = tdxApi.GetIndexBars(securityCode, kType, start, count)
			} else {
				data, err = tdxApi.GetKLine(securityCode, kType, start, count)
			}
			if err == nil && data != nil {
				break
			}
			retryTimes++
		}
		if err != nil {
			logger.Errorf("code=%s, error=%s", securityCode, err.Error())
			return []KLine{}
		}
		if data == nil || data.Count == 0 || len(data.List) == 0 {
			break
		}
		hs = append(hs, *data)
		if data.Count < count {
			// 已经是最早的记录
			// 需要排序
			break
		}
		start += count
		if start >= total {
			break
		}
	}
	// 4. 由于K线数据，每次获取数据是从后往前获取, 所以这里需要反转历史数据的切片
	hs = api.Reverse(hs)
	startDate = exchange.FixTradeDate(startDate)
	// 5. 调整成交量, 单位从手改成股, vol字段 * 100
	for _, v := range hs {
		for _, row := range v.List {
			date := exchange.FixTradeDate(row.DateTime)
			if date < startDate || date > currentTradingDate {
				continue
			}
			row.Vol = row.Vol * 100
			history = append(history, row)
		}
	}
	// 6. k线数据转换成KLine结构
	var newKLines []KLine
	for _, v := range history {
		date := exchange.FixTradeDate(v.DateTime)
		kline := KLine{
			Date:     date,
			Open:     v.Open,
			Close:    v.Close,
			High:     v.High,
			Low:      v.Low,
			Volume:   v.Vol,
			Amount:   v.Amount,
			Up:       int(v.UpCount),
			Down:     int(v.DownCount),
			Datetime: v.DateTime,
		}
		newKLines = append(newKLines, kline)
	}
	// 下面的流程只判断当前从服务器拉取的数据是否需要除权
	// 判断是否已除权的依据是当前更新K线只有1条记录
	// 并且除权了1次, 那么, 在截断klineDaysOffset的数据后重新拉取是需要复权的
	// 如果除权次数是0, 说明历史数据从尾部取klineDaysOffset条数据是不需要除权的
	isFreshFetchRequireAdjustment := adjustTimes == 1
	if isFreshFetchRequireAdjustment {
		// 只前复权当日数据
		calculatePreAdjustedStockPrice(securityCode, newKLines, startDate)
	}
	// 8. 拼接缓存和新增的数据
	var klines []KLine
	// 8.1 先截取本地缓存的数据
	if kLength > klineDaysOffset {
		klines = cacheKLines[:kLength-klineDaysOffset]
	}
	// 8.2 拼接新增的数据
	if len(klines) > 0 {
		klines = append(klines, newKLines...)
	} else {
		klines = newKLines
	}
	// 7. 前复权
	// 如果前面没复权, 这里用全量数据进行复权
	if !isFreshFetchRequireAdjustment {
		calculatePreAdjustedStockPrice(securityCode, klines, startDate)
	}
	// 9. 刷新缓存文件
	if len(klines) > 0 {
		UpdateCacheKLines(securityCode, klines)
		fname := cache.KLineFilename(securityCode)
		_ = api.SlicesToCsv(fname, klines)
	}
	return klines
}

// 计算前复权 假定缓存中的记录都是截至当日的前一个交易日已经前复权
// startDate 表示已经除权的日期
func calculatePreAdjustedStockPrice(securityCode string, kLines []KLine, startDate string) {
	rows := len(kLines)
	if rows == 0 {
		return
	}
	lastDay := kLines[rows-1].Date
	lastDayNext := exchange.NextTradeDate(lastDay)
	// 复权之前, 假定当前缓存之中的数据都是复权过的数据
	// 那么就应该只拉取缓存最后1条记录之后的除权除息记录进行复权
	// 前复权adjust
	dividends := GetCacheXdxrList(securityCode)
	timestampLength := len(api.Timestamp)
	for i := 0; i < len(dividends); i++ {
		xdxr := dividends[i]
		if xdxr.Category != 1 {
			// 忽略非除权信息
			continue
		}
		if xdxr.Date <= startDate {
			// 忽略除权数据在新数据之前的除权记录
			continue
		}
		if xdxr.Date > lastDayNext {
			// 除权除息数据有可能提前公布
			continue
		}
		xdxrDate := xdxr.Date
		factor := xdxr.Adjust()
		for j := 0; j < rows; j++ {
			kl := &kLines[j]
			barCurrentDate := kl.Date
			if barCurrentDate > xdxrDate {
				break
			}
			if barCurrentDate < xdxrDate {
				kl.Open = factor(kl.Open)
				kl.Close = factor(kl.Close)
				kl.High = factor(kl.High)
				kl.Low = factor(kl.Low)
				// 成交量复权
				// 1. 计算均价线
				maPrice := kl.Amount / kl.Volume
				// 2. 均价线复权
				// 通达信中可能存在没有量复权的情况, 需要在系统设置中的"设置1"勾选分析图中成交量复权
				maPrice = factor(maPrice)
				// 3. 以成交金额为基准, 用复权均价线计算成交量
				kl.Volume = kl.Amount / maPrice
				datetime := ""
				adjustCount := 0
				if len(kl.Datetime) == timestampLength {
					datetime = kl.Datetime[0 : timestampLength-3]
					num, err := strconv.Atoi(kl.Datetime[timestampLength-3:]) // 自动忽略前导零
					if err == nil {
						adjustCount = num
					}
				} else if len(kl.Datetime) == timestampLength-4 {
					datetime = kl.Datetime[0:timestampLength-4] + "."
				} else {
					tm, _ := api.ParseTime(kl.Datetime)
					datetime = tm.Format(api.TimeFormat) + "."
				}
				adjustCount += 1
				kl.Datetime = datetime + fmt.Sprintf("%03d", adjustCount)
			}
			if barCurrentDate == xdxrDate {
				break
			}
		}
	}
}

// TotalAdjustmentTimes 统计 除权次数
//
//	securityCode 证券代码
//	startDate 起始日期
//	endDate 结束日期
func TotalAdjustmentTimes(securityCode string, startDate, endDate string) int {
	startDate = exchange.FixTradeDate(startDate)
	endDate = exchange.FixTradeDate(endDate)
	lastDayNext := exchange.NextTradeDate(endDate)
	// 复权之前, 假定当前缓存之中的数据都是复权过的数据
	// 那么就应该只拉取缓存最后1条记录之后的除权除息记录进行复权
	// 前复权adjust
	dividends := GetCacheXdxrList(securityCode)
	times := 0
	for i := 0; i < len(dividends); i++ {
		xdxr := dividends[i]
		if xdxr.Category != 1 {
			// 忽略非除权信息
			continue
		}
		if xdxr.Date <= startDate {
			// 忽略除权数据在新数据之前的除权记录
			continue
		}
		if xdxr.Date > lastDayNext {
			// 除权除息数据有可能提前公布
			continue
		}
		times++
	}
	return times
}

// LoadKline 加载基础K线
func LoadKline(securityCode string, freq string) []KLine {
	filename := cache.KLineFilenameEx(securityCode, freq)
	var klines []KLine
	_ = api.CsvToSlices(filename, &klines)
	return klines
}

// UpdateAllKLine 更新全部日K线基础数据并保存文件
func UpdateAllKLine(securityCode string, freq ...string) []KLine {
	kType := uint16(proto.KLINE_TYPE_RI_K)
	freq_ := "D"
	period := 1
	numberOfDay := 1
	if len(freq) > 0 {
		duration, err := num.ParseFreq(freq[0])
		// 只支持最大1小时, 超过1小时的, 取日线
		if err == nil && duration <= time.Hour {
			period_ := int(duration / time.Minute)
			if period_ > 0 {
				switch period_ {
				case 5:
					freq_ = "5min"
					kType = uint16(proto.KLINE_TYPE_5MIN)
				case 15:
					freq_ = "15min"
					kType = uint16(proto.KLINE_TYPE_15MIN)
				case 30:
					freq_ = "30min"
					kType = uint16(proto.KLINE_TYPE_30MIN)
				case 60:
					freq_ = "60min"
					kType = uint16(proto.KLINE_TYPE_1HOUR)
				default:
					freq_ = "1min"
					period_ = 1
					kType = uint16(proto.KLINE_TYPE_1MIN)
				}
				period = period_
				numberOfDay = exchange.CN_DEFAULT_TOTALFZNUM / period
			}
		}
	}

	// 1. 确定本地有效数据最后1条数据作为拉取数据的开始日期
	startDate := exchange.MARKET_CH_FIRST_LISTTIME
	securityCode = exchange.CorrectSecurityCode(securityCode)
	isIndex := exchange.AssertIndexBySecurityCode(securityCode)
	cacheKLines := LoadKline(securityCode, freq_)
	kLength := len(cacheKLines)
	var klineDaysOffset = DataDaysDiff * numberOfDay
	adjustTimes := 0 // 除权除息的次数
	clearHistory := false
	if kLength > 0 {
		if klineDaysOffset > kLength {
			klineDaysOffset = kLength
		}

		timestampLength := len(api.Timestamp)
		klineFirst := cacheKLines[0]
		klineLast := cacheKLines[kLength-klineDaysOffset]
		totalTimes := TotalAdjustmentTimes(securityCode, klineFirst.Date, klineLast.Date)
		if totalTimes != klineFirst.GetAdjustmentCount() {
			clearHistory = true
		} else if len(klineLast.Datetime) == timestampLength && len(klineFirst.Datetime) == timestampLength {
			// 如果第一条数据和最后一条数据的datetime字段都包括毫秒
			startDate = klineLast.Date
			adjustTimes = klineLast.GetAdjustmentCount()
		} else {
			clearHistory = true
		}
	} else {
		//f10 := flash.GetL5F10(securityCode)
		//if f10 != nil && len(f10.IpoDate) > 0 {
		//	startDate = f10.IpoDate
		//	startDate = trading.FixTradeDate(startDate)
		//}
	}

	if clearHistory {
		// 时间戳格式不对, 清楚缓存清空
		cacheKLines = cacheKLines[0:0]
		kLength = len(cacheKLines)
		// 保持最早的开始日期, 1990-12-19
		adjustTimes = 0
	}

	// 2. 确定结束日期
	currentTradingDate := exchange.GetCurrentlyDay()
	endDate := exchange.Today()
	ts := exchange.TradingDateRange(startDate, endDate)
	history := make([]quotes.SecurityBar, 0)
	step := uint16(quotes.SECURITY_BARS_MAX)
	max_ := math.MaxUint16

	max_days := max_ / numberOfDay
	days_ := min(max_days, len(ts))
	total_ := days_ * numberOfDay
	total := uint16(total_)
	start := uint16(0)
	hs := make([]quotes.SecurityBarsReply, 0)

	tdxApi := gotdx.GetTdxApi()
	// 3. 拉取数据
	for {
		count := step
		if total-start >= step {
			count = step
		} else {
			count = total - start
		}
		var data *quotes.SecurityBarsReply
		var err error
		retryTimes := 0
		for retryTimes < quotes.DefaultRetryTimes {
			if isIndex {
				data, err = tdxApi.GetIndexBars(securityCode, kType, start, count)
			} else {
				data, err = tdxApi.GetKLine(securityCode, kType, start, count)
			}
			if err == nil && data != nil {
				break
			}
			retryTimes++
		}
		if err != nil {
			logger.Errorf("code=%s, error=%s", securityCode, err.Error())
			return []KLine{}
		}
		if data == nil || data.Count == 0 || len(data.List) == 0 {
			break
		}
		hs = append(hs, *data)
		if data.Count < count {
			// 已经是最早的记录
			// 需要排序
			break
		}
		start += count
		if start >= total {
			break
		}
	}
	// 4. 由于K线数据，每次获取数据是从后往前获取, 所以这里需要反转历史数据的切片
	hs = api.Reverse(hs)
	startDate = exchange.FixTradeDate(startDate)
	// 5. 调整成交量, 单位从手改成股, vol字段 * 100
	for _, v := range hs {
		for _, row := range v.List {
			date := exchange.FixTradeDate(row.DateTime)
			if date < startDate || date > currentTradingDate {
				continue
			}
			row.Vol = row.Vol * 100
			history = append(history, row)
		}
	}
	// 6. k线数据转换成KLine结构
	var newKLines []KLine
	for _, v := range history {
		date := exchange.FixTradeDate(v.DateTime)
		kline := KLine{
			Date:     date,
			Open:     v.Open,
			Close:    v.Close,
			High:     v.High,
			Low:      v.Low,
			Volume:   v.Vol,
			Amount:   v.Amount,
			Up:       int(v.UpCount),
			Down:     int(v.DownCount),
			Datetime: v.DateTime,
		}
		newKLines = append(newKLines, kline)
	}
	// 下面的流程只判断当前从服务器拉取的数据是否需要除权
	// 判断是否已除权的依据是当前更新K线只有1条记录
	// 并且除权了1次, 那么, 在截断klineDaysOffset的数据后重新拉取是需要复权的
	// 如果除权次数是0, 说明历史数据从尾部取klineDaysOffset条数据是不需要除权的
	isFreshFetchRequireAdjustment := adjustTimes == 1
	if isFreshFetchRequireAdjustment {
		// 只前复权当日数据
		calculatePreAdjustedStockPrice(securityCode, newKLines, startDate)
	}
	// 8. 拼接缓存和新增的数据
	var klines []KLine
	// 8.1 先截取本地缓存的数据
	if kLength > klineDaysOffset {
		klines = cacheKLines[:kLength-klineDaysOffset]
	}
	// 8.2 拼接新增的数据
	if len(klines) > 0 {
		klines = append(klines, newKLines...)
	} else {
		klines = newKLines
	}
	// 7. 前复权
	// 如果前面没复权, 这里用全量数据进行复权
	if !isFreshFetchRequireAdjustment {
		calculatePreAdjustedStockPrice(securityCode, klines, startDate)
	}
	// 9. 刷新缓存文件
	if len(klines) > 0 {
		//UpdateCacheKLines(securityCode, klines)
		fname := cache.KLineFilenameEx(securityCode, freq_)
		_ = api.SlicesToCsv(fname, klines)
	}
	return klines
}
