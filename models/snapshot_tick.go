package models

import (
	"sync"

	"gitee.com/quant1x/data/exchange"
	"gitee.com/quant1x/data/level1"
	"gitee.com/quant1x/data/level1/quotes"
	"gitee.com/quant1x/data/level1/securities"
	"xquant/config"
	"xquant/factors"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/progressbar"
	"gitee.com/quant1x/num"
)

// ============================================
// 接口定义
// ============================================

// SnapshotAPI 快照数据源接口
// 用于获取实时行情快照数据
type SnapshotAPI interface {
	// GetSnapshot 批量获取快照数据
	GetSnapshot(codes []string) ([]quotes.Snapshot, error)
	// NumOfServers 获取服务器数量（用于计算并发数）
	NumOfServers() int
}

// SnapshotCache 快照缓存接口
// 用于管理快照数据的缓存
type SnapshotCache interface {
	// Get 获取指定证券的快照
	Get(securityCode string) (*quotes.Snapshot, bool)
	// Put 存储快照到缓存
	Put(securityCode string, snapshot quotes.Snapshot)
	// PutBatch 批量存储快照
	PutBatch(snapshots []quotes.Snapshot)
	// Size 获取缓存大小
	Size() int
	// Clear 清空缓存
	Clear()
}

// SnapshotEnricher 快照数据增强接口
// 用于补充快照的额外数据（F10、历史数据等）
type SnapshotEnricher interface {
	// EnrichF10Data 补充F10基本面数据
	EnrichF10Data(snapshot *factors.QuoteSnapshot, securityCode string)
	// EnrichHistoryData 补充历史数据并计算量比
	EnrichHistoryData(snapshot *factors.QuoteSnapshot, securityCode string)
}

// SnapshotService 快照服务接口
// 提供完整的快照服务功能
type SnapshotService interface {
	// GetStrategySnapshot 获取策略快照（包含完整数据）
	GetStrategySnapshot(securityCode string) *factors.QuoteSnapshot
	// SyncAllSnapshots 同步所有快照数据
	SyncAllSnapshots(barIndex *int)
}

// ============================================
// 实现类
// ============================================

// memorySnapshotCache 内存快照缓存实现
type memorySnapshotCache struct {
	mu    sync.RWMutex
	cache map[string]quotes.Snapshot
}

// NewMemorySnapshotCache 创建内存快照缓存
func NewMemorySnapshotCache() SnapshotCache {
	return &memorySnapshotCache{
		cache: make(map[string]quotes.Snapshot),
	}
}

func (c *memorySnapshotCache) Get(securityCode string) (*quotes.Snapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, found := c.cache[securityCode]
	if !found {
		return nil, false
	}
	// 返回副本的指针，避免外部修改影响缓存
	return &v, true
}

func (c *memorySnapshotCache) Put(securityCode string, snapshot quotes.Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[securityCode] = snapshot
}

func (c *memorySnapshotCache) PutBatch(snapshots []quotes.Snapshot) {
	if len(snapshots) == 0 {
		return
	}

	// 构建新缓存（不需要加锁）
	newCache := make(map[string]quotes.Snapshot, len(snapshots))
	for i := range snapshots {
		newCache[snapshots[i].SecurityCode] = snapshots[i]
	}

	// 原子合并到现有缓存
	c.mu.Lock()
	for code, snapshot := range newCache {
		c.cache[code] = snapshot
	}
	c.mu.Unlock()
}

func (c *memorySnapshotCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func (c *memorySnapshotCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]quotes.Snapshot)
}

// defaultSnapshotEnricher 默认快照数据增强器
type defaultSnapshotEnricher struct{}

// NewDefaultSnapshotEnricher 创建默认快照数据增强器
func NewDefaultSnapshotEnricher() SnapshotEnricher {
	return &defaultSnapshotEnricher{}
}

func (e *defaultSnapshotEnricher) EnrichF10Data(snapshot *factors.QuoteSnapshot, securityCode string) {
	f10 := factors.GetL5F10(securityCode)
	if f10 == nil {
		return
	}

	snapshot.Capital = f10.Capital
	snapshot.FreeCapital = f10.FreeCapital
	snapshot.OpenTurnZ = f10.TurnZ(snapshot.OpenVolume)
}

