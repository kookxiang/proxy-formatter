package formatter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"proxy-provider/cache"
	"proxy-provider/core"
	"sort"
	"strconv"
	"strings"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func init() {
	core.RegisterAction("sing-box", func(params []string) (core.Action, error) {
		if len(params) > 2 {
			return nil, errors.New("invalid fetch rule format, expected 'sing-box [base-config-url]'")
		}
		if len(params) == 2 {
			return &SingBoxFormatter{BaseConfig: params[1]}, nil
		}
		return &SingBoxFormatter{}, nil
	})
}

type SingBoxFormatter struct {
	BaseConfig string
}

func (formatter *SingBoxFormatter) Execute(ctx *core.ExecuteContext) error {
	ctx.ResHeader.Set("Content-Type", "application/json; charset=utf-8")

	outbounds := make([]map[string]any, 0, len(ctx.AllProxies()))
	for _, proxy := range ctx.AllProxies() {
		config, err := formatter.formatProxy(proxy)
		if err != nil {
			return err
		}
		outbounds = append(outbounds, config)
	}

	config := map[string]any{}
	var err error
	if formatter.BaseConfig != "" {
		config, err = formatter.loadBaseConfig(ctx)
		if err != nil {
			return err
		}
	}

	config["outbounds"] = formatter.mergeOutbounds(config["outbounds"], outbounds)
	config["outbounds"] = formatter.attachProxyTagsToGroups(config["outbounds"], outbounds)

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	ctx.Write(string(jsonData))
	ctx.Write("\n")
	return nil
}

func (formatter *SingBoxFormatter) loadBaseConfig(ctx *core.ExecuteContext) (map[string]any, error) {
	urlObj, err := url.Parse(formatter.BaseConfig)
	if err != nil {
		return nil, err
	}

	request := &http.Request{
		Method: http.MethodGet,
		URL:    urlObj,
		Header: ctx.ReqHeader,
	}

	body, _, err := cache.RequestWithCache(ctx, request, false)
	if err != nil {
		return nil, err
	}

	config := map[string]any{}
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("invalid sing-box base config: %w", err)
	}

	return config, nil
}

func (formatter *SingBoxFormatter) mergeOutbounds(existing any, generated []map[string]any) []any {
	result := make([]any, 0)
	tagIndex := map[string]int{}

	if existingList, ok := existing.([]any); ok {
		for _, item := range existingList {
			if outboundMap, ok := item.(map[string]any); ok {
				if tag, ok := outboundMap["tag"].(string); ok && tag != "" {
					tagIndex[tag] = len(result)
				}
			}
			result = append(result, item)
		}
	}

	for _, outboundConfig := range generated {
		tag, _ := outboundConfig["tag"].(string)
		if tag != "" {
			if index, ok := tagIndex[tag]; ok {
				result[index] = outboundConfig
				continue
			}
			tagIndex[tag] = len(result)
		}
		result = append(result, outboundConfig)
	}

	return result
}

func (formatter *SingBoxFormatter) attachProxyTagsToGroups(existing any, generated []map[string]any) []any {
	existingList, ok := existing.([]any)
	if !ok {
		return nil
	}

	generatedTags := make([]string, 0, len(generated))
	for _, outboundConfig := range generated {
		tag, _ := outboundConfig["tag"].(string)
		if tag != "" {
			generatedTags = append(generatedTags, tag)
		}
	}
	if len(generatedTags) == 0 {
		return existingList
	}

	for _, item := range existingList {
		outboundMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		outboundType, _ := outboundMap["type"].(string)
		if outboundType != "selector" && outboundType != "urltest" {
			continue
		}

		outboundMap["outbounds"] = formatter.expandGroupMembers(outboundMap["outbounds"], generatedTags)
	}

	return existingList
}

func (formatter *SingBoxFormatter) formatProxy(proxy *core.ProxyItem) (map[string]any, error) {
	switch option := proxy.Option.(type) {
	case *outbound.ShadowSocksOption:
		return formatter.formatShadowSocks(proxy.Name, option)
	case *outbound.TrojanOption:
		return formatter.formatTrojan(proxy.Name, option)
	case *outbound.VlessOption:
		return formatter.formatVLESS(proxy.Name, option)
	default:
		return nil, fmt.Errorf("not implemented proxy type: %T", option)
	}
}

