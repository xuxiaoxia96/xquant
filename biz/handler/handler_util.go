package handler

import (
	"context"
	"net/http"
	"xquant/pkg/log"
	"xquant/pkg/openapi_error"

	"github.com/cloudwego/hertz/pkg/app"
)

func OpenAPISuccess(ctx context.Context, c *app.RequestContext, result interface{}) {
	log.CtxDebugf(ctx, "success response=%j", result)
	c.JSON(http.StatusOK, result)
}

func OpenAPIFail(ctx context.Context, c *app.RequestContext, err openapi_error.OpenAPIError) {
	var status int
	if err.HTTPStatusCode() == http.StatusInternalServerError {
		status = http.StatusInternalServerError
	} else {
		status = http.StatusOK
	}
	log.CtxErrorf(ctx, "status: %d, failed response=%j", status, openapi_error.ErrorResponse{Error: err})
	c.JSON(status, openapi_error.ErrorResponse{Error: err})
}
