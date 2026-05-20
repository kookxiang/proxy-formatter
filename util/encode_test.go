package util

import (
	"reflect"
	"testing"
)

type encodeEmbedded struct {
	Embedded string `proxy:"embedded,omitempty"`
	Ignored  string `proxy:"ignored,omitempty"`
}

type encodeNested struct {
	Value string `proxy:"value,omitempty"`
	Empty string `proxy:"empty,omitempty"`
}

type encodeSample struct {
	encodeEmbedded
	Name   string            `proxy:"name,omitempty"`
	Port   int               `proxy:"port,omitempty"`
	Zero   int               `proxy:"zero,omitempty"`
	Nested encodeNested      `proxy:"nested,omitempty"`
	List   []encodeNested    `proxy:"list,omitempty"`
	Map    map[string]string `proxy:"map,omitempty"`
	Ptr    *encodeNested     `proxy:"ptr,omitempty"`
	NilPtr *encodeNested     `proxy:"nil_ptr,omitempty"`
}

type encodeOverrideEmbedded struct {
	Name string `proxy:"name,omitempty"`
}

type encodeOverrideSample struct {
	encodeOverrideEmbedded
	Name string `proxy:"name,omitempty"`
}

type encodeInterfaceSample struct {
	Value any `proxy:"value,omitempty"`
	Nil   any `proxy:"nil,omitempty"`
}

type encodeDeepLeaf struct {
	Name string `proxy:"name,omitempty"`
	Zero int    `proxy:"zero,omitempty"`
}

type encodeDeepBranch struct {
	Leaf       encodeDeepLeaf            `proxy:"leaf,omitempty"`
	LeafPtr    *encodeDeepLeaf           `proxy:"leaf_ptr,omitempty"`
	LeafIface  any                       `proxy:"leaf_iface,omitempty"`
	NilIface   any                       `proxy:"nil_iface,omitempty"`
	Leaves     []encodeDeepLeaf          `proxy:"leaves,omitempty"`
	LeafPtrs   []*encodeDeepLeaf         `proxy:"leaf_ptrs,omitempty"`
	LeafMap    map[string]encodeDeepLeaf `proxy:"leaf_map,omitempty"`
	IfaceMap   map[string]any            `proxy:"iface_map,omitempty"`
	NestedList [][]encodeDeepLeaf        `proxy:"nested_list,omitempty"`
	EmptyList  []encodeDeepLeaf          `proxy:"empty_list,omitempty"`
	EmptyMap   map[string]encodeDeepLeaf `proxy:"empty_map,omitempty"`
}

type encodeDeepSample struct {
	Branch    encodeDeepBranch  `proxy:"branch,omitempty"`
	BranchPtr *encodeDeepBranch `proxy:"branch_ptr,omitempty"`
}

func TestEncodeProxyStructEncodesTagsAndOmitsEmptyValues(t *testing.T) {
	input := &encodeSample{
		encodeEmbedded: encodeEmbedded{Embedded: "embedded"},
		Name:           "node",
		Port:           8388,
		Nested:         encodeNested{Value: "nested"},
		List:           []encodeNested{{Value: "first"}, {}},
		Map:            map[string]string{"key": "value"},
		Ptr:            &encodeNested{Value: "ptr"},
	}

	got := EncodeProxyStruct(input)
	want := map[string]any{
		"embedded": "embedded",
		"name":     "node",
		"port":     8388,
		"nested":   map[string]any{"value": "nested"},
		"list":     []any{map[string]any{"value": "first"}},
		"map":      map[string]any{"key": "value"},
		"ptr":      map[string]any{"value": "ptr"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected encoded struct %#v, got %#v", want, got)
	}
}

func TestEncodeProxyStructRejectsNonStructInput(t *testing.T) {
	tests := []any{nil, "not-struct", (*encodeSample)(nil)}
	for _, tt := range tests {
		if got := EncodeProxyStruct(tt); len(got) != 0 {
			t.Fatalf("expected empty map for %#v, got %#v", tt, got)
		}
	}
}

func TestEncodeProxyStructRecursivelyPrunesEmptyValues(t *testing.T) {
	leaf := &encodeDeepLeaf{Name: "ptr"}
	input := encodeDeepSample{
		Branch: encodeDeepBranch{
			Leaf:      encodeDeepLeaf{Name: "leaf"},
			LeafPtr:   leaf,
			LeafIface: &encodeDeepLeaf{Name: "iface"},
			Leaves: []encodeDeepLeaf{
				{},
				{Name: "slice"},
				{Zero: 0},
			},
			LeafPtrs: []*encodeDeepLeaf{
				nil,
				{Name: "ptr-slice"},
				{},
			},
			LeafMap: map[string]encodeDeepLeaf{
				"empty": {},
				"full":  {Name: "map"},
			},
			IfaceMap: map[string]any{
				"empty_string": "",
				"nil":          nil,
				"struct":       encodeDeepLeaf{Name: "iface-map"},
				"ptr":          &encodeDeepLeaf{Name: "iface-ptr-map"},
			},
			NestedList: [][]encodeDeepLeaf{
				{{}},
				{{Name: "nested"}},
			},
			EmptyList: []encodeDeepLeaf{{}},
			EmptyMap:  map[string]encodeDeepLeaf{"empty": {}},
		},
		BranchPtr: &encodeDeepBranch{},
	}

	got := EncodeProxyStruct(input)
	want := map[string]any{
		"branch": map[string]any{
			"leaf":       map[string]any{"name": "leaf"},
			"leaf_ptr":   map[string]any{"name": "ptr"},
			"leaf_iface": map[string]any{"name": "iface"},
			"leaves":     []any{map[string]any{"name": "slice"}},
			"leaf_ptrs":  []any{map[string]any{"name": "ptr-slice"}},
			"leaf_map":   map[string]any{"full": map[string]any{"name": "map"}},
			"iface_map": map[string]any{
				"struct": map[string]any{"name": "iface-map"},
				"ptr":    map[string]any{"name": "iface-ptr-map"},
			},
			"nested_list": []any{
				[]any{map[string]any{"name": "nested"}},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected encoded recursive struct %#v, got %#v", want, got)
	}
}

func TestEncodeProxyStructUnwrapsInterfacesAndPointers(t *testing.T) {
	var nested any = &encodeInterfaceSample{
		Value: any(&encodeNested{Value: "wrapped"}),
		Nil:   any((*encodeNested)(nil)),
	}

	got := EncodeProxyStruct(nested)
	want := map[string]any{
		"value": map[string]any{"value": "wrapped"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected interface and pointer values to unwrap to %#v, got %#v", want, got)
	}
}

func TestEncodeProxyStructConcreteFieldOverridesEmbeddedField(t *testing.T) {
	got := EncodeProxyStruct(encodeOverrideSample{
		encodeOverrideEmbedded: encodeOverrideEmbedded{Name: "embedded"},
		Name:                   "outer",
	})

	want := map[string]any{"name": "outer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected concrete field to override embedded field %#v, got %#v", want, got)
	}
}
