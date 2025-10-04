package api_utils

import "sync"

type RequestContext interface {
	Get(key string) (value interface{}, exists bool)
	Set(key string, value interface{})
}

func NewRequestContext() RequestContext {
	return &requestContext{}
}

type requestContext struct {
	// This mutex protect Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (ctx *requestContext) Set(key string, value interface{}) {
	ctx.mu.Lock()
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}

	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (ctx *requestContext) Get(key string) (value interface{}, exists bool) {
	ctx.mu.RLock()
	value, exists = ctx.Keys[key]
	ctx.mu.RUnlock()
	return
}