func (e *defaultSnapshotEnricher) EnrichHistoryData(snapshot *factors.QuoteSnapshot, securityCode string) {
	history := factors.GetL5History(securityCode)
	if history == nil {
		return
	}

	lastMinuteVolume := history.GetMV5()
	if lastMinuteVolume <= 0 {
		// 避免除零，如果历史数据无效则跳过
		return
	}

	// 计算开盘量比
	snapshot.OpenQuantityRatio = float64(snapshot.OpenVolume) / lastMinuteVolume

	// 计算实时量比
	minutes := exchange.Minutes(snapshot.Date)
	if minutes > 0 {
		minuteVolume := float64(snapshot.Vol) / float64(minutes)
		snapshot.QuantityRatio = minuteVolume / lastMinuteVolume
	}
}

// snapshotService 快照服务实现
type snapshotService struct {
	cache    SnapshotCache
	enricher SnapshotEnricher
	api      SnapshotAPI
}

// NewSnapshotService 创建快照服务
func NewSnapshotService(cache SnapshotCache, enricher SnapshotEnricher, api SnapshotAPI) SnapshotService {
	return &snapshotService{
		cache:    cache,
		enricher: enricher,
		api:      api,
	}
}

func (s *snapshotService) GetStrategySnapshot(securityCode string) *factors.QuoteSnapshot {
	// 1. 从缓存获取原始快照
	rawSnapshot, found := s.cache.Get(securityCode)
	if !found || rawSnapshot == nil {
		return nil
	}

	// 2. 检查交易状态
	if rawSnapshot.State != quotes.SECURITY_TRADE_STATE_NORMAL {
		return nil
	}

	// 3. 转换为策略快照
	snapshot := s.convertToQuoteSnapshot(rawSnapshot, securityCode)

	// 4. 补充F10数据
	s.enricher.EnrichF10Data(&snapshot, securityCode)

	// 5. 补充历史数据
	s.enricher.EnrichHistoryData(&snapshot, securityCode)

	// 6. 计算委托方向
	snapshot.OpenBiddingDirection, snapshot.OpenVolumeDirection = rawSnapshot.CheckDirection()

	return &snapshot
}

func (s *snapshotService) convertToQuoteSnapshot(v *quotes.Snapshot, securityCode string) factors.QuoteSnapshot {
	snapshot := factors.QuoteSnapshot{}
	if err := api.Copy(&snapshot, v); err != nil {
		logger.Warnf("复制快照数据失败: %s, error: %v", securityCode, err)
	}

	snapshot.Name = securities.GetStockName(securityCode)
	snapshot.OpeningChangeRate = num.NetChangeRate(snapshot.LastClose, snapshot.Open)
	snapshot.ChangeRate = num.NetChangeRate(snapshot.LastClose, snapshot.Price)

	return snapshot
}

func (s *snapshotService) SyncAllSnapshots(barIndex *int) {
	modName := "同步快照数据"
	allCodes := securities.AllCodeList()
	count := len(allCodes)

	// 初始化进度条
	var bar *progressbar.Bar
	if barIndex != nil {
		bar = progressbar.NewBar(*barIndex, "执行["+modName+"]", count)
	}

	currentDate := exchange.GetCurrentlyDay()

	// 计算并发数
	parallelCount := s.calculateParallelCount()

	// 使用 channel 传递任务
	codeCh := make(chan []string, parallelCount)
	resultCh := make(chan []quotes.Snapshot, parallelCount)

	var wg sync.WaitGroup

	// 启动 worker goroutines
	wg.Add(parallelCount)
	for i := 0; i < parallelCount; i++ {
		go s.snapshotWorker(i, codeCh, resultCh, currentDate, &wg)
	}

	// 发送任务到 channel
	go s.sendTasks(allCodes, codeCh, bar, barIndex)

	// 收集所有结果
	var allSnapshots []quotes.Snapshot
	done := make(chan struct{})
	go func() {
		for i := 0; i < parallelCount; i++ {
			snapshots := <-resultCh
			allSnapshots = append(allSnapshots, snapshots...)
		}
		close(done)
	}()

	// 等待所有 worker 完成
	wg.Wait()
	close(resultCh)

	// 等待结果收集完成
	<-done

	// 批量更新缓存
	s.cache.PutBatch(allSnapshots)

	// 等待进度条结束
	if bar != nil {
		bar.Wait()
	}

	// 更新进度条索引
	if barIndex != nil {
		*barIndex++
	}
}

func (s *snapshotService) calculateParallelCount() int {
	parallelCount := config.GetDataConfig().Snapshot.Concurrency
	if parallelCount < 1 {
		parallelCount = s.api.NumOfServers()
		parallelCount /= 2
		if parallelCount < config.DefaultMinimumConcurrencyForSnapshots {
			parallelCount = config.DefaultMinimumConcurrencyForSnapshots
		}
	}
	return parallelCount
}

