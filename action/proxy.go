package action

import (
	"errors"
	"proxy-provider/core"
)

func init() {
	core.RegisterAction("proxy", func(params []string) (core.Action, error) {
		if len(params) == 1 {
			return &SetProxyAction{Address: ""}, nil
		} else if len(params) == 2 {
			return &SetProxyAction{Address: params[1]}, nil
		} else {
			return nil, errors.New("invalid proxy rule format, expected 'proxy [address]'")
		}
	})
}

type SetProxyAction struct {
	Address string
}

func (action *SetProxyAction) Execute(ctx *core.ExecuteContext) error {
	ctx.Proxy = action.Address
	return nil
}
