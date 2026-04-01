package action

import (
	"strconv"
	"strings"
	"time"

	"proxy-provider/core"

	"github.com/metacubex/mihomo/adapter/outbound"
)

func init() {
	create := func(params []string) (core.Action, error) {
		return &CheckTrafficAction{}, nil
	}
	core.RegisterAction("check-traffic", create)
}

type CheckTrafficAction struct{}

type SubscriptionInfo struct {
	Upload   int64
	Download int64
	Total    int64
	Expire   int64
}

func parseSubscriptionInfo(info string) SubscriptionInfo {
	result := SubscriptionInfo{}
	for part := range strings.SplitSeq(info, ";") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "upload":
			result.Upload = val
		case "download":
			result.Download = val
		case "total":
			result.Total = val
		case "expire":
			result.Expire = val
		}
	}
	return result
}

func (action *CheckTrafficAction) Execute(ctx *core.ExecuteContext) error {
	info := ctx.ResHeader.Get(core.HeaderSubscriptionUserinfo)
	if info == "" {
		return nil
	}

	si := parseSubscriptionInfo(info)

	isExpired := si.Expire > 0 && time.Now().Unix() > si.Expire
	isOverQuota := si.Total > 0 && (si.Upload+si.Download) >= si.Total

	var reason string
	if isExpired {
		reason = "订阅已过期"
	} else if isOverQuota {
		reason = "流量已用尽"
	} else {
		return nil
	}

	// 清空代理列表
	ctx.FilterProxy(func(proxy *core.ProxyItem) bool {
		return false
	})

	// 增加一个假的代理服务器，避免部分客户端认为列表有问题不加载
	ctx.RegisterProxy(&core.ProxyItem{
		Name: reason,
		Type: "ss",
		Option: &outbound.ShadowSocksOption{
			Name:     reason,
			Server:   "127.0.0.1",
			Port:     65535,
			Password: "_FAKE_PASSWORD_",
			Cipher:   "aes-128-gcm",
		},
	})

	return nil
}
