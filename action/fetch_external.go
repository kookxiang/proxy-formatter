package action

import (
	"errors"
	"fmt"
	"os/exec"
	"proxy-provider/core"
)

func init() {
	core.RegisterAction("fetch-external", func(params []string) (core.Action, error) {
		if len(params) < 2 {
			return nil, errors.New("invalid action rule format, expected 'fetch-external <command> [args...]'")
		}
		return &FetchExternalAction{
			Command: params[1],
			Args:    params[2:],
		}, nil
	})
}

type FetchExternalAction struct {
	Command string
	Args    []string
}

func (action *FetchExternalAction) Execute(ctx *core.ExecuteContext) error {
	if !ctx.AllowExternalScript {
		return errors.New("fetch-external is disabled, start proxy-formatter with -allow-external-script to enable it")
	}

	fetchLock.Lock()
	defer fetchLock.Unlock()

	command := exec.Command(action.Command, action.Args...)
	output, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return fmt.Errorf("fetch-external command failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("fetch-external command failed: %w", err)
	}

	ctx.ResHeader = make(map[string][]string)
	return ctx.RegisterProxiesFromBytes(output)
}
