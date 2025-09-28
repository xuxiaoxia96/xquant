package services

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3" // 官方成熟定时库

	"xquant/pkg/config"
	"xquant/pkg/log"
)

// Task 定时任务结构体
type Task struct {
	Name     string       // 任务名称（唯一标识）
	Spec     string       // 触发规则（Cron表达式或固定间隔）
	Callback func()       // 任务执行函数
	EntryID  cron.EntryID // 任务在cron中的唯一ID（用于后续管理）
}

var (
	ErrAlreadyExists = errors.New("the job already exists") // 任务已存在
	ErrForbidden     = errors.New("the job was forbidden")  // 任务被禁止

	jobMutex sync.RWMutex             // 读写锁（保证任务注册/查询的并发安全）
	taskMap  = make(map[string]*Task) // 任务注册表（name -> Task）
	cronObj  *cron.Cron               // cron核心实例（全局唯一）
)

// 初始化cron实例（支持秒级调度）
func init() {
	// 创建cron实例，启用秒级调度（默认cron是5位表达式，加 WithSeconds() 支持6位）
	// 可选：添加时区（如 WithLocation(time.UTC)），默认使用本地时区
	cronObj = cron.New(cron.WithSeconds(), cron.WithChain(
		cron.DelayIfStillRunning(cron.DefaultLogger), // 核心：如果任务未执行完，延迟执行（避免并发执行同一任务）
	))
}

// Register 注册定时任务（替换原自定义Register函数）
// 参数：name-任务名，spec-触发规则，callback-执行函数
func Register(name, spec string, callback func()) error {
	if callback == nil {
		return errors.New("task callback cannot be nil")
	}

	jobMutex.Lock()
	defer jobMutex.Unlock()

	// 1. 检查任务是否已存在
	if _, exists := taskMap[name]; exists {
		return ErrAlreadyExists
	}

	// 2. 从配置读取任务开关和自定义触发规则（保留原配置逻辑）
	enable := true
	if jobParam := config.GetJobParameter(name); jobParam != nil {
		enable = jobParam.Enable
		// 若配置了自定义触发规则，覆盖默认spec
		if trigger := strings.TrimSpace(jobParam.Trigger); len(trigger) > 0 {
			spec = trigger
		}
	}
	// 若任务被禁用，直接返回（不报错）
	if !enable {
		log.Infof("task [%s] is disabled by config", name)
		return nil
	}

	// 3. 验证Cron表达式合法性（提前校验，避免启动时失败）
	// 4. 暂存任务（此时不添加到cron，避免提前执行；在DaemonService中统一启动）
	taskMap[name] = &Task{
		Name:     name,
		Spec:     spec,
		Callback: callback,
	}

	log.Infof("task [%s] registered successfully, spec: [%s]", name, spec)
	return nil
}

// DaemonService 守护进程入口（替换原自定义DaemonService）
func DaemonService() {
	jobMutex.RLock()
	defer jobMutex.RUnlock()

	// 1. 启动cron调度器
	log.Infof("starting cron scheduler...")
	cronObj.Start()
	log.Infof("cron scheduler started successfully")

	// 2. 批量添加任务到cron（并记录EntryID）
	log.Infof("registering %d tasks to cron...", len(taskMap))
	for _, task := range taskMap {
		// 添加任务到cron：使用AddFunc，返回唯一EntryID
		entryID, err := cronObj.AddFunc(task.Spec, task.Callback)
		if err != nil {
			log.Errorf("failed to add task [%s] to cron: %v", task.Name, err)
			continue
		}
		// 记录任务的EntryID（后续可用于暂停/删除任务）
		task.EntryID = entryID
		log.Infof("task [%s] added to cron successfully, entryID: %d, spec: [%s]",
			task.Name, entryID, task.Spec)
	}

	// 3. 等待程序退出信号（替换原coroutine.WaitForShutdown，使用标准库实现）
	log.Infof("all tasks started, waiting for shutdown signal...")
	waitForShutdown()

	// 4. 优雅关闭cron（等待正在执行的任务完成）
	log.Infof("shutting down cron scheduler...")
	ctx := cronObj.Stop()
	select {
	case <-ctx.Done():
		log.Infof("all running tasks finished, cron scheduler stopped")
	case <-time.After(30 * time.Second): // 超时保护：30秒后强制退出
		log.Warnf("cron scheduler shutdown timed out (30s), forcing exit")
	}
}

// PrintJobList 输出所有已注册任务列表
// PrintJobList 输出所有已注册任务列表（不依赖自定义runtime库）
func PrintJobList() {
	jobMutex.RLock()
	defer jobMutex.RUnlock()

	// 打印表头
	fmt.Printf("\n%s\n", strings.Repeat("-", 100))
	fmt.Printf("%-25s %-25s %-50s\n", "Task Name", "Cron Spec", "Callback Function")
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	// 遍历任务列表，原生方式获取函数名
	for _, task := range taskMap {
		// 核心：通过标准库runtime获取回调函数的函数名
		funcName := getFunctionName(task.Callback)
		// 格式化输出（对齐字段）
		fmt.Printf("%-25s %-25s %-50s\n", task.Name, task.Spec, funcName)
	}

	// 打印统计信息
	fmt.Printf("%s\n", strings.Repeat("-", 100))
	fmt.Printf("Total registered tasks: %d\n", len(taskMap))
}

// getFunctionName 原生获取函数名的工具函数（不依赖任何外部库）
// 参数：f - 任意函数类型（如func()）
// 返回：函数的完整路径名（如"your_project/services.jobUpdateSnapshot"）
func getFunctionName(f func()) string {
	if f == nil {
		return "nil (callback not set)"
	}

	// 1. runtime.FuncForPC：通过函数指针（PC值）获取Func对象
	// - reflect.ValueOf(f).Pointer()：获取函数的PC值（函数在内存中的地址）
	// - false：表示不获取"包装函数"（如闭包的外层函数）
	funcObj := runtime.FuncForPC(reflect.ValueOf(f).Pointer())
	if funcObj == nil {
		return "unknown function (failed to get Func object)"
	}

	// 2. funcObj.Name()：获取函数的完整路径名（格式：包路径.函数名）
	// 示例："gitee.com/quant1x/your_project/services.jobUpdateSnapshot"
	fullName := funcObj.Name()

	// （可选）简化函数名：只保留"包名.函数名"，去掉前面的完整路径（按需选择）
	// 例如：将"gitee.com/quant1x/your_project/services.jobUpdateSnapshot"简化为"services.jobUpdateSnapshot"
	// if lastDotIdx := strings.LastIndex(fullName, "/"); lastDotIdx != -1 {
	// 	fullName = fullName[lastDotIdx+1:]
	// }

	return fullName
}

// waitForShutdown 等待程序退出信号（支持Ctrl+C、kill等信号）
func waitForShutdown() {
	// 创建信号通道（监听中断信号）
	sigChan := make(chan struct{}, 1)
	// 注册信号处理函数
	go func() {
		// 监听系统中断信号（Ctrl+C 或 kill -2）
		// 需导入 "os" 和 "os/signal" 包

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM) // 监听中断和终止信号
		<-sig                                               // 阻塞直到收到信号
		log.Infof("received shutdown signal, preparing to exit...")
		close(sigChan)
	}()
	// 阻塞直到收到退出信号
	<-sigChan
}
