package util

import "testing"

type emptyStruct struct {
	Value string
}

func TestIsEmptyValue(t *testing.T) {
	var nilStringSlice []string
	var nilAnyMap map[string]any
	var nilPointer *emptyStruct

	tests := []struct {
		name string
		val  any
		want bool
	}{
		{name: "nil", val: nil, want: true},
		{name: "empty string", val: "", want: true},
		{name: "non-empty string", val: "value", want: false},
		{name: "false", val: false, want: true},
		{name: "true", val: true, want: false},
		{name: "zero int", val: 0, want: true},
		{name: "non-zero int", val: 1, want: false},
		{name: "zero int8", val: int8(0), want: true},
		{name: "non-zero int8", val: int8(1), want: false},
		{name: "zero int16", val: int16(0), want: true},
		{name: "non-zero int16", val: int16(1), want: false},
		{name: "zero int32", val: int32(0), want: true},
		{name: "non-zero int32", val: int32(1), want: false},
		{name: "zero int64", val: int64(0), want: true},
		{name: "non-zero int64", val: int64(1), want: false},
		{name: "zero uint", val: uint(0), want: true},
		{name: "non-zero uint", val: uint(1), want: false},
		{name: "zero uint8", val: uint8(0), want: true},
		{name: "non-zero uint8", val: uint8(1), want: false},
		{name: "zero uint16", val: uint16(0), want: true},
		{name: "non-zero uint16", val: uint16(1), want: false},
		{name: "zero uint32", val: uint32(0), want: true},
		{name: "non-zero uint32", val: uint32(1), want: false},
		{name: "zero uint64", val: uint64(0), want: true},
		{name: "non-zero uint64", val: uint64(1), want: false},
		{name: "zero uintptr", val: uintptr(0), want: true},
		{name: "non-zero uintptr", val: uintptr(1), want: false},
		{name: "zero float", val: 0.0, want: true},
		{name: "non-zero float", val: 0.1, want: false},
		{name: "zero float32", val: float32(0), want: true},
		{name: "non-zero float32", val: float32(0.1), want: false},
		{name: "zero float64", val: float64(0), want: true},
		{name: "non-zero float64", val: float64(0.1), want: false},
		{name: "empty map any", val: map[string]any{}, want: true},
		{name: "non-empty map any", val: map[string]any{"k": "v"}, want: false},
		{name: "empty map string", val: map[string]string{}, want: true},
		{name: "non-empty map string", val: map[string]string{"k": "v"}, want: false},
		{name: "empty map string slice", val: map[string][]string{}, want: true},
		{name: "non-empty map string slice", val: map[string][]string{"k": {}}, want: false},
		{name: "empty any slice", val: []any{}, want: true},
		{name: "non-empty any slice", val: []any{"v"}, want: false},
		{name: "empty string slice", val: []string{}, want: true},
		{name: "non-empty string slice", val: []string{"v"}, want: false},
		{name: "nil string slice", val: nilStringSlice, want: true},
		{name: "nil map any", val: nilAnyMap, want: true},
		{name: "typed nil pointer", val: nilPointer, want: false},
		{name: "empty struct", val: emptyStruct{}, want: false},
		{name: "array", val: [0]string{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyValue(tt.val); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
