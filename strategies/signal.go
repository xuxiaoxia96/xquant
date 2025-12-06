package strategies

// Signal 策略交易信号
//
//	表示策略评估后产生的交易信号，包含股票信息、价格信息、板块信息等
//
// Signal 的计算过程分为三个阶段：
//
// 1. 策略评估阶段（Strategy.Evaluate）：
//   - 策略通过 Evaluate 方法评估个股，当满足策略条件时创建 Signal
//   - 从 factors.GetL5History() 获取历史K线数据（MA5, MA10, MA20等）
//   - 从 models.GetStrategySnapshot() 获取当前快照数据（价格、日期等）
//   - 计算技术指标（如移动平均线、金叉等）
//   - 如果满足策略条件，创建 Signal 并设置基础字段
//
// 2. 板块信息填充阶段（tracker/mod_sector.go）：
//   - 在策略评估完成后，遍历所有 Signal
//   - 从板块数据（__mapBlockData, __stock2Block, __stock2Rank）中填充板块相关字段
//
// 3. 趋势预测阶段（可选，当前实现已注释）：
//   - 调用 PredictTrend() 方法根据历史K线数据预测短期趋势
//   - 使用线性回归分析最近N天的开盘价、收盘价、最高价、最低价
//   - 根据预测值判断趋势（高开/平开/低开，冲高回落/探底回升/趋势向下/短线向好）
//
// 数据来源说明：
//   - QuoteSnapshot: 从 models.GetStrategySnapshot() 获取，包含实时行情数据
//   - History: 从 factors.GetL5History() 获取，包含历史K线和技术指标
//   - BlockData: 从板块扫描模块获取，包含板块相关统计信息
type Signal struct {
	// ========== 基础股票信息 ==========
	Code string `name:"证券代码" dataframe:"code"`
	// 证券代码，在策略评估时从 securityCode 参数获取

	Name string `name:"证券名称" dataframe:"name"`
	// 证券名称，通过 securities.GetStockName(securityCode) 获取

	Date string `name:"信号日期" dataframe:"date"`
	// 信号日期，从 snapshot.Date 获取（models.GetStrategySnapshot() 返回的快照数据）

	// ========== 交易相关指标 ==========
	TurnZ float64 `name:"开盘换手Z" dataframe:"turn_z"`
	// 开盘换手Z，通常从 snapshot.OpenTurnZ 获取
	// OpenTurnZ 的计算：通过 factors.GetL5F10(securityCode).TurnZ(snapshot.OpenVolume) 计算
	// 公式：开盘换手Z = (开盘成交量 / 流通股本) * 100

	Rate float64 `name:"涨跌幅%" dataframe:"rate"`
	// 涨跌幅%，通常从 snapshot.ChangeRate 获取
	// ChangeRate 的计算：通过 num.NetChangeRate(snapshot.LastClose, snapshot.Price) 计算
	// 公式：涨跌幅% = ((现价 - 昨收) / 昨收) * 100
	// 注意：部分策略（如 ModelNo1）在创建 Signal 时初始化为 0.00

	Buy float64 `name:"委托价格" dataframe:"buy"`
	// 委托买入价格，从 snapshot.Price 获取（当前价格）

	Sell float64 `name:"目标价格" dataframe:"sell"`
	// 目标卖出价格，计算为 Buy * 1.05（默认5%涨幅目标）
	// 部分策略可能使用其他计算方式

	// ========== 策略信息 ==========
	StrategyCode uint64 `name:"策略编码" dataframe:"strategy_code"`
	// 策略编码，从策略的 Code() 方法获取（如 ModelHousNo1 = 1）

	StrategyName string `name:"策略名称" dataframe:"strategy_name"`
	// 策略名称，从策略的 Name() 方法获取（如 "1号策略"）

	// ========== 板块信息（在 tracker/mod_sector.go 中填充）==========
	BlockType string `name:"板块类型" dataframe:"block_type"`
	// 板块类型，从 __mapBlockData[blockCode].Type 获取

	BlockCode string `name:"板块代码" dataframe:"block_code"`
	// 板块代码，从 __stock2Block[stockCode] 映射关系中获取

	BlockName string `name:"板块名称" dataframe:"block_name"`
	// 板块名称，从 __mapBlockData[blockCode].Name 获取

	BlockRate float64 `name:"板块涨幅%" dataframe:"block_rate"`
	// 板块涨幅%，从 __mapBlockData[blockCode].ChangeRate 获取

	BlockTop int `name:"板块排名" dataframe:"block_top"`
	// 板块排名，从 __mapBlockData[blockCode].Rank 获取

	BlockRank int `name:"个股排名" dataframe:"block_rank"`
	// 个股在板块中的排名，从 __stock2Rank[stockCode].TopNo 获取

	BlockZhangTing string `name:"板块涨停数" dataframe:"block_zhangting"`
	// 板块涨停数，格式化为 "涨停数/总数"
	// 从 __mapBlockData[blockCode].LimitUpNum 和 Count 计算
	// 格式：fmt.Sprintf("%d/%d", block.LimitUpNum, block.Count)

	BlockDescribe string `name:"涨/跌/平" dataframe:"block_describe"`
	// 板块涨跌平统计，格式化为 "上涨数/下跌数/平盘数"
	// 从 __mapBlockData[blockCode] 的 UpCount, DownCount, NoChangeNum 计算
	// 格式：fmt.Sprintf("%d/%d/%d", block.UpCount, block.DownCount, block.NoChangeNum)

	BlockTopCode string `name:"领涨股代码" dataframe:"block_top_code"`
	// 领涨股代码，从 __mapBlockData[blockCode].TopCode 获取

	BlockTopName string `name:"领涨股名称" dataframe:"block_top_name"`
	// 领涨股名称，从 __mapBlockData[blockCode].TopName 获取

	BlockTopRate float64 `name:"领涨股涨幅%" dataframe:"block_top_rate"`
	// 领涨股涨幅%，从 __mapBlockData[blockCode].TopRate 获取

	// ========== 趋势预测（可选）==========
	Tendency string `name:"短线趋势" dataframe:"tendency"`
	// 短线趋势预测，通过 PredictTrend() 方法计算
	// 当前实现已注释，计划使用线性回归分析最近N天的K线数据
	// 预测内容包括：高开/平开/低开，冲高回落/探底回升/趋势向下/短线向好
}

