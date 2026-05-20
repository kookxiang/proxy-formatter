package action

import "testing"

func TestPrefixAndSuffixToggleNames(t *testing.T) {
	ctx := testContextWithProxies("HK Node", "JP Node")

	if err := parseTestFormula(t, "prefix Auto\nsuffix VIP\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}
	requireProxyNames(t, ctx, "Auto HK Node VIP", "Auto JP Node VIP")

	if err := parseTestFormula(t, "prefix Auto\nsuffix VIP\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}
	requireProxyNames(t, ctx, "HK Node", "JP Node")
}
