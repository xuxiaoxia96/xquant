package services

import (
	"context"
	"time"

	"gitee.com/quant1x/exchange"

	"xquant/pkg/log"
	"xquant/pkg/models"
)

// 任务 - 更新快照
func jobUpdateSnapshot() {
	now := time.Now()
	updateInRealTime, status := exchange.CanUpdateInRealtime(now)

	// 交易时间更新数据
	if updateInRealTime && (IsTrading(status) || exchange.CheckCallAuctionClose(now)) {
		realtimeUpdateSnapshot()
	}
	// debug环境可以 realtimeUpdateSnapshot()
}

// 更新快照
func realtimeUpdateSnapshot() {
	log.Infof("同步snapshot...")
	models.SnapshotMgr.SyncAllSnapshots(context.Background(), nil)
	log.Infof("同步snapshot...OK")
}
