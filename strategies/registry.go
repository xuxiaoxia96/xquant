package strategies

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"gitee.com/quant1x/gox/api"
)

var (
	// _mutexStrategies 保护策略注册表的互斥锁
	_mutexStrategies sync.Mutex
	// _mapStrategies 策略注册表，key为策略编码(ModelKind)，value为策略实例
	_mapStrategies = map[ModelKind]Strategy{}
	// _mapStrategiesOverwrite 强制覆盖标记表，记录哪些策略编码已被强制覆盖
	// 当策略编码 >= ModelForceOverwrite 时，允许覆盖已存在的策略
	_mapStrategiesOverwrite = map[ModelKind]bool{}
	// ErrAlreadyExists 策略已存在的错误
	// 当尝试注册一个已存在的策略编码（且未使用强制覆盖标志）时返回
	ErrAlreadyExists = errors.New("the strategy already exists")
	// ErrNotFound 策略未找到的错误
	// 当尝试获取一个不存在的策略编码时返回
	ErrNotFound = errors.New("the strategy was not found")
)

// Register 注册策略到策略注册表
//
// 参数：
//   - strategy: 要注册的策略实例，必须实现 Strategy 接口
//
// 返回值：
//   - error: 注册失败时返回错误，成功时返回 nil
//   - ErrAlreadyExists: 策略编码已存在且未使用强制覆盖标志
//
// 注册规则：
//  1. 如果策略编码 < ModelForceOverwrite (0x80000000)：
//     - 检查策略编码是否已存在，如果存在则返回 ErrAlreadyExists
//     - 如果不存在，则注册策略
//  2. 如果策略编码 >= ModelForceOverwrite：
//     - 自动去除强制覆盖标志位（strategyCode &^ ModelForceOverwrite）
//     - 标记该策略编码已被强制覆盖
//     - 允许覆盖已存在的策略（如果存在）
//  3. 如果策略编码已在强制覆盖标记表中，直接返回 nil（避免重复处理）
//
// 线程安全：
//   - 使用互斥锁保证并发安全
//
// 示例：
//
//	err := strategies.Register(ModelNo1{})
//	if err != nil {
//	    log.Fatal(err)
//	}
func Register(strategy Strategy) error {
	_mutexStrategies.Lock()
	defer _mutexStrategies.Unlock()
	strategyCode := strategy.Code()
	// 检查是否存在覆盖策略的情况
	_, overwritten := _mapStrategiesOverwrite[strategyCode]
	if overwritten {
		return nil
	}
	if strategyCode < ModelForceOverwrite {
		_, ok := _mapStrategies[strategyCode]
		if ok {
			return ErrAlreadyExists
		}
	} else {
		strategyCode = strategyCode &^ ModelForceOverwrite
		_mapStrategiesOverwrite[strategyCode] = true
	}
	_mapStrategies[strategyCode] = strategy
	return nil
}

// CheckoutStrategy 根据策略编码获取策略实例
//
// 参数：
//   - strategyNumber: 策略编码（ModelKind），如 ModelHousNo1 = 1
//
// 返回值：
//   - Strategy: 找到的策略实例，未找到时返回 nil
//   - error: 未找到策略时返回 ErrNotFound，找到时返回 nil
//
// 线程安全：
//   - 使用互斥锁保证并发安全
//
// 使用场景：
//   - 在策略执行前根据策略编码获取对应的策略实例
//   - 用于命令行参数解析，根据用户输入的策略编号获取策略
//
// 示例：
//
//	strategy, err := strategies.CheckoutStrategy(1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tracker.ExecuteStrategy(strategy, &barIndex)
//
// 注意：
//   - 函数名 "Checkout" 在版本控制系统中常见，但这里实际含义是"获取"或"查找"
//   - 可以考虑重命名为 GetStrategy 或 FindStrategy 以提高可读性
func CheckoutStrategy(strategyNumber uint64) (Strategy, error) {
	_mutexStrategies.Lock()
	defer _mutexStrategies.Unlock()
	strategy, ok := _mapStrategies[strategyNumber]
	if ok {
		return strategy, nil
	}

	return nil, ErrNotFound
}

// UsageStrategyList 生成策略列表的用法说明字符串
//
// 返回值：
//   - string: 格式化的策略列表字符串，每行格式为 "策略编码: 策略名称\n"
//     策略按照编码（ModelKind）升序排序
//
// 使用场景：
//   - 用于命令行工具的帮助信息，显示所有可用的策略列表
//   - 通常作为 flag 的 Usage 参数，在用户输入错误时显示可用选项
//
// 输出格式示例：
//
//	"1: 1号策略\n"
//	"2: 2号策略\n"
//	"89: 89K策略\n"
//
// 示例：
//
//	engineCmd.Flags().Uint64Var(&strategyNumber, "strategy",
//	    strategies.DefaultStrategy, strategies.UsageStrategyList())
//
// 注意：
//   - 函数名 "Usage" 通常指"用法说明"，但这里实际是"列表"
//   - 可以考虑重命名为 ListStrategies 或 GetStrategyList 以提高可读性
//   - 内部变量名 "rule" 已改为 "strategy" 以保持命名一致性
func UsageStrategyList() string {
	// 策略按照编码排序
	kinds := api.Keys(_mapStrategies)
	slices.Sort(kinds)
	usage := ""
	for _, kind := range kinds {
		if strategy, ok := _mapStrategies[kind]; ok {
			usage += fmt.Sprintf("%d: %s\n", kind, strategy.Name())
		}
	}
	return usage
}
