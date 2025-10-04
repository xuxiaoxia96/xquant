package update

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"xquant/pkg/cache"
	"xquant/pkg/log"
	"xquant/pkg/storages"
	"xquant/pkg/utils"
)

// UpdateParams
// - IsFullUpdate: 是否全量更新（对应 cmd 的 --all）
// - BaseDataKeywords: 基础数据更新关键词（对应 cmd 的 --base）
// - FeaturesKeywords: 特征数据更新关键词（对应 cmd 的 --features）
type UpdateParams struct {
	IsFullUpdate     bool     // 是否全量更新
	BaseDataKeywords []string // 基础数据更新关键词（如 ["stock", "finance"]）
	FeaturesKeywords []string // 特征数据更新关键词（如 ["ma", "rsi"]）
}

func RunUpdate(ctx context.Context, params UpdateParams) error {
	// 获取并校正当前可更新日期，一定不为空，仅做防御性编程
	currentDate := cache.DefaultCanUpdateDate()
	if currentDate == "" {
		err := fmt.Errorf("未获取到可更新日期，请检查系统日期配置")
		log.CtxErrorf(ctx, "[RunUpdate] %v", err) // 假设项目有日志包，ctx 用于日志关联
		return err
	}
	cacheDate, featureDate := cache.CorrectDate(currentDate)
	log.CtxInfof(ctx, "[RunUpdate] 开始更新，日期：cache=%s, feature=%s", cacheDate, featureDate)

	switch {
	case params.IsFullUpdate:
		// 分支1：全量更新
		handleFullUpdate(ctx, cacheDate, featureDate)
	case len(params.BaseDataKeywords) > 0:
		// 分支2：基础数据定向更新
		handleUpdateBaseDataWithKeywords(cacheDate, featureDate, params.BaseDataKeywords...)
	case len(params.FeaturesKeywords) > 0:
		// 分支3：特征数据定向更新
		handleUpdateFeaturesWithKeywords(cacheDate, featureDate, params.FeaturesKeywords...)
	default:
		// 分支4：无有效参数（返回错误，由调用方处理提示）
		err := fmt.Errorf("非全量更新时，必须指定基础数据关键词或特征数据关键词")
		log.CtxWarnf(ctx, "[RunUpdate] %v", err)
		return err
	}

	log.CtxInfof(ctx, "[RunUpdate] 数据更新完成")
	return nil
}

var CmdFlags = struct {
	All      bool   // --all：全量更新
	BaseData string // --base：基础数据关键词（逗号分隔）
	Features string // --features：特征数据关键词（逗号分隔）
}{}

// --------------------------
// 4. 辅助函数（cmd 参数解析、信号处理）
// --------------------------
// parseFieldKeywords 解析逗号分隔的关键词（如 "stock,finance" → ["stock", "finance"]）
func ParseFieldKeywords(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var keywords []string
	for _, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			keywords = append(keywords, item)
		}
	}
	return keywords
}

// setupCmdSignalHandler 为 cmd 配置中断信号处理（如 Ctrl+C 终止更新）
func SetupCmdSignalHandler(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigChan:
			log.CtxWarnf(ctx, "[setupCmdSignalHandler] 收到中断信号: %v，正在停止更新...", sig)
			cancel() // 触发核心逻辑的 ctx.Done()，优雅终止
		}
	}()
}

// handleFullUpdateCore 全量更新（基础数据 + 特征数据）
func handleFullUpdate(ctx context.Context, cacheDate, featureDate string) {
	log.CtxInfof(ctx, "[handleFullUpdateCore] 开始全量更新：基础数据 + 特征数据")

	// 1. 更新所有基础数据
	basePlugins := cache.Plugins(cache.PluginMaskBaseData)
	log.CtxInfof(ctx, "[handleFullUpdateCore] 基础数据插件数量: %d", len(basePlugins))
	// 更新数据
	storages.DataSetUpdate(1, featureDate, basePlugins, cache.OpUpdate)

	// 2. 更新所有特征数据
	featurePlugins := cache.Plugins(cache.PluginMaskFeature)
	if len(featurePlugins) == 0 {
		// 1. 获取全部注册的数据集插件
		mask := cache.PluginMaskFeature
		featurePlugins = cache.Plugins(mask)
	}
	log.CtxInfof(ctx, "[handleFullUpdateCore] 特征数据插件数量: %d", len(featurePlugins))
	storages.FeaturesUpdate(utils.IntPtr(1), cacheDate, featureDate, featurePlugins, cache.OpUpdate)

	log.CtxInfof(ctx, "[handleFullUpdateCore] 全量更新完成")
	fmt.Println("全量更新完成")
}

// 更新基础数据
func handleUpdateBaseDataWithKeywords(cacheDate, featureDate string, keywords ...string) {
	plugins := cache.PluginsWithName(cache.PluginMaskBaseData, keywords...)
	if len(plugins) == 0 {
		// 1. 获取全部注册的数据集插件
		mask := cache.PluginMaskBaseData
		plugins = cache.Plugins(mask)
	}
	storages.DataSetUpdate(1, featureDate, plugins, cache.OpUpdate)
	_ = cacheDate
}

// 更新特征组合
func handleUpdateFeaturesWithKeywords(cacheDate, featureDate string, keywords ...string) {
	plugins := cache.PluginsWithName(cache.PluginMaskFeature, keywords...)
	if len(plugins) == 0 {
		// 1. 获取全部注册的数据集插件
		mask := cache.PluginMaskFeature
		plugins = cache.Plugins(mask)
	}
	storages.FeaturesUpdate(utils.IntPtr(1), cacheDate, featureDate, plugins, cache.OpUpdate)
}
