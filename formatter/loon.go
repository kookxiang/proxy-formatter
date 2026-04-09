package formatter

import (
	"fmt"
	"proxy-provider/core"
	"strconv"
	"strings"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func init() {
	core.RegisterAction("loon", func(params []string) (core.Action, error) {
		return &LoonFormatter{}, nil
	})
}

type LoonFormatter struct{}

func (formatter *LoonFormatter) Execute(ctx *core.ExecuteContext) error {
	ctx.ResHeader.Set("Content-Type", "text/plain; charset=utf-8")
	for index, proxy := range ctx.AllProxies() {
		if index > 0 {
			ctx.Write("\n")
		}
		line, err := formatter.formatProxy(proxy)
		if err != nil {
			return err
		}
		ctx.Write(line)
	}
	return nil
}

func (formatter *LoonFormatter) formatProxy(proxy *core.ProxyItem) (string, error) {
	switch option := proxy.Option.(type) {
	case *outbound.ShadowSocksOption:
		return formatter.formatShadowSocks(proxy.Name, option)
	case *outbound.TrojanOption:
		return formatter.formatTrojan(proxy.Name, option)
	case *outbound.VlessOption:
		return formatter.formatVLESS(proxy.Name, option)
	case *outbound.VmessOption:
		return formatter.formatVMess(proxy.Name, option)
	default:
		return "", fmt.Errorf("not implemented proxy type: %T", option)
	}
}

func (formatter *LoonFormatter) formatShadowSocks(name string, option *outbound.ShadowSocksOption) (string, error) {
	fields := []string{
		name + " = Shadowsocks",
		option.Server,
		strconv.Itoa(option.Port),
		option.Cipher,
		formatter.quoteValue(option.Password),
		"fast-open=" + formatter.formatBool(option.TFO),
		"udp=" + formatter.formatBool(option.UDP),
	}

	switch option.Plugin {
	case "":
	case "obfs":
		mode := formatter.stringify(option.PluginOpts["mode"])
		if mode == "" {
			return "", fmt.Errorf("ss proxy %q plugin obfs missing mode", name)
		}
		fields = append(fields, "obfs-name="+mode)
		if host := formatter.stringify(option.PluginOpts["host"]); host != "" {
			fields = append(fields, "obfs-host="+host)
		}
		if path := formatter.stringify(option.PluginOpts["path"]); path != "" {
			fields = append(fields, "obfs-uri="+path)
		}
	default:
		return "", fmt.Errorf("ss proxy %q plugin %q not implemented", name, option.Plugin)
	}

	return strings.Join(fields, ","), nil
}

func (formatter *LoonFormatter) formatTrojan(name string, option *outbound.TrojanOption) (string, error) {
	fields := []string{
		name + " = trojan",
		option.Server,
		strconv.Itoa(option.Port),
		formatter.quoteValue(option.Password),
	}

	switch option.Network {
	case "", "tcp":
	case "ws":
		fields = append(fields, "transport=ws")
		if option.WSOpts.Path != "" {
			fields = append(fields, "path="+option.WSOpts.Path)
		}
		if host := formatter.headerValue(option.WSOpts.Headers, "Host"); host != "" {
			fields = append(fields, "host="+host)
		}
	default:
		return "", fmt.Errorf("trojan proxy %q network %q not implemented", name, option.Network)
	}

	if len(option.ALPN) > 0 {
		fields = append(fields, "alpn="+strings.Join(option.ALPN, ","))
	}
	fields = append(fields, "skip-cert-verify="+formatter.formatBool(option.SkipCertVerify))
	if option.SNI != "" {
		fields = append(fields, "sni="+option.SNI)
	}
	fields = append(fields, "udp="+formatter.formatBool(option.UDP))
	return strings.Join(fields, ","), nil
}

func (formatter *LoonFormatter) formatVLESS(name string, option *outbound.VlessOption) (string, error) {
	transport, path, host, err := formatter.formatVLESSTransport(name, option)
	if err != nil {
		return "", err
	}

	fields := []string{
		name + " = VLESS",
		option.Server,
		strconv.Itoa(option.Port),
		formatter.quoteValue(option.UUID),
		"transport=" + transport,
	}

	if path != "" {
		fields = append(fields, "path="+path)
	}
	if host != "" {
		fields = append(fields, "host="+host)
	}
	if option.Flow != "" {
		fields = append(fields, "flow="+option.Flow)
	}

	fields = append(fields, "over-tls="+formatter.formatBool(option.TLS))
	if option.ServerName != "" {
		fields = append(fields, "sni="+option.ServerName)
	}
	if option.TLS || option.RealityOpts.PublicKey != "" {
		fields = append(fields, "skip-cert-verify="+formatter.formatBool(option.SkipCertVerify))
	}

	if option.RealityOpts.PublicKey != "" {
		fields = append(fields, "public-key="+formatter.quoteValue(option.RealityOpts.PublicKey))
		if option.RealityOpts.ShortID != "" {
			fields = append(fields, "short-id="+option.RealityOpts.ShortID)
		}
	}

	fields = append(fields, "udp="+formatter.formatBool(option.UDP))
	return strings.Join(fields, ","), nil
}

func (formatter *LoonFormatter) formatVMess(name string, option *outbound.VmessOption) (string, error) {
	transport, path, host, err := formatter.formatVMessTransport(name, option)
	if err != nil {
		return "", err
	}

	cipher := option.Cipher
	if cipher == "" {
		cipher = "auto"
	}

	fields := []string{
		name + " = vmess",
		option.Server,
		strconv.Itoa(option.Port),
		cipher,
		formatter.quoteValue(option.UUID),
		"transport=" + transport,
		"alterId=" + strconv.Itoa(option.AlterID),
	}

	if path != "" {
		fields = append(fields, "path="+path)
	}
	if host != "" {
		fields = append(fields, "host="+host)
	}

	fields = append(fields, "over-tls="+formatter.formatBool(option.TLS))
	if option.ServerName != "" {
		fields = append(fields, "sni="+option.ServerName)
	}
	if option.TLS {
		fields = append(fields, "skip-cert-verify="+formatter.formatBool(option.SkipCertVerify))
	}
	fields = append(fields, "udp="+formatter.formatBool(option.UDP))
	return strings.Join(fields, ","), nil
}

func (formatter *LoonFormatter) formatVLESSTransport(name string, option *outbound.VlessOption) (transport, path, host string, err error) {
	switch option.Network {
	case "", "tcp":
		return "tcp", "", "", nil
	case "ws":
		return "ws", option.WSOpts.Path, formatter.headerValue(option.WSOpts.Headers, "Host"), nil
	case "http":
		return "http", formatter.firstString(option.HTTPOpts.Path), formatter.firstHeaderValue(option.HTTPOpts.Headers, "Host"), nil
	default:
		return "", "", "", fmt.Errorf("vless proxy %q network %q not implemented", name, option.Network)
	}
}

func (formatter *LoonFormatter) formatVMessTransport(name string, option *outbound.VmessOption) (transport, path, host string, err error) {
	switch option.Network {
	case "", "tcp":
		return "tcp", "", "", nil
	case "ws":
		return "ws", option.WSOpts.Path, formatter.headerValue(option.WSOpts.Headers, "Host"), nil
	default:
		return "", "", "", fmt.Errorf("vmess proxy %q network %q not implemented", name, option.Network)
	}
}

func (formatter *LoonFormatter) formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (formatter *LoonFormatter) quoteValue(value string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
	return `"` + escaped + `"`
}

func (formatter *LoonFormatter) stringify(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

func (formatter *LoonFormatter) headerValue(headers map[string]string, key string) string {
	for headerKey, value := range headers {
		if strings.EqualFold(headerKey, key) {
			return value
		}
	}
	return ""
}

func (formatter *LoonFormatter) firstHeaderValue(headers map[string][]string, key string) string {
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) {
			return formatter.firstString(values)
		}
	}
	return ""
}

func (formatter *LoonFormatter) firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
