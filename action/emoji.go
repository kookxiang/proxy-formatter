package action

import (
	"proxy-provider/core"
	"regexp"
)

func init() {
	core.RegisterAction("emoji", func(params []string) (core.Action, error) {
		return &EmojiRemoveAction{}, nil
	})
}

type EmojiRemoveAction struct{}

var emojiFlagRegex = regexp.MustCompile(`\s*[\x{1F1E6}-\x{1F1FF}]{2}\s*`)

func (action *EmojiRemoveAction) Execute(ctx *core.ExecuteContext) error {
	for _, proxy := range ctx.Proxies() {
		proxy.Name = emojiFlagRegex.ReplaceAllString(proxy.Name, "")
	}
	return nil
}
