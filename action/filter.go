package action

import (
	"errors"
	"proxy-provider/core"
	"regexp"
	"strings"
)

func init() {
	create := func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid filter rule format, expected '<include|exclude> pattern'")
		}
		return &FilterAction{
			isExclude: params[0] == "exclude",
			Pattern:   params[1],
		}, nil
	}
	core.RegisterAction("include", create)
	core.RegisterAction("exclude", create)
}

type FilterAction struct {
	isExclude bool
	Pattern   string
}

func (action *FilterAction) Execute(ctx *core.ExecuteContext) error {
	if strings.HasPrefix(action.Pattern, "/") && strings.HasSuffix(action.Pattern, "/") {
		regex, err := regexp.Compile(action.Pattern[1 : len(action.Pattern)-1])
		if err != nil {
			return err
		}
		ctx.FilterProxy(func(proxy *core.ProxyItem) bool {
			return regex.MatchString(proxy.Name) == !action.isExclude
		})
	} else {
		ctx.FilterProxy(func(proxy *core.ProxyItem) bool {
			return strings.Contains(proxy.Name, action.Pattern) == !action.isExclude
		})
	}
	return nil
}
