package command

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"hostettler.dev/dnc/repository"
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

type DataOperation int

const (
	DataDelete = iota
	DataUpdate
	DataCreate
	DataSave
)

type DataOpMsg struct {
	Op      DataOperation
	Success bool
}

type LoadCharacterMsg struct {
	Character repository.CharacterAggregate
}

type SelectCharacterMsg struct {
	ID uuid.UUID
}

type SwitchScreenMsg struct {
	Screen ScreenIndex
}

type FocusNextElementMsg struct {
	Direction Direction
}

type ReturnFocusToParentMsg struct{}

type AppendElementMsg struct {
	Tag string
}

type SwitchToPrevScreenMsg struct{}

type LaunchConfirmationDialogueMsg struct {
	Callback func() tea.Cmd
}

type LaunchReaderScreenMsg struct {
	Content string
}

func DataOperationCommand(callback func() error, op DataOperation) tea.Cmd {
	return func() tea.Msg {
		err := callback()
		if err == nil {
			return DataOpMsg{op, true}
		} else {
			return DataOpMsg{op, false}
		}
	}
}

func LoadCharacterCmd(r repository.CharacterRepository, ctx context.Context, id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		c, err := r.GetByID(ctx, id)
		if err != nil {
			panic("Character loaded incorrectly. Panicking to avoid corruption.")
		}
		return LoadCharacterMsg{*c}
	}
}

func SelectCharacterCmd(id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		return SelectCharacterMsg{id}
	}
}

func SwitchScreenCmd(s ScreenIndex) func() tea.Msg {
	return func() tea.Msg {
		return SwitchScreenMsg{s}
	}
}

func SwitchToPrevScreenCmd() tea.Msg {
	return SwitchToPrevScreenMsg{}
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

func AppendElementCmd(tag string) func() tea.Msg {
	return func() tea.Msg { return AppendElementMsg{tag} }
}

func ReturnFocusToParentCmd() tea.Msg {
	return ReturnFocusToParentMsg{}
}

func LaunchConfirmationDialogueCmd(callback func() tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return LaunchConfirmationDialogueMsg{callback}
	}
}

func LaunchReaderScreenCmd(content string) func() tea.Msg {
	return func() tea.Msg {
		return LaunchReaderScreenMsg{content}
	}
}
