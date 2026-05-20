package action

import "testing"

func TestFreezeProtectsExistingProxiesFromLaterActions(t *testing.T) {
	ctx := testContextWithProxies("HK 01", "US 01")

	if err := parseTestFormula(t, "freeze\nreplace HK HongKong\ninclude HongKong\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "HK 01", "US 01")
	for _, proxy := range ctx.AllProxies() {
		if !proxy.Frozen {
			t.Fatalf("expected proxy %q to be frozen", proxy.Name)
		}
	}
}
