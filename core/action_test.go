package core

import (
	"errors"
	"testing"
)

type testCoreAction func(ctx *ExecuteContext) error

func (action testCoreAction) Execute(ctx *ExecuteContext) error {
	return action(ctx)
}

func registerTestAction(t *testing.T, name string, constructor func(params []string) (Action, error)) {
	t.Helper()
	original, existed := actionMap[name]
	RegisterAction(name, constructor)
	t.Cleanup(func() {
		if existed {
			actionMap[name] = original
		} else {
			delete(actionMap, name)
		}
	})
}

func TestCreateActionReturnsRegisteredAction(t *testing.T) {
	registerTestAction(t, "core-test-action", func(params []string) (Action, error) {
		if len(params) != 2 || params[1] != "value" {
			t.Fatalf("unexpected params: %v", params)
		}
		return testCoreAction(func(ctx *ExecuteContext) error {
			ctx.Write("ran")
			return nil
		}), nil
	})

	action, err := CreateAction([]string{"core-test-action", "value"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := NewExecuteContext()
	if err := action.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.output != "ran" {
		t.Fatalf("expected action to execute, got output %q", ctx.output)
	}
}

func TestCreateActionReturnsConstructorError(t *testing.T) {
	want := errors.New("bad params")
	registerTestAction(t, "core-test-error", func(params []string) (Action, error) {
		return nil, want
	})

	_, err := CreateAction([]string{"core-test-error"})
	if !errors.Is(err, want) {
		t.Fatalf("expected constructor error %v, got %v", want, err)
	}
}

func TestCreateActionRejectsUnknownAction(t *testing.T) {
	_, err := CreateAction([]string{"missing-core-test-action"})
	if err == nil {
		t.Fatal("expected unknown action to fail")
	}
	if want := "action missing-core-test-action not found"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}
