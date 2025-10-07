package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

type CharacterRow struct {
	keymap       util.KeyMap
	characterDir string
	name         string
}

func NewCharacterRow(name string, characterDir string, keymap util.KeyMap) *CharacterRow {
	return &CharacterRow{keymap, characterDir, name}
}

func (c *CharacterRow) Init() tea.Cmd {
	return nil
}

func (c *CharacterRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keymap.Select):
			return c, command.SelectCharacterCmd(c.name)
		case key.Matches(msg, c.keymap.Delete):
			return c, command.LaunchConfirmationDialogueCmd(
				func() tea.Cmd {
					return command.DeleteCharacterFileCmd(c.characterDir, c.name)
				},
			)
		}
	}
	return c, nil
}

func (c *CharacterRow) View() string {
	return c.name
}

func (c *CharacterRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}