func (s *snapshotService) snapshotWorker(
	workerID int,
	codeCh <-chan []string,
	resultCh chan<- []quotes.Snapshot,
	currentDate string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var localSnapshots []quotes.Snapshot

	// 处理任务队列
	for subCodes := range codeCh {
		// 重试机制
		var list []quotes.Snapshot
		var err error

		for retry := 0; retry < quotes.DefaultRetryTimes; retry++ {
			list, err = s.api.GetSnapshot(subCodes)
			if err == nil {
				break
			}
			logger.Errorf("Worker[%d] 网络异常: %+v, 重试: %d/%d", workerID, err, retry+1, quotes.DefaultRetryTimes)
		}

		// 处理获取到的快照
		for i := range list {
			// 修订日期
			list[i].Date = currentDate
			localSnapshots = append(localSnapshots, list[i])
		}
	}

	// 发送结果
	resultCh <- localSnapshots
}

func (s *snapshotService) sendTasks(
	allCodes []string,
	codeCh chan<- []string,
	bar *progressbar.Bar,
	barIndex *int,
) {
	defer close(codeCh) // 所有任务发送完成后关闭 channel

	count := len(allCodes)
	for start := 0; start < count; start += quotes.SECURITY_QUOTES_MAX {
		length := count - start
		if length > quotes.SECURITY_QUOTES_MAX {
			length = quotes.SECURITY_QUOTES_MAX
		}

		// 构建子任务
		subCodes := make([]string, 0, length)
		for i := 0; i < length; i++ {
			subCodes = append(subCodes, allCodes[start+i])
			if bar != nil {
				bar.Add(1)
			}
		}

		if len(subCodes) == 0 {
			continue
		}

		// 发送任务
		codeCh <- subCodes
	}
}

// ============================================
// 默认实例（向后兼容）
// ============================================

var (
	// defaultCache 默认缓存实例
	defaultCache = NewMemorySnapshotCache()
	// defaultEnricher 默认数据增强器
	defaultEnricher = NewDefaultSnapshotEnricher()
	// defaultService 默认服务实例（延迟初始化）
	defaultService SnapshotService
	// defaultServiceOnce 确保默认服务只初始化一次
	defaultServiceOnce sync.Once
)

// getDefaultService 获取默认服务实例（延迟初始化）
func getDefaultService() SnapshotService {
	defaultServiceOnce.Do(func() {
		api := &level1SnapshotAPI{api: level1.GetApi()}
		defaultService = NewSnapshotService(defaultCache, defaultEnricher, api)
	})
	return defaultService
}

// level1SnapshotAPI level1 API 适配器
type level1SnapshotAPI struct {
	api interface {
		GetSnapshot(codes []string) ([]quotes.Snapshot, error)
		NumOfServers() int
	}
}

func (a *level1SnapshotAPI) GetSnapshot(codes []string) ([]quotes.Snapshot, error) {
	return a.api.GetSnapshot(codes)
}

func (a *level1SnapshotAPI) NumOfServers() int {
	return a.api.NumOfServers()
}

// ============================================
// 全局便捷函数（向后兼容）
// ============================================

// GetTickFromMemory 从内存缓存中获取快照
// 返回快照指针，如果不存在则返回 nil
// 这是向后兼容的便捷函数，内部使用默认缓存
func GetTickFromMemory(securityCode string) *quotes.Snapshot {
	v, found := defaultCache.Get(securityCode)
	if !found {
		return nil
	}
	return v
}

// GetStrategySnapshot 从缓存中获取策略快照
// 将原始快照转换为策略使用的 QuoteSnapshot，并补充相关数据
// 如果快照不存在或证券非正常交易状态，返回 nil
// 这是向后兼容的便捷函数，内部使用默认服务
func GetStrategySnapshot(securityCode string) *factors.QuoteSnapshot {
	return getDefaultService().GetStrategySnapshot(securityCode)
}

// SyncAllSnapshots 实时更新所有快照数据
// 使用并发方式批量获取快照，并更新到内存缓存
// barIndex: 进度条索引，可以为 nil
// 这是向后兼容的便捷函数，内部使用默认服务
func SyncAllSnapshots(barIndex *int) {
	getDefaultService().SyncAllSnapshots(barIndex)
}

// GetCacheSize 获取缓存大小（用于监控和调试）
// 这是向后兼容的便捷函数，内部使用默认缓存
func GetCacheSize() int {
	return defaultCache.Size()
}

// ClearCache 清空缓存（用于测试和重置）
// 这是向后兼容的便捷函数，内部使用默认缓存
func ClearCache() {
	defaultCache.Clear()
}
