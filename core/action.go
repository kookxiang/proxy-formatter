package core

import (
	"fmt"
)

type Action interface {
	Execute(ctx *ExecuteContext) error
}

var actionMap = map[string]func(params []string) (Action, error){}

func RegisterAction(name string, constructor func(params []string) (Action, error)) {
	actionMap[name] = constructor
}

func CreateAction(params []string) (Action, error) {
	name := params[0]
	if constructor, ok := actionMap[name]; ok {
		return constructor(params)
	} else {
		return nil, fmt.Errorf("action %s not found", name)
	}
}
