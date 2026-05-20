package util

import (
	"reflect"
	"testing"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func TestParseProxyOptionsRequiresType(t *testing.T) {
	_, err := ParseProxyOptions(map[string]any{"name": "missing-type"})
	if err == nil {
		t.Fatal("expected missing type to fail")
	}
	if want := "missing type"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestParseProxyOptionsRejectsUnsupportedType(t *testing.T) {
	_, err := ParseProxyOptions(map[string]any{"type": "unknown"})
	if err == nil {
		t.Fatal("expected unsupported type to fail")
	}
	if want := "unsupport proxy type: unknown"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestParseProxyOptionsParsesShadowSocks(t *testing.T) {
	option, err := ParseProxyOptions(map[string]any{
		"name":     "ss",
		"type":     "ss",
		"server":   "127.0.0.1",
		"port":     8388,
		"cipher":   "aes-128-gcm",
		"password": "secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	ss, ok := option.(*outbound.ShadowSocksOption)
	if !ok {
		t.Fatalf("expected ShadowSocksOption, got %T", option)
	}
	if ss.Name != "ss" || ss.Server != "127.0.0.1" || ss.Port != 8388 || ss.Password != "secret" {
		t.Fatalf("unexpected shadowsocks option: %#v", ss)
	}
}

func TestParseProxyOptionsParsesCommonProxyTypes(t *testing.T) {
	tests := []struct {
		name     string
		mapping  map[string]any
		wantType any
	}{
		{
			name: "socks5",
			mapping: map[string]any{
				"name":   "socks",
				"type":   "socks5",
				"server": "127.0.0.1",
				"port":   1080,
			},
			wantType: &outbound.Socks5Option{},
		},
		{
			name: "http",
			mapping: map[string]any{
				"name":   "http",
				"type":   "http",
				"server": "127.0.0.1",
				"port":   8080,
			},
			wantType: &outbound.HttpOption{},
		},
		{
			name: "trojan",
			mapping: map[string]any{
				"name":     "trojan",
				"type":     "trojan",
				"server":   "127.0.0.1",
				"port":     443,
				"password": "secret",
			},
			wantType: &outbound.TrojanOption{},
		},
		{
			name: "vmess",
			mapping: map[string]any{
				"name":    "vmess",
				"type":    "vmess",
				"server":  "127.0.0.1",
				"port":    443,
				"uuid":    "00000000-0000-0000-0000-000000000000",
				"alterId": 0,
				"cipher":  "auto",
			},
			wantType: &outbound.VmessOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := ParseProxyOptions(tt.mapping)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(option) != reflect.TypeOf(tt.wantType) {
				t.Fatalf("expected %T, got %T", tt.wantType, option)
			}
		})
	}
}
