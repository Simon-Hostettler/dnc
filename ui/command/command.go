package command

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type ScreenIndex int

const (
	EditScreenIndex ScreenIndex = iota
	StatScreenIndex
	TitleScreenIndex
	SpellScreenIndex
	ConfirmationScreenIndex
	InventoryScreenIndex
	ReaderScreenIndex
)

type Direction int

const (
	UpDirection Direction = iota
	DownDirection
	LeftDirection
	RightDirection
)

type DeleteCharacterRequestMsg struct {
	ID uuid.UUID
}

func DeleteCharacterRequest(id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		return DeleteCharacterRequestMsg{id}
	}
}

type CreateCharacterRequestMsg struct {
	Name string
}

func CreateCharacterRequest(name string) func() tea.Msg {
	return func() tea.Msg {
		return CreateCharacterRequestMsg{name}
	}
}

type WriteBackRequestMsg struct{}

func WriteBackRequest() tea.Msg {
	return WriteBackRequestMsg{}
}

type LoadSummariesRequestMsg struct{}

func LoadSummariesRequest() tea.Msg {
	return LoadSummariesRequestMsg{}
}

type SelectCharacterMsg struct {
	ID uuid.UUID
}

func SelectCharacterCmd(id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		return SelectCharacterMsg{id}
	}
}

type SwitchScreenMsg struct {
	Screen ScreenIndex
}

func SwitchScreenCmd(s ScreenIndex) func() tea.Msg {
	return func() tea.Msg {
		return SwitchScreenMsg{s}
	}
}

type SwitchToPrevScreenMsg struct{}

func SwitchToPrevScreenCmd() tea.Msg {
	return SwitchToPrevScreenMsg{}
}

type FocusNextElementMsg struct {
	Direction Direction
}

/*
Use to switch focus to other element on same screen.
For switching to element in parent, use ReturnFocusToParentCmd
*/
func FocusNextElementCmd(d Direction) func() tea.Msg {
	return func() tea.Msg {
		return FocusNextElementMsg{d}
	}
}

type AppendElementMsg struct {
	Tag string
}

func AppendElementCmd(tag string) func() tea.Msg {
	return func() tea.Msg { return AppendElementMsg{tag} }
}

type ReturnFocusToParentMsg struct{}

func ReturnFocusToParentCmd() tea.Msg {
	return ReturnFocusToParentMsg{}
}

type LaunchConfirmationDialogueMsg struct {
	Callback func() tea.Cmd
}

func LaunchConfirmationDialogueCmd(callback func() tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return LaunchConfirmationDialogueMsg{callback}
	}
}

type LaunchReaderScreenMsg struct {
	Content string
}

func LaunchReaderScreenCmd(content string) func() tea.Msg {
	return func() tea.Msg {
		return LaunchReaderScreenMsg{content}
	}
}
