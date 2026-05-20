package action

import "testing"

func TestEmojiRemovesFlagPairs(t *testing.T) {
	ctx := testContextWithProxies("🇭🇰 HK 01", "US 🇺🇸 01", "No Flag")

	if err := parseTestFormula(t, "emoji\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "HK 01", "US01", "No Flag")
}
