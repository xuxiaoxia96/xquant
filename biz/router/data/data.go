package data

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"xquant/biz/middleware/mwctx"
)

func RegisterData(h *server.Hertz) {
	root := h.Group("/")
	{
		_data := root.Group("/data")
		{
			_data.POST("data", _dataHandler)
		}
	}
}

func _dataHandler() []app.HandlerFunc {
	mwCtx := mwctx.NewMiddlewareCtx(context.Background())
	return app.HandlerFunc{
		data.Update,
	}
}
