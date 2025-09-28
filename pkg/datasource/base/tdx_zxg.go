package base

import (
	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/pandas"
	"xquant/pkg/cache"
)

const (
	BlockPath = "/T0002/blocknew"
	ZxgBlk    = "zxg.blk"
	BkltBlk   = "BKLT.blk"
	ZdBk      = "ZDBK.blk"
)

func GetZxgList() []string {
	filename := cache.GetZxgFile()
	df := pandas.ReadCSV(filename)
	if df.Nrow() == 0 {
		return []string{}
	}
	rows := df.Col("code")
	if rows.Len() == 0 {
		return []string{}
	}
	// 校验证券代码, 统一格式前缀加代码
	cs := rows.Strings()
	for i, v := range cs {
		cs[i] = exchange.CorrectSecurityCode(v)
	}
	return cs
}
