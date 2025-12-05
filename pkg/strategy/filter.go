package strategy

import (
	"fmt"

	"xquant/pkg/config"
	"xquant/pkg/factors"
	"xquant/pkg/rules"
)

// ============================================
// 类型定义
// ============================================

// FilterFunc 过滤器函数类型
// 统一的过滤器函数签名，便于组合和复用
type FilterFunc func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error

// ============================================
// 基础过滤器
// ============================================

// GeneralFilter 通用过滤条件
//
//	执行所有在册的规则（F10规则 + 基础规则）
func GeneralFilter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	passed, failed, err := rules.Filter(ruleParameter, snapshot)
	_ = passed
	_ = failed
	return err
}

// ============================================
// 过滤器组合函数
// ============================================

// ChainFilters 链式组合多个过滤器
// 按顺序执行所有过滤器，遇到第一个错误立即返回（短路模式）
//
//	示例：
//		ChainFilters(
//			GeneralFilter,
//			PriceRangeFilter(10.0, 100.0),
//			VolumeMinFilter(1000000),
//		)
func ChainFilters(filters ...FilterFunc) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		for _, filter := range filters {
			if filter == nil {
				continue
			}
			if err := filter(ruleParameter, snapshot); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithCustomFilter 组合通用过滤器和自定义过滤器
// 这是一个便捷函数，先执行通用规则过滤，再执行自定义过滤逻辑
//
//	示例：
//		func (s MyStrategy) Filter(ruleParameter, snapshot) error {
//			return WithCustomFilter(ruleParameter, snapshot,
//				func(p config.RuleParameter, s factors.QuoteSnapshot) error {
//					if s.Price < 10.0 {
//						return fmt.Errorf("价格太低")
//					}
//					return nil
//				},
//			)
//		}
func WithCustomFilter(
	ruleParameter config.RuleParameter,
	snapshot factors.QuoteSnapshot,
	customFilters ...FilterFunc,
) error {
	// 先执行通用规则过滤
	if err := GeneralFilter(ruleParameter, snapshot); err != nil {
		return err
	}

	// 再执行自定义过滤器
	for _, filter := range customFilters {
		if filter == nil {
			continue
		}
		if err := filter(ruleParameter, snapshot); err != nil {
			return err
		}
	}

	return nil
}

// ============================================
// 常用过滤器构建函数
// ============================================

// PriceRangeFilter 价格范围过滤器
// 创建一个检查价格范围的过滤器
//
//	示例：PriceRangeFilter(10.0, 100.0)  // 价格必须在 10-100 之间
func PriceRangeFilter(minPrice, maxPrice float64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if snapshot.Price < minPrice {
			return fmt.Errorf("价格 %.2f 低于最低价 %.2f", snapshot.Price, minPrice)
		}
		if snapshot.Price > maxPrice {
			return fmt.Errorf("价格 %.2f 高于最高价 %.2f", snapshot.Price, maxPrice)
		}
		return nil
	}
}

// PriceMinFilter 最低价过滤器
func PriceMinFilter(minPrice float64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if snapshot.Price < minPrice {
			return fmt.Errorf("价格 %.2f 低于最低价 %.2f", snapshot.Price, minPrice)
		}
		return nil
	}
}

// PriceMaxFilter 最高价过滤器
func PriceMaxFilter(maxPrice float64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if snapshot.Price > maxPrice {
			return fmt.Errorf("价格 %.2f 高于最高价 %.2f", snapshot.Price, maxPrice)
		}
		return nil
	}
}

// VolumeMinFilter 最小成交量过滤器
func VolumeMinFilter(minVolume int64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if int64(snapshot.Vol) < minVolume {
			return fmt.Errorf("成交量 %d 低于最小值 %d", snapshot.Vol, minVolume)
		}
		return nil
	}
}

// ChangeRateRangeFilter 涨跌幅范围过滤器
func ChangeRateRangeFilter(minRate, maxRate float64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if snapshot.ChangeRate < minRate {
			return fmt.Errorf("涨跌幅 %.2f%% 低于最小值 %.2f%%", snapshot.ChangeRate, minRate)
		}
		if snapshot.ChangeRate > maxRate {
			return fmt.Errorf("涨跌幅 %.2f%% 高于最大值 %.2f%%", snapshot.ChangeRate, maxRate)
		}
		return nil
	}
}

// TurnoverRateMinFilter 最小换手率过滤器
func TurnoverRateMinFilter(minTurnover float64) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if snapshot.OpenTurnZ < minTurnover {
			return fmt.Errorf("换手率 %.2f 低于最小值 %.2f", snapshot.OpenTurnZ, minTurnover)
		}
		return nil
	}
}

// CodePrefixFilter 代码前缀过滤器（白名单）
// 只允许指定前缀的股票代码通过
func CodePrefixFilter(allowedPrefixes ...string) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		if len(allowedPrefixes) == 0 {
			return nil // 无限制
		}
		for _, prefix := range allowedPrefixes {
			if len(snapshot.SecurityCode) >= len(prefix) &&
				snapshot.SecurityCode[:len(prefix)] == prefix {
				return nil // 匹配成功
			}
		}
		return fmt.Errorf("代码 %s 不在允许的前缀列表中", snapshot.SecurityCode)
	}
}

// CodePrefixExcludeFilter 代码前缀排除过滤器（黑名单）
// 排除指定前缀的股票代码
func CodePrefixExcludeFilter(excludedPrefixes ...string) FilterFunc {
	return func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		for _, prefix := range excludedPrefixes {
			if len(snapshot.SecurityCode) >= len(prefix) &&
				snapshot.SecurityCode[:len(prefix)] == prefix {
				return fmt.Errorf("代码 %s 在排除的前缀列表中", snapshot.SecurityCode)
			}
		}
		return nil
	}
}

// CustomFilter 自定义过滤器构建函数
// 允许策略传入自定义的过滤逻辑
//
//	示例：
//		CustomFilter(func(p config.RuleParameter, s factors.QuoteSnapshot) error {
//			if s.Price < 10.0 {
//				return fmt.Errorf("价格太低")
//			}
//			return nil
//		})
func CustomFilter(fn FilterFunc) FilterFunc {
	return fn
}
