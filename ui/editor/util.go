package editor

import (
	"fmt"
	"reflect"
)

/*
Converts typed callback func to generic.
Will only invoke when parameter of correct type is passed, otherwise throws.
*/
func WrapTypedCallback[V any](fn func(V) error) func(interface{}) error {
	if fn == nil {
		return nil
	}
	typ := reflect.TypeOf((*V)(nil)).Elem()
	return func(i interface{}) error {
		if i == nil {
			return fmt.Errorf("expected %s, got nil", typ.String())
		}
		v, ok := i.(V)
		if !ok {
			return fmt.Errorf("expected %s, got %T", typ.String(), i)
		}
		return fn(v)
	}
}
