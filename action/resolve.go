package action

import (
	"context"
	"errors"
	"github.com/metacubex/mihomo/adapter/outbound"
	"net"
	"proxy-provider/cache"
	"proxy-provider/core"
	"time"
)

func init() {
	core.RegisterAction("resolve", func(params []string) (core.Action, error) {
		if len(params) > 2 {
			return nil, errors.New("too many parameters")
		} else if len(params) == 2 {
			return &ResolveHostAction{DNS: params[1]}, nil
		} else {
			return &ResolveHostAction{DNS: "119.29.29.29"}, nil
		}
	})
}

type ResolveHostAction struct {
	DNS string
}

func (action *ResolveHostAction) Execute(ctx *core.ExecuteContext) error {
	for _, proxy := range ctx.Proxies() {
		switch config := proxy.Option.(type) {
		case *outbound.ShadowSocksOption:
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.ShadowSocksROption:
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.VmessOption:
			if config.ServerName == "" {
				config.ServerName = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.VlessOption:
			if config.ServerName == "" {
				config.ServerName = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.TrojanOption:
			if config.SNI == "" {
				config.SNI = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.AnyTLSOption:
			if config.SNI == "" {
				config.SNI = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.HysteriaOption:
			if config.SNI == "" {
				config.SNI = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.Hysteria2Option:
			if config.SNI == "" {
				config.SNI = config.Server
			}
			config.Server = action.ResolveDNS(config.Server)
		case *outbound.TuicOption:
			if config.SNI == "" {
				config.SNI = config.Server
			}
			config.DisableSni = false
			config.Server = action.ResolveDNS(config.Server)
		default:
		}
	}
	return nil
}

func (action *ResolveHostAction) ResolveDNS(host string) string {
	item := cache.GetDNSCache(action.DNS, host)
	if item == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
		if err != nil || len(ips) == 0 {
			return host
		}
		item = &cache.DNSCacheItem{IPs: ips}
		cache.SetDNSCache(action.DNS, host, item)
	}

	// prefer IPv4 addresses
	for _, ip := range item.IPs {
		if len(ip) == net.IPv4len {
			return ip.String()
		}
	}
	return item.IPs[0].String()
}
