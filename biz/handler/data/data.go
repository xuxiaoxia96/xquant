package data

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"

	"xquant/biz/handler"
	datamodel "xquant/biz/model/data"
	updateservice "xquant/biz/service/update"
	"xquant/pkg/log"
	"xquant/pkg/openapi_error"
)

func Update(ctx context.Context, c *app.RequestContext) {
	var req datamodel.UpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		log.CtxErrorf(ctx, "[Tracker] 参数绑定失败: %s", err)
		handler.OpenAPIFail(ctx, c, openapi_error.NewInvalidParameterError(ctx, "", err.Error()))
		return
	}

	params := updateservice.UpdateParams{
		IsFullUpdate:     req.IsFullUpdate,
		BaseDataKeywords: req.BaseDataKeyWords,
		FeaturesKeywords: req.FeaturesKeyWords,
	}

	// 步骤3：调用核心更新逻辑（与 HTTP 共用）
	if err := updateservice.RunUpdate(ctx, params); err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	fmt.Println("数据更新完成")
}
