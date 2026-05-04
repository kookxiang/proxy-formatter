package action

import (
	"errors"
	"proxy-provider/core"
	"regexp"
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
	if strings.HasPrefix(replace.From, "/") && strings.HasSuffix(replace.From, "/") {
		regex, err := regexp.Compile(replace.From[1 : len(replace.From)-1])
		if err != nil {
			return err
		}
		for _, proxy := range ctx.Proxies() {
			proxy.Name = regex.ReplaceAllString(proxy.Name, replace.To)
		}
		return nil
	}

	for _, proxy := range ctx.Proxies() {
		if strings.Contains(proxy.Name, replace.From) {
			proxy.Name = strings.ReplaceAll(proxy.Name, replace.From, replace.To)
		}
	}
	return nil
}
