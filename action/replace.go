package action

import (
	"errors"
	"proxy-provider/core"
	"strings"
)

func init() {
	core.RegisterAction("replace", func(params []string) (core.Action, error) {
		if len(params) != 3 {
			return nil, errors.New("invalid replace rule format, expected 'replace <from> <to>'")
		}
		return &ReplaceAction{From: params[1], To: params[2]}, nil
	})
}

type ReplaceAction struct {
	From string
	To   string
}

func (replace *ReplaceAction) Execute(ctx *core.ExecuteContext) error {
	for _, proxy := range ctx.Proxies() {
		if strings.Contains(proxy.Name, replace.From) {
			proxy.Name = strings.ReplaceAll(proxy.Name, replace.From, replace.To)
		}
	}
	return nil
}
