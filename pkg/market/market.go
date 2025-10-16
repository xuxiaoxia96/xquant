package market

import (
	"fmt"
	"math"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gotdx/securities"
)

// 定义交易所代码生成规则：统一管理各市场的代码前缀、起止值、格式化模板
type codeRule struct {
	prefix     string // 代码前缀（如"sh"、"sz"、"hk"）
	codeBegin  int    // 代码起始值（纯数字部分）
	codeEnd    int    // 代码结束值（纯数字部分）
	fmtPattern string // 代码格式化模板（如"sh%d"、"sz000%03d"）
	needFilter bool   // 是否需要过滤（ST、退市等）
}

// generateCodesByRule 根据规则生成单个市场的代码列表
func generateCodesByRule(rule codeRule) []string {
	// 预分配容量：估算代码数量（codeEnd - codeBegin + 1），减少append扩容
	estimatedCount := rule.codeEnd - rule.codeBegin + 1
	codes := make([]string, 0, estimatedCount)

	for i := rule.codeBegin; i <= rule.codeEnd; i++ {
		// 按模板生成完整代码（如"sh600000"）
		fullCode := fmt.Sprintf(rule.fmtPattern, i)
		// 如需过滤，跳过不符合条件的代码
		if rule.needFilter && IsNeedIgnore(fullCode) {
			continue
		}
		codes = append(codes, fullCode)
	}
	return codes
}

// GetStockCodeList 生成全市场股票代码列表（过滤ST、退市、摘牌个股）
func GetStockCodeList() []string {
	// 定义所有交易所的代码生成规则，统一管理
	codeRules := []codeRule{
		// 上海主板：sh600000-sh609999，需要过滤
		{prefix: "sh", codeBegin: 600000, codeEnd: 609999, fmtPattern: "sh%d", needFilter: true},
		// 科创板：sh688000-sh688999，需要过滤
		{prefix: "sh", codeBegin: 688000, codeEnd: 689999, fmtPattern: "sh%d", needFilter: true},
		// 深圳主板：sz000000-sz000999，需要过滤（格式化后补3位，如0→000→sz000000）
		{prefix: "sz", codeBegin: 0, codeEnd: 999, fmtPattern: "sz000%03d", needFilter: true},
		// 中小板：sz001000-sz009999，需要过滤（格式化后补4位，如1000→1000→sz001000）
		{prefix: "sz", codeBegin: 1000, codeEnd: 9999, fmtPattern: "sz00%04d", needFilter: true},
		// 创业板：sz300000-sz300999，需要过滤（格式化后补6位，如300000→300000→sz300000）
		{prefix: "sz", codeBegin: 300000, codeEnd: 309999, fmtPattern: "sz%06d", needFilter: true},
		// 港股：hk00001-hk09999，需要过滤（补5位，如1→00001→hk00001）
		{prefix: "hk", codeBegin: 1, codeEnd: 9999, fmtPattern: "hk%05d", needFilter: true},
	}

	// 合并所有市场的代码（预分配总容量，进一步减少扩容）
	totalEstimatedCount := 0
	for _, rule := range codeRules {
		totalEstimatedCount += rule.codeEnd - rule.codeBegin + 1
	}
	allCodes := make([]string, 0, totalEstimatedCount)

	for _, rule := range codeRules {
		allCodes = append(allCodes, generateCodesByRule(rule)...)
	}

	return allCodes
}

// GetCodeList 加载全部代码列表（指数+板块+股票）
func GetCodeList() []string {
	// 预分配容量：指数代码数 + 板块代码数 + 股票代码数（估算）
	indexCount := len(exchange.IndexList())
	blockCount := len(securities.BlockList())
	stockCountEstimate := 20000 // 全市场股票约1.5万-2万只，预留冗余
	allCodes := make([]string, 0, indexCount+blockCount+stockCountEstimate)

	// 追加指数代码
	allCodes = append(allCodes, exchange.IndexList()...)
	// 追加板块代码
	blocks := securities.BlockList()
	for _, block := range blocks {
		allCodes = append(allCodes, block.Code)
	}
	// 追加股票代码
	stockCodes := GetStockCodeList()
	allCodes = append(allCodes, stockCodes...)

	return allCodes
}

// 保留2位小数（适用于A股）
func roundTo2Decimal(num float64) float64 {
	return math.Round(num*100) / 100
}

// PriceLimit 计算涨停板和跌停板的价格
func PriceLimit(securityCode string, lastClose float64) (limitUp, limitDown float64) {
	limitRate := exchange.MarketLimit(securityCode)
	// 计算涨跌停价并保留2位小数
	limitUp = roundTo2Decimal(lastClose * (1.0 + limitRate))
	limitDown = roundTo2Decimal(lastClose * (1.0 - limitRate))
	return limitUp, limitDown
}
