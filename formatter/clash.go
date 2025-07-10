package formatter

import (
	"bytes"
	"proxy-provider/core"
	"proxy-provider/util"

	"gopkg.in/yaml.v3"
)

func init() {
	core.RegisterAction("clash", func(params []string) (core.Action, error) {
		return &ClashFormatter{}, nil
	})
}

type ClashFormatter struct{}

func (formatter *ClashFormatter) Execute(ctx *core.ExecuteContext) error {
	ctx.SetContentType("text/yaml; charset=utf-8")
	proxies := make([]map[string]any, 0, len(ctx.AllProxies()))
	for _, proxy := range ctx.AllProxies() {
		config := map[string]any{}
		config["name"] = proxy.Name
		config["type"] = proxy.Type
		options := util.EncodeProxyStruct(proxy.Option)
		for key, value := range options {
			if key == "name" {
				continue
			}
			config[key] = value
		}
		if config["sni"] == config["server"] {
			delete(config, "sni")
		}
		if config["udp-over-tcp-version"] == 1 {
			delete(config, "udp-over-tcp-version")
		}
		proxies = append(proxies, config)
	}
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	defer encoder.Close()
	encoder.SetIndent(2)
	err := encoder.Encode(map[string]any{"items": proxies})
	ctx.Write(buffer.String())
	return err
}
