package util

import (
	"errors"
	"fmt"

	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/common/structure"
)

func ParseProxyOptions(mapping map[string]any) (any, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "proxy", WeaklyTypedInput: true, KeyReplacer: structure.DefaultKeyReplacer})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return nil, errors.New("missing type")
	}

	switch proxyType {
	case "ss":
		ssOption := &outbound.ShadowSocksOption{}
		err := decoder.Decode(mapping, ssOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewShadowSocks(*ssOption)
		defer proxy.Close()
		return ssOption, err
	case "ssr":
		ssrOption := &outbound.ShadowSocksROption{}
		err := decoder.Decode(mapping, ssrOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewShadowSocksR(*ssrOption)
		defer proxy.Close()
		return ssrOption, err
	case "socks5":
		socksOption := &outbound.Socks5Option{}
		err := decoder.Decode(mapping, socksOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewSocks5(*socksOption)
		defer proxy.Close()
		return socksOption, err
	case "http":
		httpOption := &outbound.HttpOption{}
		err := decoder.Decode(mapping, httpOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewHttp(*httpOption)
		defer proxy.Close()
		return httpOption, err
	case "vmess":
		vmessOption := &outbound.VmessOption{}
		err := decoder.Decode(mapping, vmessOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewVmess(*vmessOption)
		defer proxy.Close()
		return vmessOption, err
	case "vless":
		vlessOption := &outbound.VlessOption{}
		err := decoder.Decode(mapping, vlessOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewVless(*vlessOption)
		defer proxy.Close()
		return vlessOption, err
	case "snell":
		snellOption := &outbound.SnellOption{}
		err := decoder.Decode(mapping, snellOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewSnell(*snellOption)
		defer proxy.Close()
		return snellOption, err
	case "trojan":
		trojanOption := &outbound.TrojanOption{}
		err := decoder.Decode(mapping, trojanOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewTrojan(*trojanOption)
		defer proxy.Close()
		return trojanOption, err
	case "hysteria":
		hyOption := &outbound.HysteriaOption{}
		err := decoder.Decode(mapping, hyOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewHysteria(*hyOption)
		defer proxy.Close()
		return hyOption, err
	case "hysteria2":
		hyOption := &outbound.Hysteria2Option{}
		err := decoder.Decode(mapping, hyOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewHysteria2(*hyOption)
		defer proxy.Close()
		return hyOption, err
	case "wireguard":
		wgOption := &outbound.WireGuardOption{}
		err := decoder.Decode(mapping, wgOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewWireGuard(*wgOption)
		defer proxy.Close()
		return wgOption, err
	case "tuic":
		tuicOption := &outbound.TuicOption{}
		err := decoder.Decode(mapping, tuicOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewTuic(*tuicOption)
		defer proxy.Close()
		return tuicOption, err
	case "ssh":
		sshOption := &outbound.SshOption{}
		err := decoder.Decode(mapping, sshOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewSsh(*sshOption)
		defer proxy.Close()
		return sshOption, err
	case "mieru":
		mieruOption := &outbound.MieruOption{}
		err := decoder.Decode(mapping, mieruOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewMieru(*mieruOption)
		defer proxy.Close()
		return mieruOption, err
	case "anytls":
		anytlsOption := &outbound.AnyTLSOption{}
		err := decoder.Decode(mapping, anytlsOption)
		if err != nil {
			return nil, err
		}
		proxy, err := outbound.NewAnyTLS(*anytlsOption)
		defer proxy.Close()
		return anytlsOption, err
	default:
		return nil, fmt.Errorf("unsupport proxy type: %s", proxyType)
	}
}
