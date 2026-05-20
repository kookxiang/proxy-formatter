package action

import (
	"net/http"
	"proxy-provider/core"
	"strconv"
	"testing"
	"time"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func TestCheckTrafficKeepsActiveSubscription(t *testing.T) {
	ctx := testContextWithProxies("HK 01")
	ctx.ResHeader = http.Header{}
	ctx.ResHeader.Set(core.HeaderSubscriptionUserinfo, "upload=10; download=20; total=100; expire="+strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))

	if err := (&CheckTrafficAction{}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "HK 01")
}

func TestCheckTrafficIgnoresMissingAndMalformedSubscriptionInfo(t *testing.T) {
	tests := []struct {
		name string
		info string
	}{
		{name: "missing"},
		{name: "malformed", info: "upload=bad; ignored; total=100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testContextWithProxies("HK 01")
			ctx.ResHeader = http.Header{}
			if tt.info != "" {
				ctx.ResHeader.Set(core.HeaderSubscriptionUserinfo, tt.info)
			}

			if err := (&CheckTrafficAction{}).Execute(ctx); err != nil {
				t.Fatal(err)
			}

			requireProxyNames(t, ctx, "HK 01")
		})
	}
}

func TestCheckTrafficReplacesExpiredSubscriptionWithFakeProxy(t *testing.T) {
	ctx := testContextWithProxies("HK 01")
	ctx.ResHeader = http.Header{}
	ctx.ResHeader.Set(core.HeaderSubscriptionUserinfo, "upload=10; download=20; total=100; expire=1")

	if err := (&CheckTrafficAction{}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "订阅已过期")
	proxy := ctx.AllProxies()[0]
	option, ok := proxy.Option.(*outbound.ShadowSocksOption)
	if !ok {
		t.Fatalf("expected fake proxy to use ShadowSocksOption, got %T", proxy.Option)
	}
	if option.Server != "127.0.0.1" || option.Port != 65535 {
		t.Fatalf("expected fake proxy endpoint 127.0.0.1:65535, got %s:%d", option.Server, option.Port)
	}
}

func TestCheckTrafficReplacesOverQuotaSubscriptionWithFakeProxy(t *testing.T) {
	ctx := testContextWithProxies("HK 01")
	ctx.ResHeader = http.Header{}
	ctx.ResHeader.Set(core.HeaderSubscriptionUserinfo, "upload=50; download=50; total=100")

	if err := (&CheckTrafficAction{}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "流量已用尽")
}
