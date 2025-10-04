package factors

import (
	"sync"

	"gitee.com/quant1x/exchange"

	"xquant/pkg/cache"
	"xquant/pkg/datasource/east_money"
	"xquant/pkg/log"
)

var (
	__mapMarginTradingTargets   = map[string]east_money.SecurityMarginTrading{}
	__mutexMarginTradingTargets sync.RWMutex
)

// MarginTradingTargetInit 一次性缓存两融数据, 交易日9点后更新上一个交易的两融数据
func MarginTradingTargetInit(date string) {
	__mutexMarginTradingTargets.Lock()
	defer __mutexMarginTradingTargets.Unlock()
	clear(__mapMarginTradingTargets)
	_, featureDate := cache.CorrectDate(date)
	list := east_money.GetMarginTradingList(featureDate)
	if len(list) == 0 {
		log.Errorf("date = %s, 没有融资融券数据", date)
		return
	}
	for _, v := range list {
		securityCode := exchange.CorrectSecurityCode(v.SecuCode)
		__mapMarginTradingTargets[securityCode] = v
	}
}

// GetMarginTradingTarget 获取两融数据
func GetMarginTradingTarget(code string) (east_money.SecurityMarginTrading, bool) {
	__mutexMarginTradingTargets.RLock()
	defer __mutexMarginTradingTargets.RUnlock()
	securityCode := exchange.CorrectSecurityCode(code)
	v, ok := __mapMarginTradingTargets[securityCode]
	return v, ok
}
