package mwctx

import (
	"context"
	"fmt"
	"strings"
)

func NewMiddlewareCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, middlewareKey{}, make(middlewareChain))
}

// RegisterMiddleware check and add
func RegisterMiddleware(ctx context.Context, newMiddleware string, required []string) {
	chain := chainFromCtx(ctx)
	if chain == nil {
		panic("Middleware error: middlewareChain not found in context")
	}
	if _, missingDependencies := chain.Search(required...); len(missingDependencies) > 0 {
		panic(fmt.Sprintf("Middleware dependency error: '%s' middleware requires %s to be registered first. Missing: %s",
			newMiddleware,
			"'"+strings.Join(required, "', '")+"'",
			"'"+strings.Join(missingDependencies, "', '")+"'"))
	}
	err := chain.Add(newMiddleware, required)
	if err != nil {
		panic("Middleware error: " + err.Error())
	}
}
