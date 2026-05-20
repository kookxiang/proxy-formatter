package action

import "testing"

func TestFilterIncludeExcludeSubstringAndRegex(t *testing.T) {
	tests := []struct {
		name    string
		formula string
		want    []string
	}{
		{
			name:    "include substring",
			formula: "include HK\n",
			want:    []string{"HK 01", "HK 02"},
		},
		{
			name:    "exclude substring",
			formula: "exclude HK\n",
			want:    []string{"US 01"},
		},
		{
			name:    "include regex",
			formula: "include \"/^HK \\d+$/\"\n",
			want:    []string{"HK 01", "HK 02"},
		},
		{
			name:    "exclude regex",
			formula: "exclude \"/^HK \\d+$/\"\n",
			want:    []string{"US 01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testContextWithProxies("HK 01", "US 01", "HK 02")
			if err := parseTestFormula(t, tt.formula).Execute(ctx); err != nil {
				t.Fatal(err)
			}
			requireProxyNames(t, ctx, tt.want...)
		})
	}
}

func TestFilterKeepsFrozenProxies(t *testing.T) {
	ctx := testContextWithProxies("HK 01", "US 01", "HK 02")
	ctx.AllProxies()[1].Frozen = true

	if err := parseTestFormula(t, "include HK\n").Execute(ctx); err != nil {
		t.Fatal(err)
	}

	requireProxyNames(t, ctx, "HK 01", "US 01", "HK 02")
}