func (formatter *SingBoxFormatter) formatShadowSocks(name string, option *outbound.ShadowSocksOption) (map[string]any, error) {
	snippet := map[string]any{
		"type":        "shadowsocks",
		"tag":         name,
		"server":      option.Server,
		"server_port": option.Port,
		"method":      option.Cipher,
		"password":    option.Password,
	}

	if !option.UDP {
		snippet["network"] = "tcp"
	}

	if option.Plugin != "" {
		pluginName, pluginOpts, err := formatter.formatShadowsocksPlugin(name, option)
		if err != nil {
			return nil, err
		}
		snippet["plugin"] = pluginName
		snippet["plugin_opts"] = pluginOpts
	}

	if option.UDPOverTCP {
		if option.UDPOverTCPVersion != 0 {
			snippet["udp_over_tcp"] = map[string]any{
				"enabled": true,
				"version": option.UDPOverTCPVersion,
			}
		} else {
			snippet["udp_over_tcp"] = true
		}
	}

	return snippet, nil
}

func (formatter *SingBoxFormatter) formatShadowsocksPlugin(name string, option *outbound.ShadowSocksOption) (string, string, error) {
	switch option.Plugin {
	case "obfs":
		mode := formatter.stringify(option.PluginOpts["mode"])
		if mode == "" {
			return "", "", fmt.Errorf("shadowsocks proxy %q plugin obfs missing mode", name)
		}
		args := map[string]string{
			"obfs": mode,
		}
		if host := formatter.stringify(option.PluginOpts["host"]); host != "" {
			args["obfs-host"] = host
		}
		return "obfs-local", formatter.serializePluginOpts(args), nil
	default:
		return "", "", fmt.Errorf("shadowsocks proxy %q plugin %q not implemented", name, option.Plugin)
	}
}

func (formatter *SingBoxFormatter) formatTrojan(name string, option *outbound.TrojanOption) (map[string]any, error) {
	snippet := map[string]any{
		"type":        "trojan",
		"tag":         name,
		"server":      option.Server,
		"server_port": option.Port,
		"password":    option.Password,
		"tls":         formatter.formatTLS(option.SNI, option.Server, option.SkipCertVerify, option.ALPN, option.RealityOpts),
	}

	if !option.UDP {
		snippet["network"] = "tcp"
	}

	transport, err := formatter.formatTrojanTransport(name, option)
	if err != nil {
		return nil, err
	}
	if transport != nil {
		snippet["transport"] = transport
	}

	return snippet, nil
}

func (formatter *SingBoxFormatter) formatTrojanTransport(name string, option *outbound.TrojanOption) (map[string]any, error) {
	switch option.Network {
	case "", "tcp":
		return nil, nil
	case "ws":
		transport := map[string]any{
			"type": "ws",
		}
		if option.WSOpts.Path != "" {
			transport["path"] = option.WSOpts.Path
		}
		if len(option.WSOpts.Headers) > 0 {
			transport["headers"] = option.WSOpts.Headers
		}
		return transport, nil
	default:
		return nil, fmt.Errorf("trojan proxy %q network %q not implemented", name, option.Network)
	}
}

func (formatter *SingBoxFormatter) formatVLESS(name string, option *outbound.VlessOption) (map[string]any, error) {
	snippet := map[string]any{
		"type":        "vless",
		"tag":         name,
		"server":      option.Server,
		"server_port": option.Port,
		"uuid":        option.UUID,
	}

	if option.Flow != "" {
		snippet["flow"] = option.Flow
	}
	if !option.UDP {
		snippet["network"] = "tcp"
	}
	if option.TLS || option.RealityOpts.PublicKey != "" || option.ServerName != "" || option.SkipCertVerify || len(option.ALPN) > 0 {
		snippet["tls"] = formatter.formatTLS(option.ServerName, option.Server, option.SkipCertVerify, option.ALPN, option.RealityOpts)
	}

	transport, err := formatter.formatVLESSTransport(name, option)
	if err != nil {
		return nil, err
	}
	if transport != nil {
		snippet["transport"] = transport
	}

	return snippet, nil
}

