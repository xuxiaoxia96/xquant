package indicators

import (
	"gitee.com/quant1x/pandas"
	. "gitee.com/quant1x/pandas/formula"
)

// MACD 指数平滑异同平均线
// DIF: 12日EMA - 26日EMA
// DEA: DIF的9日EMA
// MACD: (DIF - DEA) * 2
func MACD(df pandas.DataFrame, short, long, signal int) pandas.DataFrame {
	CLOSE := df.ColAsNDArray("close")

	// 计算EMA
	emaShort := EMA(CLOSE, short)
	emaLong := EMA(CLOSE, long)

	// 计算DIF
	DIF := emaShort.Sub(emaLong)

	// 计算DEA
	DEA := EMA(DIF, signal)

	// 计算MACD
	MACD := DIF.Sub(DEA).Mul(2)

	// 创建结果DataFrame
	result := pandas.NewDataFrame(df.Col("date"), df.Col("close"))

	// 使用NewSeriesWithType添加系列
	difSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "DIF", DIF)
	deaSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "DEA", DEA)
	macdSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "MACD", MACD)

	result = result.Join(difSeries, deaSeries, macdSeries)

	return result
}
