package indicators

import (
	"gitee.com/quant1x/pandas"
	. "gitee.com/quant1x/pandas/formula"
)

// BollingerBands 布林带指标
// MB: N日移动平均线
// UP: MB + K*N日收盘价标准差
// DN: MB - K*N日收盘价标准差
func BollingerBands(df pandas.DataFrame, period int, stdDev float64) pandas.DataFrame {
	CLOSE := df.ColAsNDArray("close")

	// 计算中轨
	MB := MA(CLOSE, period)

	// 计算标准差
	STD := STDDEV(CLOSE, period)

	// 计算上轨和下轨
	UP := MB.Add(STD.Mul(stdDev))
	DN := MB.Sub(STD.Mul(stdDev))

	// 创建结果DataFrame
	result := pandas.NewDataFrame(df.Col("date"), df.Col("close"))

	// 使用NewSeriesWithType添加系列
	mbSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "MB", MB)
	upSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "UP", UP)
	dnSeries := pandas.NewSeriesWithType(pandas.SERIES_TYPE_FLOAT64, "DN", DN)

	result = result.Join(mbSeries, upSeries, dnSeries)

	return result
}