func (formatter *SingBoxFormatter) formatVLESSTransport(name string, option *outbound.VlessOption) (map[string]any, error) {
	switch option.Network {
	case "", "tcp":
		return nil, nil
	case "ws":
		transport := map[string]any{
			"type": "ws",
		}
		if option.WSOpts.Path != "" {
			transport["path"] = option.WSOpts.Path
		}
		if len(option.WSOpts.Headers) > 0 {
			transport["headers"] = option.WSOpts.Headers
		}
		return transport, nil
	case "http":
		transport := map[string]any{
			"type": "http",
		}
		if path := formatter.firstString(option.HTTPOpts.Path); path != "" {
			transport["path"] = path
		}
		if host := formatter.firstHeaderValue(option.HTTPOpts.Headers, "Host"); host != "" {
			transport["host"] = []string{host}
		}
		if method := option.HTTPOpts.Method; method != "" {
			transport["method"] = method
		}
		if len(option.HTTPOpts.Headers) > 0 {
			headers := map[string][]string{}
			for key, values := range option.HTTPOpts.Headers {
				if strings.EqualFold(key, "Host") {
					continue
				}
				headers[key] = values
			}
			if len(headers) > 0 {
				transport["headers"] = headers
			}
		}
		return transport, nil
	default:
		return nil, fmt.Errorf("vless proxy %q network %q not implemented", name, option.Network)
	}
}

func (formatter *SingBoxFormatter) formatTLS(serverName, server string, insecure bool, alpn []string, reality outbound.RealityOptions) map[string]any {
	tlsConfig := map[string]any{
		"enabled": true,
	}

	if serverName != "" && serverName != server {
		tlsConfig["server_name"] = serverName
	}
	if insecure {
		tlsConfig["insecure"] = true
	}
	if len(alpn) > 0 {
		tlsConfig["alpn"] = alpn
	}
	if reality.PublicKey != "" {
		tlsConfig["reality"] = map[string]any{
			"enabled":    true,
			"public_key": reality.PublicKey,
		}
		if reality.ShortID != "" {
			tlsConfig["reality"].(map[string]any)["short_id"] = reality.ShortID
		}
	}

	return tlsConfig
}

func (formatter *SingBoxFormatter) serializePluginOpts(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := values[key]
		if value == "" {
			continue
		}
		if value == "true" {
			parts = append(parts, key)
			continue
		}
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ";")
}

func (formatter *SingBoxFormatter) stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int8:
		return strconv.Itoa(int(typed))
	case int16:
		return strconv.Itoa(int(typed))
	case int32:
		return strconv.Itoa(int(typed))
	case int64:
		return strconv.FormatInt(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

func (formatter *SingBoxFormatter) firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (formatter *SingBoxFormatter) firstHeaderValue(headers map[string][]string, key string) string {
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) {
			return formatter.firstString(values)
		}
	}
	return ""
}

func (formatter *SingBoxFormatter) expandGroupMembers(existing any, generatedTags []string) []string {
	result := make([]string, 0)
	seen := map[string]struct{}{}

	for _, pattern := range formatter.toStringList(existing) {
		matched := false
		for _, tag := range generatedTags {
			if !strings.Contains(tag, pattern) {
				continue
			}
			matched = true
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			result = append(result, tag)
		}
		if matched || pattern == "" {
			continue
		}
		if _, ok := seen[pattern]; ok {
			continue
		}
		seen[pattern] = struct{}{}
		result = append(result, pattern)
	}

	return result
}

func (formatter *SingBoxFormatter) toStringList(values any) []string {
	result := make([]string, 0)
	switch typed := values.(type) {
	case []string:
		for _, value := range typed {
			if value != "" {
				result = append(result, value)
			}
		}
	case []any:
		for _, item := range typed {
			value, ok := item.(string)
			if ok && value != "" {
				result = append(result, value)
			}
		}
	}
	return result
}
