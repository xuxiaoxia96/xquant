package mwctx

import (
	"context"
	"fmt"
)

type middlewareKey struct{}

type dependencies []string

type middlewareChain map[string]dependencies

func chainFromCtx(ctx context.Context) middlewareChain {
	val := ctx.Value(middlewareKey{})
	if chain, ok := val.(middlewareChain); ok {
		return chain
	}
	return nil
}

func (c middlewareChain) Has(name string) bool {
	_, ok := c[name]
	return ok
}

func (c middlewareChain) Add(name string, de dependencies) error {
	if c.Has(name) {
		return fmt.Errorf("middleware %s already exists", name)
	}
	c[name] = de
	return nil
}

func (c middlewareChain) Search(name ...string) (exist, missing []string) {
	for _, dep := range name {
		if c.Has(dep) {
			exist = append(exist, dep)
		} else {
			missing = append(missing, dep)
		}
	}
	return
}
