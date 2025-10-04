package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cmder "github.com/spf13/cobra"

	updateservice "xquant/biz/service/update"
	"xquant/pkg/log"
)

const (
	updateCommand     = "update"
	updateDescription = "更新数据"
)

var CmdFlags = struct {
	All      bool   // --all：全量更新
	Base     string // --base：基础数据关键词（逗号分隔，如 "stock,finance"）
	Features string // --features：特征数据关键词（逗号分隔，如 "ma,rsi"）
}{}

// InitUpdateCmd 初始化 cmd 更新命令（对外暴露，供 cmd 根命令注册）
func InitUpdateCmd() *cmder.Command {
	cmd := &cmder.Command{
		Use:     "update",
		Short:   "数据更新命令",
		Long:    "更新系统基础数据（如行情、财务）和特征数据（如技术指标），支持全量或定向更新",
		Example: "xquant update --all\nxquant update --base=stock,finance\nxquant update --features=ma,rsi",
		Run:     runUpdateCmd, // cmd 执行入口
	}

	// 注册 cmd 命令行参数
	cmd.Flags().BoolVar(&CmdFlags.All, "all", false, "全量更新所有基础数据和特征数据")
	cmd.Flags().StringVar(&CmdFlags.Base, "base", "", "基础数据定向更新（关键词逗号分隔，如：--base=stock,finance）")
	cmd.Flags().StringVar(&CmdFlags.Features, "features", "", "特征数据定向更新（关键词逗号分隔，如：--features=ma,rsi）")

	return cmd
}

// runUpdateCmd 原 cmd 的 Run 函数（仅负责参数转换和调用核心逻辑）
func runUpdateCmd(cmd *cmder.Command, args []string) {
	// 步骤1：将 cmd 字符串参数转换为 UpdateParams 结构体
	params := updateservice.UpdateParams{
		IsFullUpdate:     CmdFlags.All,
		BaseDataKeywords: updateservice.ParseFieldKeywords(CmdFlags.Base),     // 解析逗号分隔的关键词
		FeaturesKeywords: updateservice.ParseFieldKeywords(CmdFlags.Features), // 解析逗号分隔的关键词
	}

	// 步骤2：创建 cmd 上下文（用于中断信号，如 Ctrl+C）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 监听 cmd 中断信号（可选，增强用户体验）
	setupCmdSignalHandler(ctx, cancel)

	// 步骤3：调用核心更新逻辑（与 HTTP 共用）
	if err := updateservice.RunUpdate(ctx, params); err != nil {
		fmt.Printf("更新失败: %v\n", err)
		err := cmd.Usage()
		if err != nil {
			return
		} // 仅 cmd 场景显示帮助
		return
	}
	fmt.Println("数据更新完成")
}

// setupCmdSignalHandler cmd 中断信号处理（Ctrl+C 触发取消）
func setupCmdSignalHandler(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigChan:
			log.CtxWarnf(ctx, "[setupCmdSignalHandler] 收到中断信号: %v，正在停止更新...", sig)
			cancel() // 触发核心逻辑 ctx.Done()，优雅终止
		}
	}()
}
