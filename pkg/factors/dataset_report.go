package factors

import (
	"context"
	"gitee.com/quant1x/gotdx/quotes"
	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/progressbar"
	"xquant/pkg/cache"
	"xquant/pkg/datasource/east_money"
)

// DataQuarterlyReport 季报
type DataQuarterlyReport struct {
	cache.DataSummary
	Date  string
	Code  string
	cache map[string]east_money.QuarterlyReport
}

func init() {
	summary := __mapDataSets[BaseQuarterlyReports]
	_ = cache.Register(&DataQuarterlyReport{DataSummary: summary})
}

func (r *DataQuarterlyReport) Clone(date string, code string) DataSet {
	summary := __mapDataSets[BaseQuarterlyReports]
	var dest = DataQuarterlyReport{DataSummary: summary, Date: date, Code: code}
	return &dest
}

func (r *DataQuarterlyReport) GetDate() string {
	return r.Date
}

func (r *DataQuarterlyReport) GetSecurityCode() string {
	return r.Code
}

func (r *DataQuarterlyReport) Print(code string, date ...string) {
	//TODO implement me
	panic("implement me")
}

func (r *DataQuarterlyReport) Filename(date, code string) string {
	//TODO implement me
	panic("implement me")
}

func (r *DataQuarterlyReport) Init(ctx context.Context, date string) error {
	barIndex := 1
	value, ok := ctx.Value(cache.KBarIndex).(int)
	if ok {
		barIndex = value
	}
	barIndex++
	r.cache = IntegrateQuarterlyReports(barIndex, date)
	return nil
}

func (r *DataQuarterlyReport) Checkout(securityCode, date string) {
	//TODO implement me
	panic("implement me")
}

func (r *DataQuarterlyReport) Check(cacheDate, featureDate string) error {
	//TODO implement me
	panic("implement me")
}

func (r *DataQuarterlyReport) Update(date string) {
	_ = date
}

func (r *DataQuarterlyReport) Repair(date string) {
	_ = date
}

func (r *DataQuarterlyReport) Increase(snapshot quotes.Snapshot) {
	_ = snapshot
}

// IntegrateQuarterlyReports 更新季报数据
func IntegrateQuarterlyReports(barIndex int, date string) map[string]east_money.QuarterlyReport {
	modName := "季报概要信息"
	logger.Info(modName + ", 任务开始启动...")

	allReports := []east_money.QuarterlyReport{}
	reports, pages, _ := east_money.QuarterlyReports(date)
	if pages < 1 || len(reports) == 0 {
		return nil
	}
	allReports = append(allReports, reports...)
	bar := progressbar.NewBar(barIndex, "执行["+modName+"]", pages-1)
	for pageNo := 2; pageNo < pages+1; pageNo++ {
		bar.Add(1)
		list, pages, err := east_money.QuarterlyReports(date, pageNo)
		if err != nil || pages < 1 {
			logger.Error(err)
			break
		}
		count := len(list)
		if count == 0 {
			break
		}
		allReports = append(allReports, list...)
		if count < east_money.EastmoneyQuarterlyReportAllPageSize {
			break
		}
	}
	mapReports := map[string]east_money.QuarterlyReport{}
	if len(allReports) > 0 {
		for _, v := range allReports {
			mapReports[v.SecurityCode] = v
		}
		_, qEnd := api.GetQuarterDayByDate(date)
		filename := cache.ReportsFilename(qEnd)
		err := api.SlicesToCsv(filename, allReports)
		if err != nil {
			logger.Errorf("cache %s failed, error: %+v", filename, err)
		}
	}
	return mapReports
}
