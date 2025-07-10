package action

import "proxy-provider/core"

func init() {
	core.RegisterAction("freeze", func(params []string) (core.Action, error) {
		return &FreezeConfigAction{}, nil
	})
}

type FreezeConfigAction struct{}

func (action FreezeConfigAction) Execute(ctx *core.ExecuteContext) error {
	for _, proxy := range ctx.Proxies() {
		proxy.Frozen = true
	}
	return nil
}
