package core

import (
	"net/http"
)

type ExecuteContext struct {
	items     []*ProxyItem
	output    string
	ReqHeader http.Header
	ResHeader http.Header
}

func NewExecuteContext() *ExecuteContext {
	ctx := &ExecuteContext{
		items:     []*ProxyItem{},
		ReqHeader: http.Header{},
		ResHeader: http.Header{},
	}
	ctx.ReqHeader.Set("User-Agent", "clash.meta/1.19.11")
	return ctx
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

func (ctx *ExecuteContext) Write(input string) {
	ctx.output += input
}

func (ctx *ExecuteContext) Pipe(writer http.ResponseWriter) error {
	for key, values := range ctx.ResHeader {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}
	_, err := writer.Write([]byte(ctx.output))
	return err
}
