package models

import (
	"reflect"
	"testing"
)

func TestIntListScanFromInterfaceSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		want  IntList
	}{
		{"int64 elements", []interface{}{int64(1), int64(2), int64(3)}, IntList{1, 2, 3}},
		{"int32 elements", []interface{}{int32(4), int32(5)}, IntList{4, 5}},
		{"float64 elements", []interface{}{float64(7), float64(8)}, IntList{7, 8}},
		{"string elements", []interface{}{"10", "20"}, IntList{10, 20}},
		{"byte slice elements", []interface{}{[]byte("42")}, IntList{42}},
		{"empty slice", []interface{}{}, IntList{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var il IntList
			err := il.Scan(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(il, tt.want) {
				t.Errorf("got %v, want %v", il, tt.want)
			}
		})
	}
}

func TestIntListScanFromString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  IntList
	}{
		{"bracket notation", "[1,2,3]", IntList{1, 2, 3}},
		{"empty brackets", "[]", IntList{}},
		{"with spaces", "[ 1 , 2 , 3 ]", IntList{1, 2, 3}},
		{"single element", "[42]", IntList{42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var il IntList
			err := il.Scan(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(il, tt.want) {
				t.Errorf("got %v, want %v", il, tt.want)
			}
		})
	}
}

func TestIntListScanFromBytes(t *testing.T) {
	var il IntList
	err := il.Scan([]byte("[5,10,15]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := IntList{5, 10, 15}
	if !reflect.DeepEqual(il, want) {
		t.Errorf("got %v, want %v", il, want)
	}
}

func TestIntListScanNil(t *testing.T) {
	il := IntList{1, 2, 3}
	err := il.Scan(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if il != nil {
		t.Errorf("expected nil, got %v", il)
	}
}

func TestIntListScanErrors(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"unsupported type", 42},
		{"malformed string", "not a list"},
		{"malformed bytes", []byte("xyz")},
		{"bad element in interface slice", []interface{}{struct{}{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var il IntList
			if err := il.Scan(tt.input); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestIntListValueRoundTrip(t *testing.T) {
	original := IntList{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	v, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("Value() returned %T, want string", v)
	}

	var restored IntList
	if err := restored.Scan(s); err != nil {
		t.Fatalf("Scan round-trip error: %v", err)
	}
	if !reflect.DeepEqual(original, restored) {
		t.Errorf("round-trip failed: got %v, want %v", restored, original)
	}
}

func TestIntListValueEmpty(t *testing.T) {
	var il IntList
	v, err := il.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if v != "[]" {
		t.Errorf("Value() = %q, want %q", v, "[]")
	}
}

func TestParseBracketedIntListEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    IntList
		wantErr bool
	}{
		{"empty string", "", IntList{}, false},
		{"empty brackets", "[]", IntList{}, false},
		{"whitespace only brackets", "[  ]", IntList{}, false},
		{"trailing comma", "[1,2,]", IntList{1, 2}, false},
		{"no brackets", "1,2,3", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBracketedIntList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(IntList(got), tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
