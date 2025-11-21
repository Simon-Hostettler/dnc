package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

type CharacterRow struct {
	id        uuid.UUID
	keymap    util.KeyMap
	character *models.CharacterSummary
}

func NewCharacterRow(keymap util.KeyMap, character *models.CharacterSummary) *CharacterRow {
	return &CharacterRow{uuid.New(), keymap, character}
}

func (c *CharacterRow) Id() uuid.UUID {
	return c.id
}

func (c *CharacterRow) Init() tea.Cmd {
	return nil
}

func (c *CharacterRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keymap.Select):
			return c, command.SelectCharacterCmd(c.character.ID)
		case key.Matches(msg, c.keymap.Delete):
			return c, command.LaunchConfirmationDialogueCmd(
				func() tea.Cmd {
					return command.DeleteCharacterRequest(c.character.ID)
				},
			)
		}
	}
	return c, nil
}

func (c *CharacterRow) View() string {
	return c.character.Name
}

func (c *CharacterRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}

func (c *CharacterRow) Selectable() bool {
	return true
}
