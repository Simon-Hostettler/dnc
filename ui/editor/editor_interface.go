package editor

import (
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/util"
)

/*
Type that a pointer to a value can delegate updates to.
Is responsible for saving new values to the pointer,
not saving the underlying struct to the filesystem
*/
type ValueEditor interface {
	/*
		Takes a label and pointer to a value that should be edited.
		Throws if the pointer is invalid / of the wrong type
	*/
	Init(util.KeyMap, string, interface{})

	Update(tea.Msg) tea.Cmd

	View() string // ValueEditor is not a tea.Model; returns plain string

	// Updates the delegator, does not store changes to filesystem.
	Save() tea.Cmd

	Focus()

	Blur()
}
