package rules

import (
	"testing"

	"xquant/pkg/config"
	"xquant/pkg/factors"
)

// TestRuleInterface 测试规则接口
func TestRuleInterface(t *testing.T) {
	// 测试 BaseRule 实现 Rule 接口
	execFunc := func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		return nil
	}

	rule := NewBaseRule(KRuleBase, "测试规则", execFunc)

	// 验证接口方法
	if rule.Kind() != KRuleBase {
		t.Errorf("Kind() = %d, want %d", rule.Kind(), KRuleBase)
	}

	if rule.Name() != "测试规则" {
		t.Errorf("Name() = %s, want %s", rule.Name(), "测试规则")
	}

	if rule.Description() != "" {
		t.Errorf("Description() = %s, want empty string", rule.Description())
	}

	// 测试 Exec 方法
	param := config.RuleParameter{}
	snapshot := factors.QuoteSnapshot{}
	err := rule.Exec(param, snapshot)
	if err != nil {
		t.Errorf("Exec() returned error: %v", err)
	}
}

// TestRegisterFunc 测试向后兼容的 RegisterFunc
func TestRegisterFunc(t *testing.T) {
	// 清理之前的注册
	// 注意：在实际测试中，可能需要重置 mapRules

	execFunc := func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		return nil
	}

	// 测试注册函数
	err := RegisterFunc(999, "测试规则999", execFunc)
	if err != nil {
		t.Fatalf("RegisterFunc() failed: %v", err)
	}

	// 验证规则已注册
	rule, err := GetRule(999)
	if err != nil {
		t.Fatalf("GetRule() failed: %v", err)
	}

	if rule.Name() != "测试规则999" {
		t.Errorf("GetRule() returned rule with name %s, want %s", rule.Name(), "测试规则999")
	}
}

// TestRegister 测试新的 Register 方法
func TestRegister(t *testing.T) {
	execFunc := func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		return nil
	}

	rule := NewBaseRuleWithDescription(998, "测试规则998", "这是一个测试规则", execFunc)

	// 测试注册接口
	err := Register(rule)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// 验证规则已注册
	retrievedRule, err := GetRule(998)
	if err != nil {
		t.Fatalf("GetRule() failed: %v", err)
	}

	if retrievedRule.Name() != "测试规则998" {
		t.Errorf("GetRule() returned rule with name %s, want %s", retrievedRule.Name(), "测试规则998")
	}

	if retrievedRule.Description() != "这是一个测试规则" {
		t.Errorf("GetRule() returned rule with description %s, want %s", retrievedRule.Description(), "这是一个测试规则")
	}
}

// TestGetAllRules 测试获取所有规则
func TestGetAllRules(t *testing.T) {
	rules := GetAllRules()
	if len(rules) == 0 {
		t.Log("No rules registered (this is expected if running in isolation)")
		return
	}

	// 验证返回的规则都是有效的
	for _, rule := range rules {
		if rule == nil {
			t.Error("GetAllRules() returned nil rule")
		}
		if rule.Kind() == 0 && rule.Name() == "" {
			t.Error("GetAllRules() returned invalid rule")
		}
	}
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	// 验证现有的 RegisterFunc 仍然可以工作
	execFunc := func(ruleParameter config.RuleParameter, snapshot factors.QuoteSnapshot) error {
		return nil
	}

	// 使用旧的 RegisterFunc 方式注册
	err := RegisterFunc(997, "兼容性测试", execFunc)
	if err != nil {
		t.Fatalf("RegisterFunc() failed (backward compatibility test): %v", err)
	}

	// 验证可以通过接口方式访问
	rule, err := GetRule(997)
	if err != nil {
		t.Fatalf("GetRule() failed: %v", err)
	}

	// 验证是 BaseRule 类型
	if _, ok := rule.(*BaseRule); !ok {
		t.Error("RegisterFunc() should create BaseRule instance")
	}
}
