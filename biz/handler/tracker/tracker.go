package tracker

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"xquant/biz/handler"
	trackermodel "xquant/biz/model/tracker"
	trackerservice "xquant/biz/service/tracker"
	"xquant/pkg/log"
	"xquant/pkg/openapi_error"
)

// Tracker 实时跟踪策略在当前市场的表现，输出表格
func Tracker(ctx context.Context, c *app.RequestContext) {
	// 解析并验证请求参数
	var req trackermodel.TrackerRequest
	if err := c.BindAndValidate(&req); err != nil {
		log.CtxErrorf(ctx, "[Tracker] 参数绑定失败: %s", err)
		handler.OpenAPIFail(ctx, c, openapi_error.NewInvalidParameterError(ctx, "", err.Error()))
		return
	}

	trackerStrategyCodes := req.GetTrackerStrategyCodes()
	if len(trackerStrategyCodes) == 0 {
		log.CtxWarnf(ctx, "[Tracker] 未指定跟踪的策略代码")
		return // 无跟踪目标，直接返回
	}

	trackerservice.RunTrackerCore(ctx, trackerservice.TrackerCoreParams{TrackerStrategyCodes: trackerStrategyCodes, IsDebug: req.IsDebug})
}
