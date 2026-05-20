package cache

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"proxy-provider/core"
	"strings"
	"testing"
	"time"
)

func withTempHTTPCacheFolder(t *testing.T) {
	t.Helper()
	original := HTTPCacheFolder
	HTTPCacheFolder = t.TempDir()
	t.Cleanup(func() {
		HTTPCacheFolder = original
	})
}

func TestBuildRequestCacheKeySortsHeadersAndValues(t *testing.T) {
	requestA := httptest.NewRequest(http.MethodGet, "https://example.com/sub", nil)
	requestA.Header.Add("X-B", "2")
	requestA.Header.Add("X-A", "b")
	requestA.Header.Add("X-A", "a")

	requestB := httptest.NewRequest(http.MethodGet, "https://example.com/sub", nil)
	requestB.Header.Add("X-A", "a")
	requestB.Header.Add("X-A", "b")
	requestB.Header.Add("X-B", "2")

	if got, want := buildRequestCacheKey(requestA), buildRequestCacheKey(requestB); got != want {
		t.Fatalf("expected equivalent requests to share cache key\nA:\n%s\nB:\n%s", got, want)
	}
}

func TestSaveAndGetResponseCacheRoundTrip(t *testing.T) {
	withTempHTTPCacheFolder(t)

	header := http.Header{}
	header.Add("X-Test", "one")
	header.Add("X-Test", "two")
	saveResponseCache("key", []byte("line1\nline2\n"), header)

	body, gotHeader, needRefresh := getResponseCache("key")
	if needRefresh {
		t.Fatal("expected freshly saved cache to not need refresh")
	}
	if string(body) != "line1\nline2\n" {
		t.Fatalf("expected cached body to round-trip, got %q", string(body))
	}
	if got := gotHeader.Values("X-Test"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("expected repeated headers to round-trip, got %v", got)
	}
}

func TestGetResponseCacheMarksOldFilesForRefresh(t *testing.T) {
	withTempHTTPCacheFolder(t)

	saveResponseCache("key", []byte("cached"), http.Header{})
	path := getResponseCachePath("key")
	old := time.Now().Add(-11 * time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	_, _, needRefresh := getResponseCache("key")
	if !needRefresh {
		t.Fatal("expected old cache file to need refresh")
	}
}

func TestRequestWithCacheSavesAndReusesResponse(t *testing.T) {
	withTempHTTPCacheFolder(t)

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set(core.HeaderSubscriptionUserinfo, "upload=1; download=2; total=3")
		w.Header().Set("X-Request-Number", "first")
		_, _ = w.Write([]byte("proxy-body"))
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, header, err := RequestWithCache(core.NewExecuteContext(), request, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "proxy-body" {
		t.Fatalf("expected response body from server, got %q", string(body))
	}
	if header.Get(core.HeaderSubscriptionUserinfo) == "" {
		t.Fatal("expected subscription userinfo header on non-prefer cache request")
	}

	body, header, err = RequestWithCache(core.NewExecuteContext(), request, true)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("expected second request to use cache, server saw %d requests", requests)
	}
	if string(body) != "proxy-body\n" {
		t.Fatalf("expected cached body with scanner newline, got %q", string(body))
	}
	if header.Get(core.HeaderSubscriptionUserinfo) != "" {
		t.Fatal("expected preferCache request to strip subscription userinfo")
	}
}

func TestRequestWithCacheRefreshesStaleCacheWhenNotPreferred(t *testing.T) {
	withTempHTTPCacheFolder(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fresh"))
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	cacheKey := buildRequestCacheKey(request)
	saveResponseCache(cacheKey, []byte("stale"), http.Header{})
	path := getResponseCachePath(cacheKey)
	old := time.Now().Add(-11 * time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	body, _, err := RequestWithCache(core.NewExecuteContext(), request, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "fresh" {
		t.Fatalf("expected stale cache to refresh, got %q", string(body))
	}
}

func TestRequestWithoutCacheRejectsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = requestWithoutCache(core.NewExecuteContext(), request)
	if err == nil {
		t.Fatal("expected non-OK response to fail")
	}
	if !strings.Contains(err.Error(), "status code: 418") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestRequestWithoutCacheRejectsInvalidProxyURL(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := core.NewExecuteContext()
	ctx.Proxy = "http://[::1"

	_, _, err = requestWithoutCache(ctx, request)
	if err == nil {
		t.Fatal("expected invalid proxy URL to fail")
	}
	if !strings.Contains(err.Error(), "invalid proxy URL") {
		t.Fatalf("expected invalid proxy URL error, got %v", err)
	}
}

func TestGetResponseCacheMiss(t *testing.T) {
	withTempHTTPCacheFolder(t)

	body, header, needRefresh := getResponseCache("missing")
	if body != nil || header != nil || needRefresh {
		t.Fatalf("expected cache miss zero values, got body=%q header=%v needRefresh=%v", string(body), header, needRefresh)
	}
}

func TestGetResponseCachePathUsesConfiguredFolder(t *testing.T) {
	withTempHTTPCacheFolder(t)

	path := getResponseCachePath("key")
	if filepath.Dir(path) != HTTPCacheFolder {
		t.Fatalf("expected cache path to use configured folder %q, got %q", HTTPCacheFolder, path)
	}
}
