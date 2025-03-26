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
	logs.SetCallDepth(4)
}

// preFormat support %j format decorator which render output as json marshal
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

		s[0] = 's'

		if enableIndentJsonLog {
			v[cnt] = utils.MustMarshalIndent(v[cnt])
		} else {
			v[cnt] = utils.MustMarshal(v[cnt])
		}
	}
	return utils.ImmutableBytesToString(ss), v
}

func Debugf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelDebug {
		return
	}

	format, v = preFormat(format, v...)
	logs.Debug(format, v...)
}

func Infof(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelInfo {
		return
	}
	format, v = preFormat(format, v...)
	logs.Info(format, v...)
}

func Errorf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelError {
		return
	}
	format, v = preFormat(format, v...)
	logs.Error(format, v...)
}

func Warnf(format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelWarn {
		return
	}
	format, v = preFormat(format, v...)
	logs.Warn(format, v...)
}

func CtxDebugf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelDebug {
		return
	}

	format, v = preFormat(format, v...)
	logs.CtxDebug(ctx, format, v...)
}

func CtxInfof(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelInfo {
		return
	}
	format, v = preFormat(format, v...)
	logs.CtxInfo(ctx, format, v...)
}

func CtxErrorf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelError {
		return
	}
	format, v = preFormat(format, v...)
	logs.CtxError(ctx, format, v...)
}

func CtxWarnf(ctx context.Context, format string, v ...any) {
	if config.CurrentLogLevel > hlog.LevelWarn {
		return
	}
	format, v = preFormat(format, v...)
	logs.CtxWarn(ctx, format, v...)
}
