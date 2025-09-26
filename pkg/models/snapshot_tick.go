package models

import (
	"context"
	"sync"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx"
	"gitee.com/quant1x/gotdx/quotes"
	"gitee.com/quant1x/gotdx/securities"
	"gitee.com/quant1x/gox/progressbar"
	"gitee.com/quant1x/num"
	"github.com/jinzhu/copier"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/log"
)

var SnapshotMgr *SnapshotManager

func init() {
	SnapshotMgr = NewSnapshotManager()
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
	mu     sync.RWMutex
	cache  map[string]quotes.Snapshot
	tdxAPI *quotes.StdApi
	config config.DataParameter
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		cache:  make(map[string]quotes.Snapshot),
		tdxAPI: gotdx.GetTdxApi(),
		config: config.GetDataConfig(),
	}
}

var (
	__mutexTicks sync.RWMutex
	__cacheTicks = map[string]quotes.Snapshot{}
)

// GetTickFromMemory 从缓存获取快照
func (sm *SnapshotManager) GetTickFromMemory(securityCode string) *quotes.Snapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if snapshot, exists := sm.cache[securityCode]; exists {
		return &snapshot
	}
	return nil
}

// GetStrategySnapshot 获取增强的策略快照
func (sm *SnapshotManager) GetStrategySnapshot(securityCode string) *factors.QuoteSnapshot {
	baseSnapshot := sm.GetTickFromMemory(securityCode)
	if baseSnapshot == nil || baseSnapshot.State != quotes.TDX_SECURITY_TRADE_STATE_NORMAL {
		return nil
	}

	return sm.enhanceSnapshot(baseSnapshot)
}

// enhanceSnapshot 增强快照数据
func (sm *SnapshotManager) enhanceSnapshot(v *quotes.Snapshot) *factors.QuoteSnapshot {
	snapshot := factors.QuoteSnapshot{}
	_ = copier.Copy(&snapshot, v)

	snapshot.Name = securities.GetStockName(v.SecurityCode)
	snapshot.OpeningChangeRate = num.NetChangeRate(snapshot.LastClose, snapshot.Open)
	snapshot.ChangeRate = num.NetChangeRate(snapshot.LastClose, snapshot.Price)

	sm.enrichFinancialData(&snapshot, v.SecurityCode)
	sm.enrichHistoricalData(&snapshot, v.SecurityCode)

	snapshot.OpenBiddingDirection, snapshot.OpenVolumeDirection = v.CheckDirection()

	return &snapshot
}

// enrichFinancialData 丰富财务数据
func (sm *SnapshotManager) enrichFinancialData(snapshot *factors.QuoteSnapshot, securityCode string) {
	if f10 := factors.GetL5F10(securityCode); f10 != nil {
		snapshot.Capital = f10.Capital
		snapshot.FreeCapital = f10.FreeCapital
		snapshot.OpenTurnZ = f10.TurnZ(snapshot.OpenVolume)
	}
}

// enrichHistoricalData 丰富历史数据
func (sm *SnapshotManager) enrichHistoricalData(snapshot *factors.QuoteSnapshot, securityCode string) {
	if history := factors.GetL5History(securityCode); history != nil {
		lastMinuteVolume := history.GetMV5()
		if lastMinuteVolume > 0 {
			snapshot.OpenQuantityRatio = float64(snapshot.OpenVolume) / lastMinuteVolume
			minuteVolume := float64(snapshot.Vol) / float64(exchange.Minutes(snapshot.Date))
			snapshot.QuantityRatio = minuteVolume / lastMinuteVolume
		}
	}
}

// SyncAllSnapshots 实时更新快照
func SyncAllSnapshots(ctx context.Context, barIndex *int) {
	modName := "同步快照数据"
	allCodes := securities.AllCodeList()
	count := len(allCodes)
	var bar *progressbar.Bar = nil
	if barIndex != nil {
		bar = progressbar.NewBar(*barIndex, "执行["+modName+"]", count)
	}
	currentDate := exchange.GetCurrentlyDay()
	tdxApi := gotdx.GetTdxApi()
	// 读取配置的并发数
	parallelCount := config.GetDataConfig().Snapshot.Concurrency
	if parallelCount < 1 {
		parallelCount = tdxApi.NumOfServers()
		parallelCount /= 2
		if parallelCount < config.DefaultMinimumConcurrencyForSnapshots {
			parallelCount = config.DefaultMinimumConcurrencyForSnapshots
		}
	}
	var snapshots []quotes.Snapshot
	var wg sync.WaitGroup
	var mutex sync.Mutex
	codeCh := make(chan []string, parallelCount)

	// 启动goroutine来处理快照获取
	for i := 0; i < parallelCount; i++ {
		go func() {
			for subCodes := range codeCh {
				for i := 0; i < quotes.DefaultRetryTimes; i++ {
					list, err := tdxApi.GetSnapshot(subCodes)
					if err != nil {
						log.CtxErrorf(ctx, "ZS: 网络异常: %+v, 重试: %d", err, i+1)
						continue
					}
					mutex.Lock()
					for _, v := range list {
						// 修订日期
						v.Date = currentDate
						snapshots = append(snapshots, v)
					}
					mutex.Unlock()

					break
				}
			}
			wg.Done()
		}()
	}

	for start := 0; start < count; start += quotes.SECURITY_QUOTES_MAX {
		length := count - start
		if length >= quotes.SECURITY_QUOTES_MAX {
			length = quotes.SECURITY_QUOTES_MAX
		}
		var subCodes []string
		for i := 0; i < length; i++ {
			securityCode := allCodes[start+i]
			subCodes = append(subCodes, securityCode)
			if barIndex != nil {
				bar.Add(1)
			}
		}
		if len(subCodes) == 0 {
			continue
		}
		codeCh <- subCodes
	}
	// channel 关闭后, 仍然可以读, 一直到读完全部数据
	close(codeCh)

	wg.Add(parallelCount)
	wg.Wait()
	// 如果有进度条
	if bar != nil {
		// 等待进度条结束
		bar.Wait()
	}

	__mutexTicks.Lock()
	for _, v := range snapshots {
		__cacheTicks[v.SecurityCode] = v
	}
	__mutexTicks.Unlock()

	if barIndex != nil {
		*barIndex++
	}
}
