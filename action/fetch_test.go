package action

import (
	"net/http"
	"net/http/httptest"
	"proxy-provider/cache"
	"proxy-provider/core"
	"testing"
)

func withTempFetchCacheFolder(t *testing.T) {
	t.Helper()
	original := cache.HTTPCacheFolder
	cache.HTTPCacheFolder = t.TempDir()
	t.Cleanup(func() {
		cache.HTTPCacheFolder = original
	})
}

func TestFetchRegistersProxiesAndSanitizesResponseHeaders(t *testing.T) {
	withTempFetchCacheFolder(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "request-header" {
			t.Fatalf("expected request header to be forwarded, got %q", got)
		}
		w.Header().Set(core.HeaderSubscriptionUserinfo, "upload=1; download=2; total=3")
		w.Header().Set("Content-Disposition", "attachment")
		_, _ = w.Write([]byte(`
proxies:
  - name: fetched
    type: ss
    server: 127.0.0.1
    port: 8388
    cipher: aes-128-gcm
    password: secret
`))
	}))
	defer server.Close()

	ctx := core.NewExecuteContext()
	ctx.ReqHeader.Set("X-Test", "request-header")
	if err := (&FetchAction{URL: server.URL}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "fetched")
	if ctx.ResHeader.Get(core.HeaderSubscriptionUserinfo) == "" {
		t.Fatal("expected subscription userinfo to be retained for fetch")
	}
	if got := ctx.ResHeader.Get("Content-Disposition"); got != "" {
		t.Fatalf("expected Content-Disposition to be stripped, got %q", got)
	}
}

func TestFetchOnceStripsSubscriptionUserinfo(t *testing.T) {
	withTempFetchCacheFolder(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(core.HeaderSubscriptionUserinfo, "upload=1; download=2; total=3")
		_, _ = w.Write([]byte(`
proxies:
  - name: fetched-once
    type: ss
    server: 127.0.0.1
    port: 8388
    cipher: aes-128-gcm
    password: secret
`))
	}))
	defer server.Close()

	ctx := core.NewExecuteContext()
	if err := (&FetchAction{URL: server.URL, Once: true}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "fetched-once")
	if got := ctx.ResHeader.Get(core.HeaderSubscriptionUserinfo); got != "" {
		t.Fatalf("expected subscription userinfo to be stripped, got %q", got)
	}
}

func TestFetchRejectsInvalidURL(t *testing.T) {
	err := (&FetchAction{URL: "http://[::1"}).Execute(core.NewExecuteContext())
	if err == nil {
		t.Fatal("expected invalid URL to fail")
	}
}
