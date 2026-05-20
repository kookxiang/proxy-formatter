package cache

import (
	"net"
	"testing"
	"time"
)

func resetDNSCache(t *testing.T) {
	t.Helper()
	dnsCacheLock.Lock()
	original := dnsCacheStore
	dnsCacheStore = make(map[string]*DNSCacheItem)
	dnsCacheLock.Unlock()

	t.Cleanup(func() {
		dnsCacheLock.Lock()
		dnsCacheStore = original
		dnsCacheLock.Unlock()
	})
}

func TestDNSCacheStoresResolverHostPairsSeparately(t *testing.T) {
	resetDNSCache(t)

	SetDNSCache("resolver-a", "example.com", &DNSCacheItem{IPs: []net.IP{{192, 0, 2, 1}}})
	SetDNSCache("resolver-b", "example.com", &DNSCacheItem{IPs: []net.IP{{192, 0, 2, 2}}})

	if got := GetDNSCache("resolver-a", "example.com"); got == nil || !got.IPs[0].Equal(net.IP{192, 0, 2, 1}) {
		t.Fatalf("expected resolver-a cache item, got %#v", got)
	}
	if got := GetDNSCache("resolver-b", "example.com"); got == nil || !got.IPs[0].Equal(net.IP{192, 0, 2, 2}) {
		t.Fatalf("expected resolver-b cache item, got %#v", got)
	}
	if got := GetDNSCache("resolver-a", "missing.example.com"); got != nil {
		t.Fatalf("expected missing host cache miss, got %#v", got)
	}
}

func TestPurgeExpiredDNSCacheRecordKeepsFreshItems(t *testing.T) {
	resetDNSCache(t)

	dnsCacheStore["resolver@fresh.example.com"] = &DNSCacheItem{
		IPs:    []net.IP{{192, 0, 2, 1}},
		expire: time.Now().Add(time.Minute),
	}
	dnsCacheStore["resolver@expired.example.com"] = &DNSCacheItem{
		IPs:    []net.IP{{192, 0, 2, 2}},
		expire: time.Now().Add(-time.Minute),
	}

	purgeExpiredDNSCacheRecord()

	if got := GetDNSCache("resolver", "fresh.example.com"); got == nil {
		t.Fatal("expected fresh cache item to remain")
	}
	if got := GetDNSCache("resolver", "expired.example.com"); got != nil {
		t.Fatalf("expected expired cache item to be purged, got %#v", got)
	}
}
