package editor

import (
	tea "charm.land/bubbletea/v2"
)

/*
Type that a pointer to a value can delegate updates to.
Is responsible for saving new values to the pointer,
not saving the underlying struct to the filesystem
*/
type ValueEditor interface {
	Update(tea.Msg) tea.Cmd

	View() string // ValueEditor is not a tea.Model; returns plain string

	// Updates the delegator, does not store changes to filesystem.
	Save() tea.Cmd

	Focus()

	Blur()
}
