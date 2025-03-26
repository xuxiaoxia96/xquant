package factors

// SecurityFeature 证券特征信息
type SecurityFeature struct {
	Date           string  `name:"日期" dataframe:"date,string"`
	Open           float64 `name:"开盘" dataframe:"open,float64"`
	Close          float64 `name:"收盘" dataframe:"close,float64"`
	High           float64 `name:"最高" dataframe:"high,float64"`
	Low            float64 `name:"最低" dataframe:"low,float64"`
	Volume         int64   `name:"成交量" dataframe:"volume,int64"`
	Amount         float64 `name:"成交额" dataframe:"amount,float64"`
	Up             int     `name:"上涨家数" dataframe:"up,int64"`   // 个股无效
	Down           int     `name:"下跌家数" dataframe:"down,int64"` // 个股无效
	LastClose      float64 `name:"昨收" dataframe:"last_close,float64"`
	ChangeRate     float64 `name:"涨跌幅" dataframe:"change_rate,float64"`
	OpenVolume     int64   `name:"开盘量" dataframe:"open_volume,int64"`
	OpenTurnZ      float64 `name:"开盘换手z" dataframe:"open_turnz,float64"`
	OpenUnmatched  int64   `name:"开盘未匹配" dataframe:"open_unmatched,int64"` // K线无效, 调取misc特征数据获取
	CloseVolume    int64   `name:"收盘量" dataframe:"close_volume,int64"`
	CloseTurnZ     float64 `name:"收盘换手z" dataframe:"close_turnz,float64"`
	CloseUnmatched int64   `name:"收盘未匹配" dataframe:"close_unmatched,int64"` // K线无效, 调取misc特征数据获取
	InnerVolume    int64   `name:"内盘" dataframe:"inner_volume,int64"`
	OuterVolume    int64   `name:"外盘" dataframe:"outer_volume,int64"`
	InnerAmount    float64 `name:"流出金额" dataframe:"inner_amount,float64"`
	OuterAmount    float64 `name:"流入金额" dataframe:"outer_amount,float64"`
	//State          int     `name:"数据状态" dataframe:"state"`
}

// CheckSum 校验和
func (this SecurityFeature) CheckSum() int {
	sign := 0
	sign += int(this.OpenVolume)
	sign += int(this.OpenTurnZ)
	sign += int(this.OpenUnmatched)
	sign += int(this.CloseVolume)
	sign += int(this.CloseTurnZ)
	sign += int(this.CloseUnmatched)
	sign += int(this.InnerVolume)
	sign += int(this.OuterVolume)
	sign += int(this.InnerAmount)
	sign += int(this.OuterAmount)
	return sign
}

// TurnoverDataSummary 换手数据概要
type TurnoverDataSummary struct {
	OpenVolume     int64   `name:"开盘量" dataframe:"open_volume,int64"`
	OpenTurnZ      float64 `name:"开盘换手z" dataframe:"open_turnz,float64"`
	OpenUnmatched  int64   `name:"开盘未匹配" dataframe:"open_unmatched,int64"` // K线无效, 调取misc特征数据获取
	CloseVolume    int64   `name:"收盘量" dataframe:"close_volume,int64"`
	CloseTurnZ     float64 `name:"收盘换手z" dataframe:"close_turnz,float64"`
	CloseUnmatched int64   `name:"收盘未匹配" dataframe:"close_unmatched,int64"` // K线无效, 调取misc特征数据获取
	InnerVolume    int64   `name:"内盘" dataframe:"inner_volume,int64"`
	OuterVolume    int64   `name:"外盘" dataframe:"outer_volume,int64"`
	InnerAmount    float64 `name:"流出金额" dataframe:"inner_amount,float64"`
	OuterAmount    float64 `name:"流入金额" dataframe:"outer_amount,float64"`
}
