package progressbar

import (
	"github.com/schollz/progressbar/v3"
)

// Bar 进度条接口（兼容原有API）
type Bar interface {
	Add(int)
	Wait()
}

// progressBarAdapter 适配器，兼容原有API
type progressBarAdapter struct {
	bar *progressbar.ProgressBar
}

// NewBar 创建进度条（兼容原有API）
// index: 进度条索引（新库不使用，保留以兼容原有调用）
// description: 进度条描述
// total: 总数
func NewBar(index int, description string, total int) Bar {
	if total <= 0 {
		total = 1 // 避免除零错误
	}

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
		progressbar.OptionOnCompletion(func() {
			// 完成时自动换行，保持输出格式
		}),
	)
	return &progressBarAdapter{bar: bar}
}

// Add 增加进度
func (p *progressBarAdapter) Add(n int) {
	if p.bar != nil && n > 0 {
		_ = p.bar.Add(n)
	}
}

// Wait 等待完成并清理
func (p *progressBarAdapter) Wait() {
	if p.bar != nil {
		_ = p.bar.Finish()
	}
}
