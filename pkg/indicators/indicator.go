package indicators

import (
	"gitee.com/quant1x/pandas"
)

// Indicator 指标接口定义
type Indicator interface {
	// Calculate 计算指标
	Calculate(df pandas.DataFrame) pandas.DataFrame
	// Name 返回指标名称
	Name() string
}

// BaseIndicator 基础指标结构体，可作为其他指标的嵌入类型
type BaseIndicator struct {
	indicatorName string
	params        map[string]interface{}
}

// NewBaseIndicator 创建基础指标实例
func NewBaseIndicator(name string, params map[string]interface{}) BaseIndicator {
	return BaseIndicator{
		indicatorName: name,
		params:        params,
	}
}

// Name 返回指标名称
func (b *BaseIndicator) Name() string {
	return b.indicatorName
}

// GetParam 获取参数
func (b *BaseIndicator) GetParam(key string, defaultValue interface{}) interface{} {
	if val, ok := b.params[key]; ok {
		return val
	}
	return defaultValue
}
