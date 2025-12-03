package cmd

import (
	"fmt"
	"os"
	"strings"

	goruntime "runtime"

	"gitee.com/quant1x/gox/runtime"
	"gitee.com/quant1x/num"
	"github.com/klauspost/cpuid/v2"
	cli "github.com/spf13/cobra"

	"xquant/pkg/log"
	"xquant/pkg/models"
	"xquant/pkg/tracker"
)

// AppConfig 应用配置集中管理
type AppConfig struct {
	Application    string
	MinVersion     string
	StrategyNumber uint64
	BusinessDebug  bool
	CpuAvx2        bool
	CpuNum         int
}

// 全局配置实例
var cfg = &AppConfig{
	Application:    runtime.ApplicationName(),
	MinVersion:     "0.0.1",
	StrategyNumber: models.DefaultStrategy,
	BusinessDebug:  runtime.Debug(),
	CpuNum:         goruntime.NumCPU() / 2,
}

// UpdateApplicationName 更新应用名称
func UpdateApplicationName(name string) {
	cfg.Application = name
}

// UpdateApplicationVersion 更新版本号
func UpdateApplicationVersion(v string) {
	cfg.MinVersion = v
}

// 初始化所有子命令
func initSubCommands() {
	// 初始化命令（但不添加到根命令，在 GlobalFlags 中添加）
	// InitUpdateCmd()
	//InitServerCmd()

	// 其他子命令可以在这里初始化
	// initPrint()
	// initRepair()
	// ...
}

// 打印系统信息
func printSystemInfo() {
	fmt.Printf("CPU: %s, %dCores, AVX2: %t\n",
		cpuid.CPU.BrandName,
		goruntime.NumCPU(),
		num.GetAvx2Enabled())
	fmt.Println()
}

// 应用初始化配置
func setupApplication() {
	runtime.SetDebug(cfg.BusinessDebug)
	num.SetAvx2Enabled(cfg.CpuAvx2)
	goruntime.GOMAXPROCS(cfg.CpuNum)
}

// 执行默认策略
func executeDefaultStrategy() {
	model, err := models.CheckoutStrategy(cfg.StrategyNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	printSystemInfo()
	barIndex := 1
	tracker.ExecuteStrategy(model, &barIndex)
}

// GlobalFlags 创建主命令
func GlobalFlags() *cli.Command {
	initSubCommands()

	rootCmd := &cli.Command{
		Use: cfg.Application,
		Run: func(cmd *cli.Command, args []string) {
			log.Warnf("stock default args:%+v", os.Args)
			executeDefaultStrategy()
		},
		PersistentPreRun: func(cmd *cli.Command, args []string) {
			setupApplication()
		},
	}

	// 注册全局标志
	rootCmd.Flags().Uint64Var(&cfg.StrategyNumber, "strategy",
		cfg.StrategyNumber, models.UsageStrategyList())
	rootCmd.Flags().IntVar(&models.CountDays, "count", 0, "统计多少天")
	rootCmd.Flags().IntVar(&models.CountTopN, "top",
		models.AllStockTopN(), "输出前排几名")

	// 注册持久化标志
	rootCmd.PersistentFlags().BoolVar(&cfg.BusinessDebug, "debug",
		cfg.BusinessDebug, "打开业务调试开关, 慎重使用!")
	rootCmd.PersistentFlags().BoolVar(&cfg.CpuAvx2, "avx2",
		false, "Avx2 加速开关")
	rootCmd.PersistentFlags().IntVar(&cfg.CpuNum, "cpu",
		cfg.CpuNum, "设置CPU最大核数")

	// 添加子命令
	rootCmd.AddCommand(InitUpdateCmd())
	// rootCmd.AddCommand(cmdBackTest)

	return rootCmd
}

// 解析标志错误信息
func parseFlagError(err error) (flag, value string) {
	if err == nil {
		return
	}

	before, _, ok := strings.Cut(err.Error(), "flag:")
	if !ok {
		return
	}

	arr := strings.Split(strings.TrimSpace(before), "\"")
	if len(arr) != 5 {
		return
	}

	return strings.TrimSpace(arr[3]), strings.TrimSpace(arr[1])
}
