# ProgressBar 适配器

## 概述

本模块提供了对 `github.com/schollz/progressbar/v3` 的适配器封装，兼容原有的 `gitee.com/quant1x/gox/progressbar` API。

## 迁移完成

✅ **已成功从 Gitee 迁移到 GitHub 开源库**

- 使用 `github.com/schollz/progressbar/v3` 作为底层实现
- 保持原有 API 完全兼容，无需修改调用代码
- 添加了错误处理和边界检查
- 完整的单元测试覆盖

## API 使用

```go
import "xquant/pkg/progressbar"

// 创建进度条
bar := progressbar.NewBar(index, "描述", total)

// 更新进度
bar.Add(1)

// 等待完成
bar.Wait()
```

## 特性

1. **完全兼容原有API** - 无需修改现有代码
2. **错误处理** - 自动处理边界情况（total <= 0, nil 检查等）
3. **美观显示** - 使用统一的主题和样式
4. **性能优化** - 底层使用高效的进度条库

## 测试

运行测试：
```bash
go test ./pkg/progressbar/... -v
```

测试覆盖：
- ✅ 创建进度条
- ✅ 增加进度
- ✅ 等待完成
- ✅ 边界情况（total=0, 负数等）

## 依赖

- `github.com/schollz/progressbar/v3` - 底层进度条实现

## 使用示例

```go
// tracker/tracker.go 中的使用
bar := progressbar.NewBar(*barIndex, "执行["+model.Name()+"全市场扫描]", stockCount)
for start := 0; start < stockCount; start++ {
    bar.Add(1)
    // ... 处理逻辑 ...
}
bar.Wait()
```

## 迁移说明

所有使用 `xquant/pkg/progressbar` 的模块已自动使用新的 GitHub 实现，无需任何代码修改。

