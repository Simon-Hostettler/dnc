package command

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
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

type WriteBackRequestMsg struct{}

type LoadSummariesRequestMsg struct{}

type LoadSummariesMsg struct {
	Summaries []models.CharacterSummary
}

type LoadCharacterMsg struct {
	Agg *repository.CharacterAggregate
}

type CreateCharacterRequestMsg struct {
	Name string
}

type CreateCharacterMsg struct {
	ID uuid.UUID
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

type DeleteCharacterRequestMsg struct {
	ID uuid.UUID
}

type DeleteCharacterMsg struct {
	Success bool
}

func DeleteCharacterCmd(r repository.CharacterRepository, ctx context.Context, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		err := r.Delete(ctx, id)
		if err != nil {
			return DeleteCharacterMsg{false}
		}
		return DeleteCharacterMsg{true}
	}
}

func WriteBackRequest() tea.Msg {
	return WriteBackRequestMsg{}
}

func LoadSummariesRequest() tea.Msg {
	return LoadSummariesRequestMsg{}
}

func LoadSummariesCommand(r repository.CharacterRepository, ctx context.Context) func() tea.Msg {
	return func() tea.Msg {
		if sum, err := r.ListSummary(ctx); err != nil {
			return LoadSummariesMsg{[]models.CharacterSummary{}}
		} else {
			return LoadSummariesMsg{sum}
		}
	}
}

func DeleteCharacterRequest(id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		return DeleteCharacterRequestMsg{id}
	}
}

func CreateCharacterRequest(name string) func() tea.Msg {
	return func() tea.Msg {
		return CreateCharacterRequestMsg{name}
	}
}

func CreateCharacterCmd(r repository.CharacterRepository, ctx context.Context, name string) func() tea.Msg {
	return func() tea.Msg {
		if id, err := r.CreateEmpty(ctx, name); err != nil {
			return CreateCharacterMsg{}
		} else {
			return CreateCharacterMsg{id}
		}
	}
}

func WriteBackCmd(r repository.CharacterRepository, ctx context.Context, c *repository.CharacterAggregate) func() tea.Msg {
	return func() tea.Msg {
		err := r.Update(ctx, c)
		return DataOpMsg{DataSave, err == nil}
	}
}

func LoadCharacterCmd(r repository.CharacterRepository, ctx context.Context, id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		c, err := r.GetByID(ctx, id)
		if err != nil {
			return LoadCharacterMsg{nil}
		}
		return LoadCharacterMsg{c}
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
