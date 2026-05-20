package action

import "testing"

func TestReplaceSubstringAndRegexSkipsFrozenProxies(t *testing.T) {
	ctx := testContextWithProxies("HK 01", "US 01", "HK 02")
	ctx.AllProxies()[1].Frozen = true

	if err := parseTestFormula(t, "replace HK HongKong\nreplace \"/(\\d+)$/\" \"#${1}\"\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "HongKong #01", "US 01", "HongKong #02")
}
