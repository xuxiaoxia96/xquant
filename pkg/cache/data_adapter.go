package cache

import (
	"errors"
	"slices"
	"sync"
)

type Kind = uint64

const (
	PluginMaskBaseData Kind = 0x1000000000000000 // 基础数据
	PluginMaskFeature  Kind = 0x2000000000000000 // 特征数据
	PluginMaskStrategy Kind = 0x3000000000000000 // 策略
)

const (
	// DefaultDataProvider 默认数据提供者
	DefaultDataProvider = "engine"
)

// DataAdapter 数据插件
type DataAdapter interface {
	// Schema 继承基础特性接口
	Schema
	// Print 控制台输出指定日期的数据
	Print(code string, date ...string)
}

// Handover 缓存切换接口
type Handover interface {
	// ChangingOverDate 缓存数据转换日期
	//	数据集等基础数据不需要切换日期
	ChangingOverDate(date string)
}

type Depend interface {
	DependOn() []Kind
}

var (
	ErrAlreadyExists = errors.New("the plugin already exists")
)

var (
	pluginMutex    sync.Mutex
	mapDataPlugins = map[Kind]DataAdapter{}
)

// Register 注册插件
func Register(plugin DataAdapter) error {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()
	_, ok := mapDataPlugins[plugin.Kind()]
	if ok {
		return ErrAlreadyExists
	}
	mapDataPlugins[plugin.Kind()] = plugin
	return nil
}

// GetDataAdapter 获取数据适配器
func GetDataAdapter(kind Kind) DataAdapter {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()
	adapter, ok := mapDataPlugins[kind]
	if ok {
		return adapter
	}
	return nil
}

// Plugins 按照类型标志位筛选数据插件（优化内存分配）
func Plugins(mask ...Kind) (list []DataAdapter) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()

	pluginType := Kind(0)
	if len(mask) > 0 {
		if mask[0] == PluginMaskBaseData || mask[0] == PluginMaskFeature {
			pluginType = mask[0]
		}
	}

	kinds := filterKindsByType(pluginType, mapDataPlugins)
	return kindsToPlugins(kinds, mapDataPlugins)
}

func PluginsWithName(pluginType Kind, keywords ...string) (list []DataAdapter) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()

	if len(keywords) == 0 {
		return
	}

	kinds := filterKindsByTypeAndKeywords(pluginType, keywords, mapDataPlugins)
	return kindsToPlugins(kinds, mapDataPlugins)
}

// filterKindsByType 按类型筛选插件 Kind（返回已排序的 Kind 切片）
func filterKindsByType(pluginType Kind, plugins map[Kind]DataAdapter) []Kind {
	matchCount := 0
	for kind := range plugins {
		if pluginType == 0 || kind&pluginType == pluginType {
			matchCount++
		}
	}
	kinds := make([]Kind, 0, matchCount)
	for kind := range plugins {
		if pluginType == 0 || kind&pluginType == pluginType {
			kinds = append(kinds, kind)
		}
	}
	slices.Sort(kinds)
	return kinds
}

// filterKindsByTypeAndKeywords 按类型+关键词筛选插件 Kind（返回已排序的 Kind 切片）
func filterKindsByTypeAndKeywords(pluginType Kind, keywords []string, plugins map[Kind]DataAdapter) []Kind {
	matchCount := 0
	for kind, plugin := range plugins {
		if kind&pluginType == pluginType && slices.Contains(keywords, plugin.Key()) {
			matchCount++
		}
	}
	kinds := make([]Kind, 0, matchCount)
	for kind, plugin := range plugins {
		if kind&pluginType == pluginType && slices.Contains(keywords, plugin.Key()) {
			kinds = append(kinds, kind)
		}
	}
	slices.Sort(kinds)
	return kinds
}

// kindsToPlugins 将 Kind 切片转换为 DataAdapter 切片
func kindsToPlugins(kinds []Kind, plugins map[Kind]DataAdapter) []DataAdapter {
	list := make([]DataAdapter, 0, len(kinds))
	for _, kind := range kinds {
		if plugin, ok := plugins[kind]; ok {
			list = append(list, plugin)
		}
	}
	return list
}
