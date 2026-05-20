package action

import (
	"proxy-provider/core"
	"testing"
)

func TestHeaderSetsAndDeletesRequestHeaders(t *testing.T) {
	ctx := core.NewExecuteContext()

	if err := parseTestFormula(t, "header User-Agent TestAgent\nheader X-Token secret\nheader X-Token\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}

	if got := ctx.ReqHeader.Get("User-Agent"); got != "TestAgent" {
		t.Fatalf("expected User-Agent TestAgent, got %q", got)
	}
	if got := ctx.ReqHeader.Get("X-Token"); got != "" {
		t.Fatalf("expected X-Token to be deleted, got %q", got)
	}
}
