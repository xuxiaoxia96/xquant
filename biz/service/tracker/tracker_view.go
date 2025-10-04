package update

import (
	"context"
	"fmt"
	_ "os"
	"strings"
	_ "syscall"

	_ "github.com/spf13/cobra"

	"xquant/pkg/cache"
	"xquant/pkg/log"
	"xquant/pkg/storages"
)

// --------------------------
// 1. 核心参数结构体（对齐 TrackerCoreParams 风格）
// UpdateCoreParams 数据更新的核心参数，支持 cmd/HTTP 共用
type UpdateCoreParams struct {
	IsFullUpdate     bool     // 是否全量更新（对应 cmd --all，HTTP 请求 is_full 字段）
	BaseKeywords     []string // 基础数据更新关键词（对应 cmd --base，HTTP base_keywords 字段）
	FeaturesKeywords []string // 特征数据更新关键词（对应 cmd --features，HTTP features_keywords 字段）
}

// 全局进度条索引（若仅更新逻辑使用，可后续改为参数传入，当前保持兼容）
var barIndex = 1

// --------------------------
// 2. 核心更新函数（对齐 RunTrackerCore 风格）
// RunUpdateCore 数据更新核心逻辑，支持 cmd/HTTP 调用
// ctx: 上下文（用于日志关联、中断信号传递，如 HTTP 请求取消、cmd Ctrl+C）
// params: 更新核心参数（统一 cmd/HTTP 入参格式）
func RunUpdateCore(ctx context.Context, params UpdateCoreParams) {
	// 步骤1：校验参数合法性（提前拦截无效参数，减少后续无效执行）
	if err := validateCoreParams(params); err != nil {
		log.CtxErrorf(ctx, "[RunUpdateCore] 参数校验失败: %v", err)
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 步骤2：获取并校正可更新日期（依赖 cache 包，逻辑不变）
	currentDate := cache.DefaultCanUpdateDate()
	if currentDate == "" {
		errMsg := "未获取到可更新日期，请检查系统日期配置"
		log.CtxErrorf(ctx, "[RunUpdateCore] %s", errMsg)
		fmt.Printf("错误: %s\n", errMsg)
		return
	}
	cacheDate, featureDate := cache.CorrectDate(currentDate)
	log.CtxInfof(ctx, "[RunUpdateCore] 开始更新，日期：cache=%s, feature=%s", cacheDate, featureDate)

	// 步骤3：按参数分支执行更新（核心业务逻辑）
	switch {
	case params.IsFullUpdate:
		handleFullUpdateCore(ctx, cacheDate, featureDate)
	case len(params.BaseKeywords) > 0:
		handleBaseUpdateCore(ctx, cacheDate, featureDate, params.BaseKeywords...)
	case len(params.FeaturesKeywords) > 0:
		handleFeaturesUpdateCore(ctx, cacheDate, featureDate, params.FeaturesKeywords...)
	}
}

// --------------------------
// 3. 参数校验辅助函数（提前拦截无效场景）
func validateCoreParams(params UpdateCoreParams) error {
	// 校验规则：全量更新 和 定向更新（基础/特征）二选一，且定向更新关键词非空
	if !params.IsFullUpdate {
		if len(params.BaseKeywords) == 0 && len(params.FeaturesKeywords) == 0 {
			return fmt.Errorf("非全量更新时，必须指定基础数据关键词（BaseKeywords）或特征数据关键词（FeaturesKeywords）")
		}
		// 可选：校验关键词去重（避免重复更新同一插件）
		params.BaseKeywords = removeDuplicateKeywords(params.BaseKeywords)
		params.FeaturesKeywords = removeDuplicateKeywords(params.FeaturesKeywords)
	}
	return nil
}

// 关键词去重（辅助函数）
func removeDuplicateKeywords(keywords []string) []string {
	seen := make(map[string]struct{}, len(keywords))
	unique := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if _, ok := seen[kw]; !ok {
			seen[kw] = struct{}{}
			unique = append(unique, kw)
		}
	}
	return unique
}

