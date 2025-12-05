package rules

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/runtime"
	bitmap "github.com/bits-and-blooms/bitset"

	"xquant/pkg/config"
	"xquant/pkg/factors"
)

// Kind 规则类型
type Kind = uint

const (
	Pass Kind = 0
)

const (
	engineBaseRule Kind = 1
	KRuleF10            = engineBaseRule + 0 // 基本面规则
	KRuleBase           = engineBaseRule + 1 // 基础规则
)

// 规则错误码, 每一组规则错误拟1000个错误码
const (
	errorRuleF10  = (iota + 1) * 1000 // F10错误码
	errorRuleBase                     // 基础规则错误码
)

// ============================================
// 核心接口定义
// ============================================

// Rule 规则接口
// 所有规则必须实现此接口
type Rule interface {
	// Kind 返回规则类型
	Kind() Kind

	// Name 返回规则名称
	Name() string

	// Description 返回规则描述（可选，用于文档和调试）
	Description() string

	// Exec 执行规则检查
	// 返回 nil 表示通过，返回 error 表示不通过
	Exec(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error
}

// ============================================
// 基础实现类（适配器模式）
// ============================================

// BaseRule 基础规则实现
// 用于包装函数类型的规则，提供接口实现
type BaseRule struct {
	kind        Kind
	name        string
	description string
	exec        func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error
}

// NewBaseRule 创建基础规则
func NewBaseRule(kind Kind, name string, exec func(config.RuleParameter, factors.QuoteSnapshot) error) *BaseRule {
	return &BaseRule{
		kind:        kind,
		name:        name,
		description: "", // 默认空描述
		exec:        exec,
	}
}

// NewBaseRuleWithDescription 创建带描述的基础规则
func NewBaseRuleWithDescription(kind Kind, name, description string, exec func(config.RuleParameter, factors.QuoteSnapshot) error) *BaseRule {
	return &BaseRule{
		kind:        kind,
		name:        name,
		description: description,
		exec:        exec,
	}
}

// Kind 实现 Rule 接口
func (r *BaseRule) Kind() Kind {
	return r.kind
}

// Name 实现 Rule 接口
func (r *BaseRule) Name() string {
	return r.name
}

// Description 实现 Rule 接口
func (r *BaseRule) Description() string {
	return r.description
}

// Exec 实现 Rule 接口
func (r *BaseRule) Exec(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
	if r.exec == nil {
		return errors.New("rule executor is nil")
	}
	return r.exec(ruleParameter, snapshot)
}

// ============================================
// 规则注册表
// ============================================

var (
	mutex    sync.RWMutex
	mapRules = map[Kind]Rule{} // 改为存储 Rule 接口
)

var (
	ErrAlreadyExists = errors.New("the rule already exists") // 规则已经存在
	ErrNotFound      = errors.New("the rule was not found")  // 规则不存在
)

// ============================================
// 注册函数（向后兼容 + 新接口）
// ============================================

// Register 注册规则接口实现（新方法，推荐使用）
func Register(rule Rule) error {
	if rule == nil {
		return errors.New("rule cannot be nil")
	}

	mutex.Lock()
	defer mutex.Unlock()

	kind := rule.Kind()
	_, ok := mapRules[kind]
	if ok {
		return ErrAlreadyExists
	}

	mapRules[kind] = rule
	return nil
}

// RegisterFunc 注册规则回调函数（向后兼容方法）
// 为了保持向后兼容，保留此函数
// 内部使用 BaseRule 适配器包装函数
func RegisterFunc(kind Kind, name string, cb func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error) error {
	if cb == nil {
		return errors.New("rule callback cannot be nil")
	}

	// 使用适配器模式，将函数包装为 Rule 接口
	rule := NewBaseRule(kind, name, cb)
	return Register(rule)
}

// ============================================
// 规则执行
// ============================================

// Filter 遍历所有规则并执行
// 返回：通过的规则列表、失败的规则类型、错误信息
func Filter(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) (passed []uint64, failed Kind, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if len(mapRules) == 0 {
		return
	}

	var bitset bitmap.BitSet

	// 规则按照kind排序
	kinds := api.Keys(mapRules)
	slices.Sort(kinds)

	// 遍历执行规则
	for _, kind := range kinds {
		rule, ok := mapRules[kind]
		if !ok {
			continue
		}

		// 检查是否忽略此规则组
		if slices.Contains(ruleParameter.IgnoreRuleGroup, int(rule.Kind())) {
			continue
		}

		// 执行规则
		err = rule.Exec(ruleParameter, snapshot)
		if err != nil {
			failed = rule.Kind()
			break // 短路模式：遇到错误立即停止
		}

		// 记录通过的规则
		bitset.Set(rule.Kind())
	}

	return bitset.Bytes(), failed, err
}

// ============================================
// 规则查询和管理
// ============================================

// GetRule 获取指定类型的规则
func GetRule(kind Kind) (Rule, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	rule, ok := mapRules[kind]
	if !ok {
		return nil, ErrNotFound
	}

	return rule, nil
}

// GetAllRules 获取所有已注册的规则
func GetAllRules() []Rule {
	mutex.RLock()
	defer mutex.RUnlock()

	rules := make([]Rule, 0, len(mapRules))
	kinds := api.Keys(mapRules)
	slices.Sort(kinds)

	for _, kind := range kinds {
		if rule, ok := mapRules[kind]; ok {
			rules = append(rules, rule)
		}
	}

	return rules
}

// PrintRuleList 输出规则列表（增强版）
func PrintRuleList() {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("规则总数:", len(mapRules))

	kinds := api.Keys(mapRules)
	slices.Sort(kinds)

	for _, kind := range kinds {
		rule, ok := mapRules[kind]
		if !ok {
			continue
		}

		desc := rule.Description()
		if desc == "" {
			desc = "(无描述)"
		}

		// 尝试获取函数名（如果是 BaseRule）
		funcName := "N/A"
		if baseRule, ok := rule.(*BaseRule); ok && baseRule.exec != nil {
			funcName = runtime.FuncName(baseRule.exec)
		}

		fmt.Printf("kind: %d, name: %s, desc: %s, method: %s\n",
			rule.Kind(), rule.Name(), desc, funcName)
	}
}
