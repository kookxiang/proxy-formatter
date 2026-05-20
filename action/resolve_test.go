package action

import (
	"net"
	"proxy-provider/cache"
	"proxy-provider/core"
	"testing"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func TestResolveUsesCacheAndPreservesOriginalHostForSNI(t *testing.T) {
	const dns = "test-dns"
	const host = "node.example.com"
	cache.SetDNSCache(dns, host, &cache.DNSCacheItem{IPs: []net.IP{{203, 0, 113, 10}}})

	ctx := core.NewExecuteContext()
	option := &outbound.TrojanOption{Server: host}
	ctx.RegisterProxy(&core.ProxyItem{Name: "trojan", Type: "trojan", Option: option})

	if err := (&ResolveHostAction{DNS: dns}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	if option.Server != "203.0.113.10" {
		t.Fatalf("expected server to be resolved from cache, got %q", option.Server)
	}
	if option.SNI != host {
		t.Fatalf("expected SNI to keep original host, got %q", option.SNI)
	}
}

func TestResolveHandlesCommonProxyOptionTypesFromCache(t *testing.T) {
	const dns = "test-dns-common"
	cache.SetDNSCache(dns, "ss.example.com", &cache.DNSCacheItem{IPs: []net.IP{{203, 0, 113, 1}}})
	cache.SetDNSCache(dns, "vmess.example.com", &cache.DNSCacheItem{IPs: []net.IP{{203, 0, 113, 2}}})
	cache.SetDNSCache(dns, "vless.example.com", &cache.DNSCacheItem{IPs: []net.IP{{203, 0, 113, 3}}})
	cache.SetDNSCache(dns, "tuic.example.com", &cache.DNSCacheItem{IPs: []net.IP{{203, 0, 113, 4}}})

	ss := &outbound.ShadowSocksOption{Server: "ss.example.com"}
	vmess := &outbound.VmessOption{Server: "vmess.example.com"}
	vless := &outbound.VlessOption{Server: "vless.example.com"}
	tuic := &outbound.TuicOption{Server: "tuic.example.com", DisableSni: true}

	ctx := core.NewExecuteContext()
	ctx.RegisterProxy(&core.ProxyItem{Name: "ss", Option: ss})
	ctx.RegisterProxy(&core.ProxyItem{Name: "vmess", Option: vmess})
	ctx.RegisterProxy(&core.ProxyItem{Name: "vless", Option: vless})
	ctx.RegisterProxy(&core.ProxyItem{Name: "tuic", Option: tuic})

	if err := (&ResolveHostAction{DNS: dns}).Execute(ctx); err != nil {
		t.Fatal(err)
	}

	if ss.Server != "203.0.113.1" {
		t.Fatalf("expected ss server to resolve, got %q", ss.Server)
	}
	if vmess.Server != "203.0.113.2" || vmess.ServerName != "vmess.example.com" {
		t.Fatalf("unexpected vmess option after resolve: %#v", vmess)
	}
	if vless.Server != "203.0.113.3" || vless.ServerName != "vless.example.com" {
		t.Fatalf("unexpected vless option after resolve: %#v", vless)
	}
	if tuic.Server != "203.0.113.4" || tuic.SNI != "tuic.example.com" || tuic.DisableSni {
		t.Fatalf("unexpected tuic option after resolve: %#v", tuic)
	}
}

func TestResolveDNSPrefersIPv4FromCache(t *testing.T) {
	const dns = "test-dns-ipv4"
	const host = "dualstack.example.com"
	cache.SetDNSCache(dns, host, &cache.DNSCacheItem{IPs: []net.IP{
		net.ParseIP("2001:db8::1"),
		net.IP{198, 51, 100, 10},
	}})

	got := (&ResolveHostAction{DNS: dns}).ResolveDNS(host)
	if got != "198.51.100.10" {
		t.Fatalf("expected cached IPv4 to be preferred, got %q", got)
	}
}
