package action

import (
	"proxy-provider/core"
	"testing"
)

func TestProxySetsAndClearsAddress(t *testing.T) {
	ctx := core.NewExecuteContext()

	if err := parseTestFormula(t, "proxy http://127.0.0.1:7890\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Proxy != "http://127.0.0.1:7890" {
		t.Fatalf("expected proxy address to be set, got %q", ctx.Proxy)
	}

	if err := parseTestFormula(t, "proxy\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Proxy != "" {
		t.Fatalf("expected proxy address to be cleared, got %q", ctx.Proxy)
	}
}
