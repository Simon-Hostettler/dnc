package editor

import (
	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/util"
)

/*
Type that value can delegate updates to.
Is responsible for invoking save callback.
*/
type ValueEditor interface {
	/*
		Takes a label, a value that should be edited and a callback that can be invoked to store the value.
	*/
	Init(util.KeyMap, string, interface{}, func(interface{}) error)

	Update(tea.Msg) tea.Cmd

	View() string

	Save() error

	Focus()

	Blur()
}

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

// returns a save callback that assigns v to target and then calls persist().
func BindString(target *string, persist func() error) func(string) error {
	return func(v string) error {
		*target = v
		return persist()
	}
}

// returns a save callback that assigns v to target and then calls persist().
func BindInt(target *int, persist func() error) func(int) error {
	return func(v int) error {
		*target = v
		return persist()
	}
}
