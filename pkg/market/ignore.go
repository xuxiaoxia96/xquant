package market

import (
	"strings"

	"gitee.com/quant1x/gotdx/securities"
)

// 定义需要过滤的关键词常量，便于统一维护和扩展
const (
	keywordST       = "ST"
	keywordDelist   = "退"
	keywordDelisted = "摘牌"
)

// 使用集合存储关键词，配合strings.Contains提高多关键词检查效率
var ignoreKeywords = map[string]struct{}{
	keywordST:       {},
	keywordDelist:   {},
	keywordDelisted: {},
}

// IsNeedIgnore 判断个股是否需要忽略（ST、退市、摘牌或信息不存在的个股）
func IsNeedIgnore(securityCode string) bool {
	securityInfo, ok := securities.CheckoutSecurityInfo(securityCode)
	if !ok {
		return true
	}

	name := strings.ToUpper(securityInfo.Name)

	// 检查名称中是否包含任何需要忽略的关键词
	for kw := range ignoreKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}
