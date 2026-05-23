package models

import (
	"reflect"
	"testing"
)

func TestIntListScanFromInterfaceSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  IntList
	}{
		{"int32 elements", []any{int32(1), int32(2), int32(3)}, IntList{1, 2, 3}},
		{"empty slice", []any{}, IntList{}},
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
		{"unsupported source type", 42},
		{"unsupported source type string", "[1,2,3]"},
		{"unsupported element type int64", []any{int64(1)}},
		{"unsupported element type", []any{struct{}{}}},
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

func TestIntListValue(t *testing.T) {
	tests := []struct {
		name string
		in   IntList
		want string
	}{
		{"empty", IntList{}, "[]"},
		{"nil", nil, "[]"},
		{"populated", IntList{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, "[0,1,2,3,4,5,6,7,8,9]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.in.Value()
			if err != nil {
				t.Fatalf("Value() error: %v", err)
			}
			got, ok := v.(string)
			if !ok {
				t.Fatalf("Value() returned %T, want string", v)
			}
			if got != tt.want {
				t.Errorf("Value() = %q, want %q", got, tt.want)
			}
		})
	}
}
