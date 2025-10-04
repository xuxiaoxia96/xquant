package east_money

import (
	"encoding/json"
	"fmt"
	urlpkg "net/url"

	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gox/http"
)

const (
	// https://data.eastmoney.com/rzrq/detail/all.html
	// https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPTA_WEB_RZRQ_GGMX&columns=ALL&source=WEB&pageNumber=1&pageSize=10&sortColumns=rzjme&sortTypes=-1&filter=(DATE%3D%272023-12-28%27)&callback=jQuery112303199655251283524_1703887938254&_=1703887938257
	urlEastMoneyApiRZRQ = "https://datacenter-web.eastmoney.com/api/data/v1/get"
	rzrqPageSize        = 500
)

// SecurityMarginTrading 融资融券
type SecurityMarginTrading struct {
	DATE            string  `name:"日期" json:"DATE"`
	MARKET          string  `name:"市场" json:"MARKET"`
	SCODE           string  `name:"代码" json:"SCODE"`
	SecName         string  `name:"证券名称" json:"SECNAME"`
	RZYE            float64 `name:"融资余额(元)" json:"RZYE"`
	RQYL            float64 `name:"融券余量(股)" json:"RQYL"`
	RZRQYE          float64 `name:"融资融券余额(元)" json:"RZRQYE"`
	RQYE            float64 `name:"融券余额(元)" json:"RQYE"`
	RQMCL           float64 `name:"融券卖出量(股)" json:"RQMCL"`
	RZRQYECZ        float64 `name:"融资融券余额差值(元)" json:"RZRQYECZ"`
	RZMRE           float64 `name:"融资买入额(元)" json:"RZMRE"`
	SZ              float64 `name:"SZ" json:"SZ"`
	RZYEZB          float64 `name:"融资余额占流通市值比(%)" json:"RZYEZB"`
	RZMRE3D         float64 `name:"3日融资买入额(元)" json:"RZMRE3D"`
	RZMRE5D         float64 `name:"5日融资买入额(元)" json:"RZMRE5D"`
	RZMRE10D        float64 `name:"10日融资买入额(元)" json:"RZMRE10D"`
	RZCHE           float64 `name:"融资偿还额(元)" json:"RZCHE"`
	RZCHE3D         float64 `name:"3日融资偿还额(元)" json:"RZCHE3D"`
	RZCHE5D         float64 `name:"5日融资偿还额(元)" json:"RZCHE5D"`
	RZCHE10D        float64 `name:"10日融资偿还额(元)" json:"RZCHE10D"`
	RZJME           float64 `name:"融资净买额(元)" json:"RZJME"`
	RZJME3D         float64 `name:"3日融资净买额(元)" json:"RZJME3D"`
	RZJME5D         float64 `name:"5日融资净买额(元)" json:"RZJME5D"`
	RZJME10D        float64 `name:"10日融资净买额(元)" json:"RZJME10D"`
	RQMCL3D         float64 `name:"3日融券卖出量(股)" json:"RQMCL3D"`
	RQMCL5D         float64 `name:"5日融券卖出量(股)" json:"RQMCL5D"`
	RQMCL10D        float64 `name:"10日融券卖出量(股)" json:"RQMCL10D"`
	RQCHL           float64 `name:"融券偿还量(股)" json:"RQCHL"`
	RQCHL3D         float64 `name:"3日融券偿还量(股)" json:"RQCHL3D"`
	RQCHL5D         float64 `name:"5日融券偿还量(股)" json:"RQCHL5D"`
	RQCHL10D        float64 `name:"10日融券偿还量(股)" json:"RQCHL10D"`
	RQJMG           float64 `name:"融券净卖出(股)" json:"RQJMG"`
	RQJMG3D         float64 `name:"3日融券净卖出(股)" json:"RQJMG3D"`
	RQJMG5D         float64 `name:"5日融券净卖出(股)" json:"RQJMG5D"`
	RQJMG10D        float64 `name:"10日融券净卖出(股)" json:"RQJMG10D"`
	SPJ             float64 `name:"收盘价" json:"SPJ"`
	ZDF             float64 `name:"涨跌幅" json:"ZDF"`
	RChange3DCP     float64 `name:"3日未识别" json:"RCHANGE3DCP"`
	RChange5DCP     float64 `name:"5日未识别" json:"RCHANGE5DCP"`
	RChange10DCP    float64 `name:"10日未识别" json:"RCHANGE10DCP"`
	KCB             int     `name:"科创板"  json:"KCB"`
	TradeMarketCode string  `name:"二级市场代码" json:"TRADE_MARKET_CODE"`
	TradeMarket     string  `name:"二级市场" json:"TRADE_MARKET"`
	FinBalanceGr    float64 `json:"FIN_BALANCE_GR"`
	SecuCode        string  `name:"证券代码" json:"SECUCODE"`
}

type rawMarginTrading struct {
	Version string `json:"version"`
	Result  struct {
		Pages int                     `json:"pages"`
		Data  []SecurityMarginTrading `json:"data"`
		Count int                     `json:"count"`
	} `json:"result"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func rawMarginTradingList(date string, pageNumber int) ([]SecurityMarginTrading, int, error) {
	tradeDate := exchange.FixTradeDate(date)
	params := urlpkg.Values{
		"reportName":  {"RPTA_WEB_RZRQ_GGMX"},
		"columns":     {"ALL"},
		"source":      {"WEB"},
		"sortColumns": {"scode"},
		"sortTypes":   {"1"},
		"pageSize":    {fmt.Sprintf("%d", rzrqPageSize)},
		"pageNumber":  {fmt.Sprintf("%d", pageNumber)},
		"client":      {"WEB"},
		"filter":      {fmt.Sprintf(`(DATE='%s')`, tradeDate)},
	}

	url := urlEastMoneyApiRZRQ + "?" + params.Encode()
	data, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	var raw rawMarginTrading
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, 0, err
	}
	return raw.Result.Data, raw.Result.Pages, nil
}

// GetMarginTradingList 获取两融列表
//
//	东方财富两融数据只有前一个交易日的数据
func GetMarginTradingList(date string) []SecurityMarginTrading {
	var list []SecurityMarginTrading
	pages := 1
	for i := 0; i < pages; i++ {
		tmpList, tmpPages, err := rawMarginTradingList(date, i+1)
		if err != nil {
			break
		}
		list = append(list, tmpList...)
		if len(tmpList) < rzrqPageSize {
			break
		}
		if pages == 1 {
			pages = tmpPages
		}
	}
	return list
}
