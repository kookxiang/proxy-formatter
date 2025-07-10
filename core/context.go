package core

import (
	"context"
	"net/http"
)

type ExecuteContext struct {
	context context.Context
	items   []*ProxyItem
	output  string
	header  map[string]string
	Cancel  *context.CancelFunc
}

func NewExecuteContext() *ExecuteContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExecuteContext{
		context: ctx,
		items:   []*ProxyItem{},
		header:  map[string]string{},
		Cancel:  &cancel,
	}
}

func (ctx *ExecuteContext) RegisterProxy(proxy *ProxyItem) {
	ctx.items = append(ctx.items, proxy)
}

func (ctx *ExecuteContext) FilterProxy(filter func(*ProxyItem) bool) {
	filtered := make([]*ProxyItem, 0, len(ctx.items))
	for _, proxy := range ctx.items {
		if proxy.Frozen || filter(proxy) {
			filtered = append(filtered, proxy)
		}
	}
	ctx.items = filtered
}

func (ctx *ExecuteContext) Proxies() []*ProxyItem {
	result := make([]*ProxyItem, 0, len(ctx.items))
	for _, proxy := range ctx.items {
		if !proxy.Frozen {
			result = append(result, proxy)
		}
	}
	return result
}

func (ctx *ExecuteContext) AllProxies() []*ProxyItem {
	return ctx.items
}

func (ctx *ExecuteContext) SetContentType(contentType string) {
	ctx.WriteHeader("Content-Type", contentType)
}

func (ctx *ExecuteContext) Write(input string) {
	ctx.output += input
}

func (ctx *ExecuteContext) WriteHeader(key, value string) {
	ctx.header[key] = value
}

func (ctx *ExecuteContext) Pipe(writer http.ResponseWriter) error {
	for key, value := range ctx.header {
		writer.Header().Set(key, value)
	}
	_, err := writer.Write([]byte(ctx.output))
	return err
}
