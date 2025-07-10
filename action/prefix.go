package action

import (
	"proxy-provider/core"
	"strings"
)

func init() {
	core.RegisterAction("prefix", func(params []string) (core.Action, error) {
		if !strings.HasSuffix(params[1], " ") {
			params[1] += " "
		}
		return &PrefixAction{Content: params[1]}, nil
	})
	core.RegisterAction("suffix", func(params []string) (core.Action, error) {
		if !strings.HasPrefix(params[1], " ") {
			params[1] = " " + params[1]
		}
		return &PrefixAction{Content: params[1], isSuffix: true}, nil
	})
}

type PrefixAction struct {
	Content  string
	isSuffix bool
}

func (action *PrefixAction) Execute(ctx *core.ExecuteContext) error {
	for _, proxy := range ctx.Proxies() {
		if action.isSuffix {
			if strings.HasSuffix(proxy.Name, action.Content) {
				proxy.Name = strings.TrimSuffix(proxy.Name, action.Content)
			} else {
				proxy.Name = proxy.Name + action.Content
			}
		} else {
			if strings.HasPrefix(proxy.Name, action.Content) {
				proxy.Name = strings.TrimPrefix(proxy.Name, action.Content)
			} else {
				proxy.Name = action.Content + proxy.Name
			}
		}
	}
	return nil
}