// --------------------------
// 4. 具体更新实现（对齐核心函数命名风格）
// handleFullUpdateCore 全量更新（基础数据 + 特征数据）
func handleFullUpdateCore(ctx context.Context, cacheDate, featureDate string) {
	log.CtxInfof(ctx, "[handleFullUpdateCore] 开始全量更新：基础数据 + 特征数据")

	// 1. 更新所有基础数据
	basePlugins := cache.Plugins(cache.PluginMaskBaseData)
	log.CtxInfof(ctx, "[handleFullUpdateCore] 基础数据插件数量: %d", len(basePlugins))
	storages.DataSetUpdate(barIndex, featureDate, basePlugins, cache.OpUpdate)

	// 2. 更新所有特征数据
	featurePlugins := cache.Plugins(cache.PluginMaskFeature)
	log.CtxInfof(ctx, "[handleFullUpdateCore] 特征数据插件数量: %d", len(featurePlugins))
	if err := storages.FeaturesUpdate(&barIndex, cacheDate, featureDate, featurePlugins, cache.OpUpdate); err != nil {
		log.CtxErrorf(ctx, "[handleFullUpdateCore] 特征数据全量更新失败: %v", err)
		fmt.Printf("警告: 特征数据全量更新失败: %v\n", err)
		return
	}

	log.CtxInfof(ctx, "[handleFullUpdateCore] 全量更新完成")
	fmt.Println("全量更新完成")
}

// handleBaseUpdateCore 基础数据定向更新（按关键词）
func handleBaseUpdateCore(ctx context.Context, cacheDate, featureDate string, keywords ...string) {
	// 按关键词筛选插件，无匹配则提示并退出（避免默认全量，减少意外更新）
	plugins := cache.PluginsWithName(cache.PluginMaskBaseData, keywords...)
	if len(plugins) == 0 {
		errMsg := fmt.Sprintf("无匹配关键词【%v】的基础数据插件，请检查关键词是否正确", keywords)
		log.CtxWarnf(ctx, "[handleBaseUpdateCore] %s", errMsg)
		fmt.Printf("警告: %s\n", errMsg)
		return
	}

	log.CtxInfof(ctx, "[handleBaseUpdateCore] 开始基础数据定向更新，关键词：%v，插件数量：%d", keywords, len(plugins))
	// 执行更新（忽略 cacheDate，明确标注）
	_ = cacheDate
	storages.DataSetUpdate(barIndex, featureDate, plugins, cache.OpUpdate)

	log.CtxInfof(ctx, "[handleBaseUpdateCore] 基础数据定向更新完成")
	fmt.Println("基础数据定向更新完成")
}

// handleFeaturesUpdateCore 特征数据定向更新（按关键词）
func handleFeaturesUpdateCore(ctx context.Context, cacheDate, featureDate string, keywords ...string) {
	// 按关键词筛选插件，无匹配则提示并退出
	plugins := cache.PluginsWithName(cache.PluginMaskFeature, keywords...)
	if len(plugins) == 0 {
		errMsg := fmt.Sprintf("无匹配关键词【%v】的特征数据插件，请检查关键词是否正确", keywords)
		log.CtxWarnf(ctx, "[handleFeaturesUpdateCore] %s", errMsg)
		fmt.Printf("警告: %s\n", errMsg)
		return
	}

	log.CtxInfof(ctx, "[handleFeaturesUpdateCore] 开始特征数据定向更新，关键词：%v，插件数量：%d", keywords, len(plugins))
	// 执行更新
	if err := storages.FeaturesUpdate(&barIndex, cacheDate, featureDate, plugins, cache.OpUpdate); err != nil {
		log.CtxErrorf(ctx, "[handleFeaturesUpdateCore] 特征数据定向更新失败: %v", err)
		fmt.Printf("警告: 特征数据定向更新失败: %v\n", err)
		return
	}

	log.CtxInfof(ctx, "[handleFeaturesUpdateCore] 特征数据定向更新完成")
	fmt.Println("特征数据定向更新完成")
}