// PredictTrend 预测趋势
//
//	根据历史K线数据预测短期趋势（当前实现已注释）
//
// 计算逻辑（已注释，待实现）：
//  1. 获取最近N天（N=3）的K线数据：factors.BasicKLine(s.Code)
//  2. 提取开盘价(OPEN)、收盘价(CLOSE)、最高价(HIGH)、最低价(LOW)
//  3. 对每个价格序列进行线性回归（linear.CurveRegression）
//  4. 获取回归后的预测值（po, pc, ph, pl）
//  5. 判断开盘趋势：
//     - 如果 po > lastClose: "高开"
//     - 如果 po == lastClose: "平开"
//     - 如果 po < lastClose: "低开"
//  6. 判断盘中趋势：
//     - 如果 pl > ph: ",冲高回落"
//     - 如果 pl > pc: ",探底回升"
//     - 如果 pc < pl: ",趋势向下"
//     - 否则: ",短线向好"
//  7. 将趋势描述写入 s.Tendency 字段
func (s *Signal) PredictTrend() {
	//N := 3
	//df := factors.BasicKLine(s.Code)
	//if df.Nrow() < N+1 {
	//	return
	//}
	//limit := api.RangeFinite(-N)
	//OPEN := df.Col("open").Select(limit)
	//CLOSE := df.Col("close").Select(limit)
	//HIGH := df.Col("high").Select(limit)
	//LOW := df.Col("low").Select(limit)
	//lastClose := num.AnyToFloat64(CLOSE.IndexOf(-1))
	//po := linear.CurveRegression(OPEN).IndexOf(-1).(num.DType)
	//pc := linear.CurveRegression(CLOSE).IndexOf(-1).(num.DType)
	//ph := linear.CurveRegression(HIGH).IndexOf(-1).(num.DType)
	//pl := linear.CurveRegression(LOW).IndexOf(-1).(num.DType)
	//if po > lastClose {
	//	s.Tendency = "高开"
	//} else if po == lastClose {
	//	s.Tendency = "平开"
	//} else {
	//	s.Tendency = "低开"
	//}
	//if pl > ph {
	//	s.Tendency += ",冲高回落"
	//} else if pl > pc {
	//	s.Tendency += ",探底回升"
	//} else if pc < pl {
	//	s.Tendency += ",趋势向下"
	//} else {
	//	s.Tendency += ",短线向好"
	//}
	//
	//fs := []float64{float64(po), float64(pc), float64(ph), float64(pl)}
	//sort.Float64s(fs)
	//
	//_ = lastClose
}
