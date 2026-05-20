package action

import (
	"proxy-provider/core"
	"reflect"
	"testing"
)

func testContextWithProxies(names ...string) *core.ExecuteContext {
	ctx := core.NewExecuteContext()
	for _, name := range names {
		ctx.RegisterProxy(&core.ProxyItem{Name: name, Type: "ss"})
	}
	return ctx
}

func proxyNames(proxies []*core.ProxyItem) []string {
	names := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		names = append(names, proxy.Name)
	}
	return names
}

func requireProxyNames(t *testing.T, ctx *core.ExecuteContext, want ...string) {
	t.Helper()
	if got := proxyNames(ctx.AllProxies()); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected proxy names %v, got %v", want, got)
	}
}
