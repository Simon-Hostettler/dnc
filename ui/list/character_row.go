package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

type CharacterRow struct {
	keymap         util.KeyMap
	character      *models.CharacterSummary
	deleteCallback tea.Cmd
}

func NewCharacterRow(character *models.CharacterSummary, deleteCallback tea.Cmd, keymap util.KeyMap) *CharacterRow {
	return &CharacterRow{keymap, character, deleteCallback}
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
					return c.deleteCallback
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
