package log

import (
	"bytes"
	"context"
	"os"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"xquant/pkg/config"
	"xquant/pkg/utils"
)

var enableIndentJsonLog = os.Getenv("MAAS_API_PROXY_ENABLE_INDENT_JSON_LOG") == "true"

func init() {
	// 配置hlog的调用深度，确保日志中显示正确的调用位置
}

// preFormat 处理%j格式装饰器，将输出渲染为JSON格式
func preFormat(format string, v ...any) (string, []any) {
	s := []byte(format)
	ss := s
	for cnt, idx := 0, bytes.IndexByte(s, '%'); idx >= 0 && len(s) > 0 && cnt < len(v); cnt, idx = cnt+1, bytes.IndexByte(s, '%') {
		s = s[idx+1:]
		if len(s) == 0 {
			break
		}

		if s[0] != 'j' {
			continue
		}

		// 将%j替换为%s，因为我们已经将值序列化为字符串
		s[0] = 's'

		// 根据配置决定是否缩进JSON输出
		if enableIndentJsonLog {
			v[cnt] = utils.MustMarshalIndent(v[cnt])
		} else {
			v[cnt] = utils.MustMarshal(v[cnt])
		}
	}
	return string(ss), v
}

// Debugf 输出调试级别日志
func Debugf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelDebug {
		return
	}

	format, v = preFormat(format, v...)
	hlog.Debugf(format, v...)
}

// Infof 输出信息级别日志
func Infof(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelInfo {
		return
	}
	format, v = preFormat(format, v...)
	hlog.Infof(format, v...)
}

// Errorf 输出错误级别日志
func Errorf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelError {
		return
	}
	format, v = preFormat(format, v...)
	hlog.Errorf(format, v...)
}

// Warnf 输出警告级别日志
func Warnf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelWarn {
		return
	}
	format, v = preFormat(format, v...)
	hlog.Warnf(format, v...)
}

// CtxDebugf 输出带上下文的调试级别日志
func CtxDebugf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelDebug {
		return
	}

	format, v = preFormat(format, v...)
	hlog.CtxDebugf(ctx, format, v...)
}

// CtxInfof 输出带上下文的信息级别日志
func CtxInfof(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelInfo {
		return
	}
	format, v = preFormat(format, v...)
	hlog.CtxInfof(ctx, format, v...)
}

// CtxErrorf 输出带上下文的错误级别日志
func CtxErrorf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelError {
		return
	}
	format, v = preFormat(format, v...)
	hlog.CtxErrorf(ctx, format, v...)
}

// CtxWarnf 输出带上下文的警告级别日志
func CtxWarnf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelWarn {
		return
	}
	format, v = preFormat(format, v...)
	hlog.CtxWarnf(ctx, format, v...)
}
