package services

import (
	"github.com/jinzhu/copier"

	"xquant/pkg/cache"
	"xquant/pkg/factors"
	"xquant/pkg/log"
	"xquant/pkg/market"
)

func jobUpdateMarginTrading() {
	log.Infof("同步融资融券...")
	date := cache.DefaultCanReadDate()
	factors.MarginTradingTargetInit(date)
	updateMarginTradingForMisc(date)
	updateMarginTradingForMarginTrading(date)
	log.Infof("同步融资融券...OK")
}

func updateMarginTradingForMisc(cacheDate string) {
	allCodes := market.GetCodeList()
	for _, securityCode := range allCodes {
		misc := factors.GetL5Misc(securityCode, cacheDate)
		if misc == nil {
			continue
		}
		marginTrading, ok := factors.GetMarginTradingTarget(securityCode)
		if ok {
			misc.RZYEZB = marginTrading.RZYEZB
			misc.UpdateTime = factors.GetTimestamp()
			factors.UpdateL5Misc(misc)
		}
	}
	factors.RefreshL5Misc()
}

func updateMarginTradingForMarginTrading(cacheDate string) {
	allCodes := market.GetCodeList()
	for _, securityCode := range allCodes {
		smt := factors.GetL5SecuritiesMarginTrading(securityCode, cacheDate)
		if smt == nil {
			continue
		}
		marginTrading, ok := factors.GetMarginTradingTarget(securityCode)
		if ok {
			_ = copier.Copy(smt, &marginTrading)
			smt.UpdateTime = factors.GetTimestamp()
			smt.State |= factors.FeatureSecuritiesMarginTrading
			factors.UpdateL5SecuritiesMarginTrading(smt)
		}
	}
	factors.RefreshL5SecuritiesMarginTrading()
}
