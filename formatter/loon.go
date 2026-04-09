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
		quoteValue(option.Password),
		"fast-open=" + formatBool(option.TFO),
		"udp=" + formatBool(option.UDP),
	}

	switch option.Plugin {
	case "":
	case "obfs":
		mode := stringify(option.PluginOpts["mode"])
		if mode == "" {
			return "", fmt.Errorf("ss proxy %q plugin obfs missing mode", name)
		}
		fields = append(fields, "obfs-name="+mode)
		if host := stringify(option.PluginOpts["host"]); host != "" {
			fields = append(fields, "obfs-host="+host)
		}
		if path := stringify(option.PluginOpts["path"]); path != "" {
			fields = append(fields, "obfs-uri="+path)
		}
	case "shadow-tls":
		password := stringify(option.PluginOpts["password"])
		host := stringify(option.PluginOpts["host"])
		if password == "" || host == "" {
			return "", fmt.Errorf("ss proxy %q plugin shadow-tls missing password or host", name)
		}
		fields = append(fields,
			"shadow-tls-password="+quoteValue(password),
			"shadow-tls-sni="+host,
		)
		version := intValue(option.PluginOpts["version"])
		if version == 0 {
			version = 2
		}
		fields = append(fields, "shadow-tls-version="+strconv.Itoa(version))
		if udpPort := intValue(option.PluginOpts["udp-port"]); udpPort != 0 {
			fields = append(fields, "udp-port="+strconv.Itoa(udpPort))
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
		quoteValue(option.Password),
	}

	switch option.Network {
	case "", "tcp":
	case "ws":
		fields = append(fields, "transport=ws")
		if option.WSOpts.Path != "" {
			fields = append(fields, "path="+option.WSOpts.Path)
		}
		if host := headerValue(option.WSOpts.Headers, "Host"); host != "" {
			fields = append(fields, "host="+host)
		}
	default:
		return "", fmt.Errorf("trojan proxy %q network %q not implemented", name, option.Network)
	}

	if len(option.ALPN) > 0 {
		fields = append(fields, "alpn="+strings.Join(option.ALPN, ","))
	}
	fields = append(fields, "skip-cert-verify="+formatBool(option.SkipCertVerify))
	if option.SNI != "" {
		fields = append(fields, "sni="+option.SNI)
	}
	fields = append(fields, "udp="+formatBool(option.UDP))
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
		quoteValue(option.UUID),
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

	fields = append(fields, "over-tls="+formatBool(option.TLS))
	if option.ServerName != "" {
		fields = append(fields, "sni="+option.ServerName)
	}
	if option.TLS || option.RealityOpts.PublicKey != "" {
		fields = append(fields, "skip-cert-verify="+formatBool(option.SkipCertVerify))
	}

	if option.RealityOpts.PublicKey != "" {
		fields = append(fields, "public-key="+quoteValue(option.RealityOpts.PublicKey))
		if option.RealityOpts.ShortID != "" {
			fields = append(fields, "short-id="+option.RealityOpts.ShortID)
		}
	}

	fields = append(fields, "udp="+formatBool(option.UDP))
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
		quoteValue(option.UUID),
		"transport=" + transport,
		"alterId=" + strconv.Itoa(option.AlterID),
	}

	if path != "" {
		fields = append(fields, "path="+path)
	}
	if host != "" {
		fields = append(fields, "host="+host)
	}

	fields = append(fields, "over-tls="+formatBool(option.TLS))
	if option.ServerName != "" {
		fields = append(fields, "sni="+option.ServerName)
	}
	if option.TLS {
		fields = append(fields, "skip-cert-verify="+formatBool(option.SkipCertVerify))
	}
	fields = append(fields, "udp="+formatBool(option.UDP))
	return strings.Join(fields, ","), nil
}

func (formatter *LoonFormatter) formatVLESSTransport(name string, option *outbound.VlessOption) (transport, path, host string, err error) {
	switch option.Network {
	case "", "tcp":
		return "tcp", "", "", nil
	case "ws":
		return "ws", option.WSOpts.Path, headerValue(option.WSOpts.Headers, "Host"), nil
	case "http":
		return "http", firstString(option.HTTPOpts.Path), firstHeaderValue(option.HTTPOpts.Headers, "Host"), nil
	case "h2":
		return "http", option.HTTP2Opts.Path, firstString(option.HTTP2Opts.Host), nil
	default:
		return "", "", "", fmt.Errorf("vless proxy %q network %q not implemented", name, option.Network)
	}
}

func (formatter *LoonFormatter) formatVMessTransport(name string, option *outbound.VmessOption) (transport, path, host string, err error) {
	switch option.Network {
	case "", "tcp":
		return "tcp", firstString(option.HTTPOpts.Path), firstHeaderValue(option.HTTPOpts.Headers, "Host"), nil
	case "ws":
		return "ws", option.WSOpts.Path, headerValue(option.WSOpts.Headers, "Host"), nil
	case "http":
		return "http", firstString(option.HTTPOpts.Path), firstHeaderValue(option.HTTPOpts.Headers, "Host"), nil
	case "h2":
		return "http", option.HTTP2Opts.Path, firstString(option.HTTP2Opts.Host), nil
	default:
		return "", "", "", fmt.Errorf("vmess proxy %q network %q not implemented", name, option.Network)
	}
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func quoteValue(value string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
	return `"` + escaped + `"`
}

func stringify(value any) string {
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

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		number, _ := strconv.Atoi(typed)
		return number
	default:
		return 0
	}
}

func headerValue(headers map[string]string, key string) string {
	for headerKey, value := range headers {
		if strings.EqualFold(headerKey, key) {
			return value
		}
	}
	return ""
}

func firstHeaderValue(headers map[string][]string, key string) string {
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) {
			return firstString(values)
		}
	}
	return ""
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
