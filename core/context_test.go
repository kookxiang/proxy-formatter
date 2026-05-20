package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func TestNewExecuteContextInitializesHeadersAndEmptyProxyList(t *testing.T) {
	ctx := NewExecuteContext()

	if ctx.ReqHeader.Get("User-Agent") != "clash.meta/1.19.11" {
		t.Fatalf("unexpected default User-Agent: %q", ctx.ReqHeader.Get("User-Agent"))
	}
	if ctx.ResHeader == nil {
		t.Fatal("expected response header to be initialized")
	}
	if len(ctx.AllProxies()) != 0 {
		t.Fatalf("expected no proxies, got %d", len(ctx.AllProxies()))
	}
}

func TestRegisterProxyAndProxyViewsRespectFrozenState(t *testing.T) {
	ctx := NewExecuteContext()
	active := &ProxyItem{Name: "active"}
	frozen := &ProxyItem{Name: "frozen", Frozen: true}
	ctx.RegisterProxy(active)
	ctx.RegisterProxy(frozen)

	if got := ctx.AllProxies(); len(got) != 2 || got[0] != active || got[1] != frozen {
		t.Fatalf("expected all proxies to include active and frozen items, got %#v", got)
	}
	if got := ctx.Proxies(); len(got) != 1 || got[0] != active {
		t.Fatalf("expected unfrozen proxies only, got %#v", got)
	}
}

func TestFilterProxyAlwaysKeepsFrozenProxies(t *testing.T) {
	ctx := NewExecuteContext()
	ctx.RegisterProxy(&ProxyItem{Name: "drop"})
	ctx.RegisterProxy(&ProxyItem{Name: "keep-frozen", Frozen: true})

	ctx.FilterProxy(func(proxy *ProxyItem) bool {
		return false
	})

	got := ctx.AllProxies()
	if len(got) != 1 || got[0].Name != "keep-frozen" {
		t.Fatalf("expected frozen proxy to remain, got %#v", got)
	}
}

func TestWriteAndPipeBodyAndHeaders(t *testing.T) {
	ctx := NewExecuteContext()
	ctx.ResHeader = http.Header{}
	ctx.ResHeader.Add("X-Test", "one")
	ctx.ResHeader.Add("X-Test", "two")
	ctx.Write("hello")
	ctx.Write(" world")

	recorder := httptest.NewRecorder()
	if err := ctx.Pipe(recorder); err != nil {
		t.Fatal(err)
	}

	if body := recorder.Body.String(); body != "hello world" {
		t.Fatalf("expected written body, got %q", body)
	}
	if got := recorder.Header().Values("X-Test"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("expected repeated headers to be piped, got %v", got)
	}
}

func TestRegisterProxiesFromBytesParsesClashSchema(t *testing.T) {
	ctx := NewExecuteContext()

	body := []byte(`
proxies:
  - name: ss-one
    type: ss
    server: 127.0.0.1
    port: 8388
    cipher: aes-128-gcm
    password: secret
`)
	if err := ctx.RegisterProxiesFromBytes(body); err != nil {
		t.Fatal(err)
	}

	proxies := ctx.AllProxies()
	if len(proxies) != 1 {
		t.Fatalf("expected one proxy, got %d", len(proxies))
	}
	if proxies[0].Name != "ss-one" || proxies[0].Type != "ss" {
		t.Fatalf("unexpected proxy metadata: %#v", proxies[0])
	}
	option, ok := proxies[0].Option.(*outbound.ShadowSocksOption)
	if !ok {
		t.Fatalf("expected shadowsocks option, got %T", proxies[0].Option)
	}
	if option.Server != "127.0.0.1" || option.Port != 8388 || option.Password != "secret" {
		t.Fatalf("unexpected shadowsocks option: %#v", option)
	}
}

func TestRegisterProxiesFromBytesRequiresProxyName(t *testing.T) {
	ctx := NewExecuteContext()

	body := []byte(`
proxies:
  - type: ss
    server: 127.0.0.1
    port: 8388
    cipher: aes-128-gcm
    password: secret
`)
	err := ctx.RegisterProxiesFromBytes(body)
	if err == nil {
		t.Fatal("expected proxy without name to fail")
	}
	if want := "proxy name is required in the schema"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestRegisterProxiesFromBytesReturnsErrorForInvalidInput(t *testing.T) {
	ctx := NewExecuteContext()

	if err := ctx.RegisterProxiesFromBytes([]byte("not a valid proxy subscription")); err == nil {
		t.Fatal("expected invalid proxy subscription to fail")
	}
}
