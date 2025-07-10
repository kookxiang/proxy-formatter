package formatter

import (
	"errors"
	"fmt"
	"proxy-provider/core"
	"proxy-provider/util"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func init() {
	core.RegisterAction("surge", func(params []string) (core.Action, error) {
		if len(params) > 2 {
			return nil, errors.New("invalid fetch rule format, expected 'surge [version]'")
		}
		if len(params) == 2 {
			return &SurgeFormatter{Version: params[1]}, nil
		} else {
			return &SurgeFormatter{}, nil
		}
	})
}

type SurgeFormatter struct {
	Version string
}

func (formatter *SurgeFormatter) Execute(ctx *core.ExecuteContext) error {
	ctx.SetContentType("text/plain; charset=utf-8")
	ctx.Write("[Proxy]")
	for _, proxy := range ctx.AllProxies() {
		ctx.Write("\n")
		switch value := proxy.Option.(type) {
		case *outbound.ShadowSocksOption:
			ctx.Write(fmt.Sprintf("%s = ss, %s, %d", proxy.Name, value.Server, value.Port))
			ctx.Write(formatter.formatShadowSocks(value))
		case *outbound.TrojanOption:
			ctx.Write(fmt.Sprintf("%s = trojan, %s, %d", proxy.Name, value.Server, value.Port))
			ctx.Write(formatter.formatTrojan(value))
		default:
			return errors.New("not implemented proxy type: " + fmt.Sprintf("%T", value))
		}
	}
	return nil
}

func (formatter *SurgeFormatter) formatExtraConfig(config map[string]any) string {
	result := ""
	for k, v := range config {
		if util.IsEmptyValue(v) {
			continue
		}
		result += fmt.Sprintf(", %s=%v", k, v)
	}
	return result
}

func (formatter *SurgeFormatter) formatShadowSocks(option *outbound.ShadowSocksOption) string {
	extraConfig := map[string]any{
		"encrypt-method": option.Cipher,
		"password":       option.Password,
		"udp":            option.UDP,
	}
	if option.Plugin == "obfs" {
		extraConfig["obfs"] = option.PluginOpts["mode"]
		extraConfig["obfs-host"] = option.PluginOpts["host"]
		extraConfig["obfs-uri"] = option.PluginOpts["path"]
	}
	return formatter.formatExtraConfig(extraConfig)
}

func (formatter *SurgeFormatter) formatTrojan(option *outbound.TrojanOption) string {
	extraConfig := map[string]any{
		"password":         option.Password,
		"skip-cert-verify": option.SkipCertVerify,
		"ws":               false,
		"ws-path":          option.WSOpts.Path,
		"ws-headers":       "",
	}
	for k, v := range option.WSOpts.Headers {
		if extraConfig["ws-headers"] != "" {
			extraConfig["ws-headers"] = extraConfig["ws-headers"].(string) + "|"
		}
		extraConfig["ws-headers"] = extraConfig["ws-headers"].(string) + fmt.Sprintf("%s=\"%s\"", k, v)
	}
	if !util.IsEmptyValue(extraConfig["ws-path"]) || !util.IsEmptyValue(extraConfig["ws-headers"]) {
		extraConfig["ws"] = true
	}
	if option.SNI != option.Server {
		extraConfig["sni"] = option.SNI
	}
	return formatter.formatExtraConfig(extraConfig)
}
