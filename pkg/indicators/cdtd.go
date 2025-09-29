package indicators

import (
	"gitee.com/quant1x/pandas"
	. "gitee.com/quant1x/pandas/formula"
)

// CDTD 抄底逃顶指标
func CDTD(df pandas.DataFrame) pandas.DataFrame {
	var (
		CLOSE = df.ColAsNDArray("close") // 收盘价
		HIGH  = df.ColAsNDArray("high")  // 最高价
		LOW   = df.ColAsNDArray("low")   // 最低价
	)

	N1 := 3
	N2 := 9
	N3 := 27
	N4 := 5

	HV3 := HHV(HIGH, N3)
	LV3 := LLV(LOW, N3)
	HHVLLV3 := HV3.Sub(LV3)
	CLLV3 := CLOSE.Sub(LV3)

	// 计算RSV1
	LV2 := LLV(LOW, N2)
	HV2 := HHV(HIGH, N2)
	rsv11 := CLOSE.Sub(LV2)
	rsv12 := HV2.Sub(LV2)
	RSV1 := rsv11.Div(rsv12).Mul(100)

	// 计算RSV2和RSV3
	RSV2 := CLLV3.Div(HHVLLV3).Mul(100)
	RSV3 := SMA(RSV2, N4, 1)

	// 计算WEN
	RSV4 := SMA(RSV3, N1, 1)
	WEN := RSV3.Mul(N1).Sub(RSV4.Mul(2))

	// 计算J1和J2
	J1 := SMA(RSV1, N1, 1)
	J2 := SMA(J1, N1, 1)

	// 计算买入卖出信号
	Trend := WEN
	S1 := CROSS(J2, J1)
	S2 := J2.Gt(85)
	S := S1.And(S2)
	B := Trend.Lt(REF(Trend, 1)).And(Trend.Lte(5))
	WS := Trend.Gte(85.00)
	WB := Trend.Lte(5.00)

	// 创建结果DataFrame
	result := pandas.NewDataFrame(df.Col("date"), df.Col("close"))

	// 使用NewSeriesWithType添加系列
	os1 := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "S1", S1)
	os2 := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "S2", S2)
	os := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "S", S)
	ob := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "B", B)
	ows := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "WS", WS)
	owb := pandas.NewSeriesWithType(pandas.SERIES_TYPE_BOOL, "WB", WB)

	result = result.Join(ob, os, os1, os2, ows, owb)
	return result
}
