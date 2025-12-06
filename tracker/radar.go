package tracker

const (
	// SecurityUnknown 未知证券标识
	SecurityUnknown = "unknown"
)

// 市场雷达
// 通过扫描板块，动态发现表现好的股票，用于策略筛选
//
// 市场雷达的工作原理：
// 1. 扫描所有板块，按涨幅、成交额等指标排序
// 2. 筛选出符合条件的板块（涨幅、成交额等满足条件）
// 3. 从筛选出的板块中提取个股
// 4. 返回股票代码列表，供策略进一步筛选

// ScanSectorForTick 市场雷达：扫描板块并返回股票代码列表
//
// 这是市场雷达的核心功能，通过扫描板块动态发现表现好的股票。
// 函数会：
//   - 扫描所有概念板块（BK_GAINIAN）
//   - 对每个板块内的个股进行排序和筛选
//   - 输出板块排行表格
//   - 返回符合条件的股票代码列表
//
// 参数:
//   - barIndex: 进度条索引指针，用于显示扫描进度，函数执行后会自动递增
//
// 返回:
//   - []string: 从表现好的板块中提取的股票代码列表（已去重）
//
// 示例:
//
//	barIndex := 1
//	stockCodes := ScanSectorForTick(&barIndex)
//	if len(stockCodes) > 0 {
//	    fmt.Printf("市场雷达发现 %d 只股票\n", len(stockCodes))
//	}
func ScanSectorForTick(barIndex *int) []string {
	// 调用 tracker_sector.go 中的实现
	// 该函数会扫描板块、筛选个股、输出表格并返回股票列表
	return scanSectorForTickInternal(barIndex)
}
