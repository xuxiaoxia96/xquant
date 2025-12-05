# 盘内跟踪交易流程说明

## 概述

本文档说明重构后的盘内跟踪交易流程。重构后的代码结构更清晰，职责分离更明确。

## 完整流程

```
Tracker (主循环)
  └─> snapshotTracker (快照跟踪)
       └─> ProcessStrategyResults (处理策略结果)
            ├─> getCurrentTradeDateAndTime (获取交易日期时间)
            ├─> buildStatisticsFromSnapshots (构建统计数据)
            ├─> renderTableAndWinRate (渲染表格和胜率)
            │    ├─> calculateWinRateStatistics (计算胜率)
            │    └─> printWinRateStatistics (打印胜率)
            └─> UpdateStockPoolAndExecuteTrading (更新股票池并执行交易)
                 └─> mergeStockPoolAndExecuteTrading (合并股票池并执行交易)
                      ├─> mergeStockPool (合并股票池)
                      │    ├─> buildStockPoolMapFromStatistics (构建股票池映射)
                      │    ├─> processExistingStockPool (处理已存在的股票池)
                      │    └─> extractNewStocks (提取新增标的)
                      └─> checkOrderForBuy (检查并执行买入)
```

## 详细说明

### 1. Tracker (tracker/tracker.go)
**职责**: 主循环，控制盘中跟踪的整体流程
- 检查交易时段
- 同步快照数据
- 遍历策略并执行跟踪

### 2. snapshotTracker (tracker/tracker.go)
**职责**: 单个策略的快照跟踪
- 获取股票代码列表
- 加载快照数据
- 过滤不符合条件的个股
- 对结果集排序
- 调用 `ProcessStrategyResults` 处理结果

### 3. ProcessStrategyResults (tracker/tracker.go)
**职责**: 处理策略扫描结果的核心函数
- **获取交易日期时间**: `getCurrentTradeDateAndTime()`
- **构建统计数据**: `buildStatisticsFromSnapshots()` - 将快照数据转换为统计数据
- **渲染表格和胜率**: `renderTableAndWinRate()` - 输出表格并计算胜率统计
- **更新股票池并执行交易**: `UpdateStockPoolAndExecuteTrading()` - 更新股票池，如有新增则执行交易

### 4. buildStatisticsFromSnapshots (tracker/tracker.go)
**职责**: 从快照数据构建统计数据
- 转换快照为 `Statistics` 结构
- 计算趋势 (`calculateTendency`)
- 填充板块信息 (`fillBlockInfo`)

### 5. renderTableAndWinRate (tracker/tracker.go)
**职责**: 渲染表格并计算胜率统计
- 渲染控制台表格
- 计算胜率统计 (`calculateWinRateStatistics`)
- 打印胜率统计 (`printWinRateStatistics`)

### 6. UpdateStockPoolAndExecuteTrading (storages/stockpool_merge.go)
**职责**: 更新股票池并执行交易的入口函数
- 检查策略配置是否有效
- 调用 `mergeStockPoolAndExecuteTrading` 执行合并和交易

### 7. mergeStockPoolAndExecuteTrading (storages/stockpool_merge.go)
**职责**: 合并股票池并执行交易的核心函数
- **合并股票池**: `mergeStockPool()` - 将策略扫描结果合并到本地股票池
- **执行交易**: 如果有新增标的，调用 `checkOrderForBuy()` 执行买入

### 8. mergeStockPool (storages/stockpool_merge.go)
**职责**: 合并股票池逻辑
- **构建股票池映射**: `buildStockPoolMapFromStatistics()` - 将统计数据转换为股票池格式
- **处理已存在的股票池**: `processExistingStockPool()` - 标记已存在的标的，召回不再出现的标的
- **提取新增标的**: `extractNewStocks()` - 提取未在本地股票池中存在的标的

### 9. checkOrderForBuy (storages/stockpool_trade.go)
**职责**: 检查并执行买入交易
- 检查交易日、交易时段等条件
- 计算可用资金
- 执行买入订单

## 重构改进点

### 1. 函数职责单一化
- **之前**: `OutputTable` 既输出表格，又计算统计，还执行交易
- **之后**: 拆分为多个职责单一的函数：
  - `buildStatisticsFromSnapshots` - 构建数据
  - `renderTableAndWinRate` - 渲染和统计
  - `UpdateStockPoolAndExecuteTrading` - 更新股票池和交易

### 2. 命名更清晰
- **之前**: `OutputStatistics` - 名称不清晰，实际是更新股票池并执行交易
- **之后**: `UpdateStockPoolAndExecuteTrading` - 明确表达函数职责

### 3. 交易逻辑显式化
- **之前**: `stockPoolMerge` 隐含了交易逻辑，在函数内部直接调用 `checkOrderForBuy`
- **之后**: 
  - `mergeStockPool` - 纯合并逻辑
  - `mergeStockPoolAndExecuteTrading` - 显式地先合并，再执行交易

### 4. 代码可读性提升
- 每个函数都有清晰的职责
- 函数名能准确表达其功能
- 流程更易于理解和维护

## 关键数据结构

### Statistics (models/statistics.go)
策略扫描结果的统计数据，包含：
- 日期、代码、名称
- 价格信息（开盘、现价、涨跌幅等）
- 趋势描述
- 板块信息

### StockPool (storages/stockpool.go)
股票池中的标的，包含：
- 状态（命中、召回、已下单等）
- 策略信息
- 订单状态和ID
- 价格信息

## 注意事项

1. **线程安全**: `mergeStockPoolAndExecuteTrading` 使用 `poolMutex` 保证线程安全
2. **交易条件**: 只有在策略配置有效、有新增标的时才执行交易
3. **股票池缓存**: 股票池数据保存在本地 CSV 文件中，通过 `getStockPoolFromCache` 和 `saveStockPoolToCache` 管理
