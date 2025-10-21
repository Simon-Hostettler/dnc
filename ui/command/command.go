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

type DeleteCharacterRequestMsg struct {
	ID uuid.UUID
}

func DeleteCharacterRequest(id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		return DeleteCharacterRequestMsg{id}
	}
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

type CreateCharacterRequestMsg struct {
	Name string
}

func CreateCharacterRequest(name string) func() tea.Msg {
	return func() tea.Msg {
		return CreateCharacterRequestMsg{name}
	}
}

type CreateCharacterMsg struct {
	ID uuid.UUID
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

type WriteBackRequestMsg struct{}

func WriteBackRequest() tea.Msg {
	return WriteBackRequestMsg{}
}

type WriteBackMsg struct {
	Success bool
}

func WriteBackCmd(r repository.CharacterRepository, ctx context.Context, c *repository.CharacterAggregate) func() tea.Msg {
	return func() tea.Msg {
		err := r.Update(ctx, c)
		return WriteBackMsg{err == nil}
	}
}

type LoadSummariesRequestMsg struct{}

func LoadSummariesRequest() tea.Msg {
	return LoadSummariesRequestMsg{}
}

type LoadSummariesMsg struct {
	Summaries []models.CharacterSummary
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

type LoadCharacterMsg struct {
	Agg *repository.CharacterAggregate
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
