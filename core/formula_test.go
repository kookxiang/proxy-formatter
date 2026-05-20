package core

import (
	"errors"
	"testing"
)

func TestFormulaParseSkipsBlankLinesAndComments(t *testing.T) {
	var calls []string
	registerTestAction(t, "core-test-record", func(params []string) (Action, error) {
		if len(params) != 2 {
			t.Fatalf("unexpected params: %v", params)
		}
		value := params[1]
		return testCoreAction(func(ctx *ExecuteContext) error {
			calls = append(calls, value)
			return nil
		}), nil
	})

	formula := &Formula{}
	err := formula.Parse([]byte("\n# comment\n  core-test-record first\n\ncore-test-record second\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := formula.Execute(NewExecuteContext()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Fatalf("expected two recorded calls, got %v", calls)
	}
}

func TestFormulaParseQuotedAndEscapedParams(t *testing.T) {
	var got []string
	registerTestAction(t, "core-test-params", func(params []string) (Action, error) {
		got = append([]string(nil), params...)
		return testCoreAction(func(ctx *ExecuteContext) error { return nil }), nil
	})

	formula := &Formula{}
	err := formula.Parse([]byte(`core-test-params "hello world" "quote: \" ok" keep\slash` + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 params, got %v", got)
	}
	want := []string{"core-test-params", "hello world", `quote: " ok`, `keep\slash`}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("param %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestFormulaParseActionWithoutTrailingNewline(t *testing.T) {
	var executed bool
	registerTestAction(t, "core-test-no-newline", func(params []string) (Action, error) {
		if len(params) != 1 {
			t.Fatalf("unexpected params: %v", params)
		}
		return testCoreAction(func(ctx *ExecuteContext) error {
			executed = true
			return nil
		}), nil
	})

	formula := &Formula{}
	if err := formula.Parse([]byte("core-test-no-newline")); err != nil {
		t.Fatal(err)
	}
	if err := formula.Execute(NewExecuteContext()); err != nil {
		t.Fatal(err)
	}
	if !executed {
		t.Fatal("expected action without trailing newline to execute")
	}
}

func TestFormulaParseEmptyInput(t *testing.T) {
	formula := &Formula{}
	if err := formula.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if len(formula.actions) != 0 {
		t.Fatalf("expected no actions, got %d", len(formula.actions))
	}
}

func TestFormulaParseWrapsActionErrorsWithLineNumber(t *testing.T) {
	registerTestAction(t, "core-test-parse-error", func(params []string) (Action, error) {
		return nil, errors.New("invalid test action")
	})

	formula := &Formula{}
	err := formula.Parse([]byte("\ncore-test-parse-error\n"))
	if err == nil {
		t.Fatal("expected parse error")
	}
	if want := "formula parse failed: invalid test action at line 2"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestFormulaExecuteStopsOnFirstError(t *testing.T) {
	executedAfterError := false
	registerTestAction(t, "core-test-fail", func(params []string) (Action, error) {
		return testCoreAction(func(ctx *ExecuteContext) error {
			return errors.New("execute failed")
		}), nil
	})
	registerTestAction(t, "core-test-after-fail", func(params []string) (Action, error) {
		return testCoreAction(func(ctx *ExecuteContext) error {
			executedAfterError = true
			return nil
		}), nil
	})

	formula := &Formula{}
	if err := formula.Parse([]byte("core-test-fail\ncore-test-after-fail\n")); err != nil {
		t.Fatal(err)
	}
	err := formula.Execute(NewExecuteContext())
	if err == nil {
		t.Fatal("expected execute error")
	}
	if executedAfterError {
		t.Fatal("expected formula execution to stop after first error")
	}
}
