package action

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"proxy-provider/core"
	"strconv"
	"testing"

	_ "proxy-provider/formatter"
)

func init() {
	core.RegisterAction("test-write", func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid test-write rule format")
		}
		return testWriteAction(params[1]), nil
	})
	core.RegisterAction("test-require-depth", func(params []string) (core.Action, error) {
		if len(params) != 2 {
			return nil, errors.New("invalid test-require-depth rule format")
		}
		depth, err := strconv.Atoi(params[1])
		if err != nil {
			return nil, err
		}
		return testRequireDepthAction(depth), nil
	})
}

type testWriteAction string

func (action testWriteAction) Execute(ctx *core.ExecuteContext) error {
	ctx.Write(string(action))
	return nil
}

type testRequireDepthAction int

func (action testRequireDepthAction) Execute(ctx *core.ExecuteContext) error {
	if ctx.OutputSuppressionDepth != int(action) {
		return fmt.Errorf("expected output suppression depth %d, got %d", action, ctx.OutputSuppressionDepth)
	}
	return nil
}

func parseTestFormula(t *testing.T, content string) *core.Formula {
	t.Helper()
	formula := &core.Formula{Name: "test"}
	if err := formula.Parse([]byte(content)); err != nil {
		t.Fatal(err)
	}
	return formula
}

func testLoader(t *testing.T, formulas map[string]string) func(name string) (*core.Formula, error) {
	t.Helper()
	return func(name string) (*core.Formula, error) {
		content, ok := formulas[name]
		if !ok {
			return nil, fmt.Errorf("formula %q not found", name)
		}
		formula := &core.Formula{Name: name}
		if err := formula.Parse([]byte(content)); err != nil {
			return nil, err
		}
		return formula, nil
	}
}

func executeAndBody(t *testing.T, formula *core.Formula, ctx *core.ExecuteContext) string {
	t.Helper()
	if err := formula.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	if err := ctx.Pipe(recorder); err != nil {
		t.Fatal(err)
	}
	return recorder.Body.String()
}

func TestExtendExecutesInheritedFormulaOnSameContext(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{
		"base": "test-write inherited\n",
	})

	body := executeAndBody(t, parseTestFormula(t, "extend base\ntest-write child\n"), ctx)
	if body != "inheritedchild" {
		t.Fatalf("expected inherited action and child action to write, got %q", body)
	}
	if ctx.OutputSuppressionDepth != 0 {
		t.Fatalf("expected suppression depth to be restored, got %d", ctx.OutputSuppressionDepth)
	}
}

func TestExtendSuppressesFormatterOutputAndHeaders(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{
		"base": "surge\n",
	})

	formula := parseTestFormula(t, "extend base\ntest-write child\n")
	if err := formula.Execute(ctx); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	if err := ctx.Pipe(recorder); err != nil {
		t.Fatal(err)
	}
	if body := recorder.Body.String(); body != "child" {
		t.Fatalf("expected only child output, got %q", body)
	}
	if contentType := ctx.ResHeader.Get("Content-Type"); contentType != "" {
		t.Fatalf("expected suppressed formatter to skip headers, got Content-Type %q", contentType)
	}
}

func TestExtendRestoresNestedOutputSuppressionDepth(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{
		"a": "test-require-depth 1\nextend b\ntest-require-depth 1\n",
		"b": "test-require-depth 2\n",
	})

	if err := parseTestFormula(t, "extend a\ntest-require-depth 0\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.OutputSuppressionDepth != 0 {
		t.Fatalf("expected suppression depth to be restored, got %d", ctx.OutputSuppressionDepth)
	}
}

func TestExtendExecutionLimitRejectsCycles(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{
		"a": "extend a\n",
	})

	err := parseTestFormula(t, "extend a\n").Execute(ctx)
	if err == nil {
		t.Fatal("expected cyclic extend to fail")
	}
	if want := "extend execution limit exceeded: max 8"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
	if ctx.OutputSuppressionDepth != 0 {
		t.Fatalf("expected suppression depth to be restored, got %d", ctx.OutputSuppressionDepth)
	}
}

func TestExtendExecutionLimitCountsSequentialExtends(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{
		"base": "",
	})

	err := parseTestFormula(t, "extend base\nextend base\nextend base\nextend base\nextend base\nextend base\nextend base\nextend base\nextend base\n").Execute(ctx)
	if err == nil {
		t.Fatal("expected ninth extend to fail")
	}
	if want := "extend execution limit exceeded: max 8"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestExtendReturnsLoaderErrors(t *testing.T) {
	ctx := core.NewExecuteContext()
	ctx.FormulaLoader = testLoader(t, map[string]string{})

	err := parseTestFormula(t, "extend missing\n").Execute(ctx)
	if err == nil {
		t.Fatal("expected missing formula to fail")
	}
	if want := `formula "missing" not found`; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}
