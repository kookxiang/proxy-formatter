package action

import (
	"errors"
	"fmt"
	"proxy-provider/core"
)

const maxExtendExecutions = 8

func init() {
	core.RegisterAction("extend", func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid extend rule format, expected 'extend <formula>'")
		}
		return &ExtendAction{Name: params[1]}, nil
	})
}

type ExtendAction struct {
	Name string
}

func (action *ExtendAction) Execute(ctx *core.ExecuteContext) error {
	if ctx.FormulaLoader == nil {
		return errors.New("extend action is unavailable because formula loader is not configured")
	}

	ctx.ExtendExecutionCount++
	if ctx.ExtendExecutionCount > maxExtendExecutions {
		return fmt.Errorf("extend execution limit exceeded: max %d", maxExtendExecutions)
	}

	formula, err := ctx.FormulaLoader(action.Name)
	if err != nil {
		return err
	}

	ctx.OutputSuppressionDepth++
	defer func() {
		ctx.OutputSuppressionDepth--
	}()

	return formula.Execute(ctx)
}
