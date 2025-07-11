package action

import (
	"errors"
	"proxy-provider/core"
)

func init() {
	core.RegisterAction("header", func(params []string) (core.Action, error) {
		if len(params) == 2 {
			// If only key is provided, set value to empty string
			return &RequestHeaderAction{Key: params[1], Value: ""}, nil
		} else if len(params) == 3 {
			return &RequestHeaderAction{Key: params[1], Value: params[2]}, nil
		} else {
			return nil, errors.New("invalid header rule format, expected 'header <key> [value]'")
		}
	})
}

type RequestHeaderAction struct {
	Key   string
	Value string
}

func (action *RequestHeaderAction) Execute(ctx *core.ExecuteContext) error {
	if action.Value == "" {
		ctx.ReqHeader.Del(action.Key)
	} else {
		ctx.ReqHeader.Set(action.Key, action.Value)
	}
	return nil
}
