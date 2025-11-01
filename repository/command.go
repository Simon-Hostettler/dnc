package repository

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

type LoadSummariesMsg struct {
	Summaries []models.CharacterSummary
}

func LoadSummariesCommand(r CharacterRepository, ctx context.Context) func() tea.Msg {
	return func() tea.Msg {
		if sum, err := r.ListSummary(ctx); err != nil {
			return LoadSummariesMsg{[]models.CharacterSummary{}}
		} else {
			return LoadSummariesMsg{sum}
		}
	}
}

type DeleteCharacterMsg struct {
	Success bool
}

func DeleteCharacterCmd(r CharacterRepository, ctx context.Context, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		err := r.Delete(ctx, id)
		if err != nil {
			return DeleteCharacterMsg{false}
		}
		return DeleteCharacterMsg{true}
	}
}

type CreateCharacterMsg struct {
	ID uuid.UUID
}

func CreateCharacterCmd(r CharacterRepository, ctx context.Context, name string) func() tea.Msg {
	return func() tea.Msg {
		if id, err := r.CreateEmpty(ctx, name); err != nil {
			return CreateCharacterMsg{}
		} else {
			return CreateCharacterMsg{id}
		}
	}
}

type WriteBackMsg struct {
	Success bool
}

func WriteBackCmd(r CharacterRepository, ctx context.Context, c *CharacterAggregate) func() tea.Msg {
	return func() tea.Msg {
		err := r.Update(ctx, c)
		return WriteBackMsg{err == nil}
	}
}

type LoadCharacterMsg struct {
	Agg *CharacterAggregate
}

func LoadCharacterCmd(r CharacterRepository, ctx context.Context, id uuid.UUID) func() tea.Msg {
	return func() tea.Msg {
		c, err := r.GetByID(ctx, id)
		if err != nil {
			return LoadCharacterMsg{nil}
		}
		return LoadCharacterMsg{c}
	}
}
