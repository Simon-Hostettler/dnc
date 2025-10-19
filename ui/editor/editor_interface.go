package editor

import (
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
