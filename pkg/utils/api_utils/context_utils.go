package api_utils

import (
	"context"
)

type ctxRequestContextKey struct{}

func WithCtxRequestContext(ctx context.Context, c RequestContext) context.Context {
	return context.WithValue(ctx, ctxRequestContextKey{}, c)
}

func FromCtxRequestContext(ctx context.Context) (RequestContext, bool) {
	if val := ctx.Value(ctxRequestContextKey{}); val != nil {
		if t, ok := val.(RequestContext); ok {
			return t, true
		}
	}
	return nil, false
}

type ctxRequestIdKey struct{}

func WithCtxRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, ctxRequestIdKey{}, requestId)
}

func FromCtxRequestId(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxRequestIdKey{}); val != nil {
		if t, ok := val.(string); ok {
			return t, true
		}
	}
	return "", false
}

func MustFromCtxRequestId(ctx context.Context) string {
	requestId, ok := FromCtxRequestId(ctx)
	if ok {
		return requestId
	}
	return ""
}

type ctxClientRequestIdKey struct{}

func WithCtxClientRequestId(ctx context.Context, clientRequestId string) context.Context {
	return context.WithValue(ctx, ctxClientRequestIdKey{}, clientRequestId)
}

func FromCtxClientRequestId(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxClientRequestIdKey{}); val != nil {
		if t, ok := val.(string); ok {
			return t, true
		}
	}
	return "", false
}

func MustFromCtxClientRequestId(ctx context.Context) string {
	clientRequestId, ok := FromCtxClientRequestId(ctx)
	if ok {
		return clientRequestId
	}
	return ""
}

type ctxModelNameKey struct{}

func WithCtxModelName(ctx context.Context, modelName string) context.Context {
	return context.WithValue(ctx, ctxModelNameKey{}, modelName)
}

func FromCtxModelName(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxModelNameKey{}); val != nil {
		if modelName, ok := val.(string); ok {
			return modelName, true
		}
	}
	return "", false
}

func MustFromCtxModelName(ctx context.Context) string {
	modelName, ok := FromCtxModelName(ctx)
	if ok {
		return modelName
	}
	return ""
}

type ctxModelVersionKey struct{}

func WithCtxModelVersion(ctx context.Context, modelVersion string) context.Context {
	return context.WithValue(ctx, ctxModelVersionKey{}, modelVersion)
}

func FromCtxModelVersion(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxModelVersionKey{}); val != nil {
		if modelVersion, ok := val.(string); ok {
			return modelVersion, true
		}
	}
	return "", false
}

func MustFromCtxModelVersion(ctx context.Context) string {
	modelVersion, ok := FromCtxModelVersion(ctx)
	if ok {
		return modelVersion
	}
	return ""
}

type ctxAccountIdKey struct{}

func WithCtxAccountId(ctx context.Context, accountId string) context.Context {
	return context.WithValue(ctx, ctxAccountIdKey{}, accountId)
}

func FromCtxAccountId(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxAccountIdKey{}); val != nil {
		if accountId, ok := val.(string); ok {
			return accountId, true
		}
	}
	return "", false
}

func MustFromCtxAccountId(ctx context.Context) string {
	accountId, ok := FromCtxAccountId(ctx)
	if ok {
		return accountId
	}
	return ""
}

type ctxUserIdKey struct{}

func WithCtxUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, ctxUserIdKey{}, userId)
}

func FromCtxUserId(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxUserIdKey{}); val != nil {
		if userId, ok := val.(string); ok {
			return userId, true
		}
	}
	return "", false
}

func MustFromCtxUserId(ctx context.Context) string {
	userId, ok := FromCtxUserId(ctx)
	if ok {
		return userId
	}
	return ""
}

type ctxGuidanceKey struct{}

func WithCtxGuidance(ctx context.Context, guidance bool) context.Context {
	return context.WithValue(ctx, ctxGuidanceKey{}, guidance)
}

func FromCtxGuidance(ctx context.Context) bool {
	if val := ctx.Value(ctxGuidanceKey{}); val != nil {
		if guidance, ok := val.(bool); ok {
			return guidance
		}
	}
	return false
}

type ctxModerationSceneKey struct{}

func WithCtxModerationScene(ctx context.Context, scene string) context.Context {
	return context.WithValue(ctx, ctxModerationSceneKey{}, scene)
}

func FromCtxModerationScene(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxModerationSceneKey{}); val != nil {
		if scene, ok := val.(string); ok {
			return scene, true
		}
	}
	return "", false
}

func MustFromCtxModerationScene(ctx context.Context) string {
	scene, ok := FromCtxModerationScene(ctx)
	if ok {
		return scene
	}
	return ""
}

type ctxRollingStatusKey struct{}

func WithCtxRollingStatus(ctx context.Context, status string) context.Context {
	return context.WithValue(ctx, ctxRollingStatusKey{}, status)
}

func FromCtxRollingStatus(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxRollingStatusKey{}); val != nil {
		if scene, ok := val.(string); ok {
			return scene, true
		}
	}
	return "", false
}

func MustFromCtxRollingStatus(ctx context.Context) string {
	scene, ok := FromCtxRollingStatus(ctx)
	if ok {
		return scene
	}
	return ""
}

type ctxOriginServiceTypeKey struct{}

func WithCtxOriginServiceType(ctx context.Context, originServiceType string) context.Context {
	return context.WithValue(ctx, ctxOriginServiceTypeKey{}, originServiceType)
}

func FromCtxOriginServiceType(ctx context.Context) (string, bool) {
	if val := ctx.Value(ctxOriginServiceTypeKey{}); val != nil {
		if scene, ok := val.(string); ok {
			return scene, true
		}
	}
	return "", false
}

type ctxAllowDataCollected struct{}

func WithCtxAllowDataCollected(ctx context.Context, allowDataCollected bool) context.Context {
	return context.WithValue(ctx, ctxAllowDataCollected{}, allowDataCollected)
}

func MustFromCtxAllowDataCollected(ctx context.Context) bool {
	if val := ctx.Value(ctxAllowDataCollected{}); val != nil {
		if allowDataCollected, ok := val.(bool); ok {
			return allowDataCollected
		}
	}
	return false
}
