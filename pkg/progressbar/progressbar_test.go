package progressbar

import (
	"testing"
	"time"
)

// TestNewBar 测试创建进度条
func TestNewBar(t *testing.T) {
	bar := NewBar(1, "测试进度条", 100)
	if bar == nil {
		t.Fatal("NewBar 返回 nil")
	}
}

// TestBarAdd 测试增加进度
func TestBarAdd(t *testing.T) {
	bar := NewBar(1, "测试增加进度", 10)

	// 测试正常增加
	bar.Add(1)
	bar.Add(2)
	bar.Add(3)

	// 测试完成
	bar.Wait()
}

// TestBarWait 测试等待完成
func TestBarWait(t *testing.T) {
	bar := NewBar(1, "测试等待完成", 5)

	for i := 0; i < 5; i++ {
		bar.Add(1)
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
	}

	bar.Wait()
}

// TestBarZeroTotal 测试总数为0的情况
func TestBarZeroTotal(t *testing.T) {
	bar := NewBar(1, "测试总数为0", 0)
	if bar == nil {
		t.Fatal("NewBar 在总数为0时返回 nil")
	}
	bar.Add(1)
	bar.Wait()
}

// TestBarNegativeAdd 测试负数增加（应该被忽略）
func TestBarNegativeAdd(t *testing.T) {
	bar := NewBar(1, "测试负数增加", 10)
	bar.Add(-1) // 应该被忽略
	bar.Add(1)  // 正常增加
	bar.Wait()
}
