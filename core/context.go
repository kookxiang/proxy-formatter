package core

import (
	"errors"
	"net/http"
	"proxy-provider/util"

	"github.com/metacubex/mihomo/adapter/provider"
	"github.com/metacubex/mihomo/common/convert"
	"gopkg.in/yaml.v3"
)

type ExecuteContext struct {
	items               []*ProxyItem
	output              string
	Proxy               string
	AllowExternalScript bool
	ReqHeader           http.Header
	ResHeader           http.Header
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

func (ctx *ExecuteContext) RegisterProxiesFromBytes(body []byte) error {
	schema := &provider.ProxySchema{}
	if err := yaml.Unmarshal(body, schema); err != nil {
		proxies, err := convert.ConvertsV2Ray(body)
		if err != nil {
			return err
		}
		schema.Proxies = proxies
	}

	for _, proxy := range schema.Proxies {
		name, ok := proxy["name"].(string)
		if !ok || name == "" {
			return errors.New("proxy name is required in the schema")
		}
		option, err := util.ParseProxyOptions(proxy)
		if err != nil {
			return err
		}
		ctx.RegisterProxy(&ProxyItem{
			Name:   name,
			Type:   proxy["type"].(string),
			Option: option,
		})
	}

	return nil
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
